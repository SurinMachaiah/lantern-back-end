// +build integration

package postgresql

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_PersistFHIREndpointInfo(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	// capability statement
	path := filepath.Join("../../testdata", "cerner_capability_dstu2.json")
	csJSON, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
	}
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	if err != nil {
		t.Error(err)
	}

	// add an endpoint that can later be referenced
	var endpoint1 = &endpointmanager.FHIREndpoint{
		URL:              "example.com/FHIR/DSTU2/",
		OrganizationName: "Example Inc.",
		ListSource:       "https://github.com/cerner/ignite-endpoints"}
	var endpoint2 = &endpointmanager.FHIREndpoint{
		URL:              "other.example.com/FHIR/DSTU2/",
		OrganizationName: "Other Example Inc."}
	store.AddFHIREndpoint(ctx, endpoint1)
	store.AddFHIREndpoint(ctx, endpoint2)

	// endpointInfos
	var endpointInfo1 = &endpointmanager.FHIREndpointInfo{
		FHIREndpointID:      endpoint1.ID,
		TLSVersion:          "TLS 1.1",
		MIMETypes:           []string{"application/json+fhir"},
		HTTPResponse:        200,
		Errors:              "Example Error",
		Vendor:              "Cerner",
		CapabilityStatement: cs}
	var endpointInfo2 = &endpointmanager.FHIREndpointInfo{
		FHIREndpointID: endpoint2.ID,
		TLSVersion:     "TLS 1.2",
		MIMETypes:      []string{"application/fhir+json"},
		HTTPResponse:   404,
		Errors:         "Example Error 2"}

	// add endpointInfos

	err = store.AddFHIREndpointInfo(ctx, endpointInfo1)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %s", err.Error())
	}

	err = store.AddFHIREndpointInfo(ctx, endpointInfo2)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %+v", err)
	}

	// retrieve endpointInfos

	e1, err := store.GetFHIREndpointInfoUsingFHIREndpointID(ctx, endpoint1.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if !e1.Equal(endpointInfo1) {
		t.Errorf("retrieved endpointInfo is not equal to saved endpointInfo.")
	}

	e2, err := store.GetFHIREndpointInfoUsingFHIREndpointID(ctx, endpoint2.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if !e2.Equal(endpointInfo2) {
		t.Errorf("retrieved endpointInfo is not equal to saved endpointInfo.")
	}

	eID1, err := store.GetFHIREndpointInfo(ctx, e1.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if !eID1.Equal(endpointInfo1) {
		t.Errorf("retrieved endpointInfo is not equal to saved endpointInfo.")
	}

	eID2, err := store.GetFHIREndpointInfo(ctx, e2.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if !eID2.Equal(endpointInfo2) {
		t.Errorf("retrieved endpointInfo is not equal to saved endpointInfo.")
	}

	// update endpointInfo

	e1.HTTPResponse = 700

	err = store.UpdateFHIREndpointInfo(ctx, e1)
	if err != nil {
		t.Errorf("Error updating fhir endpointInfo: %s", err.Error())
	}

	e1, err = store.GetFHIREndpointInfo(ctx, endpointInfo1.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if e1.Equal(endpointInfo1) {
		t.Errorf("retrieved UPDATED endpointInfo is equal to original endpointInfo.")
	}
	if e1.UpdatedAt.Equal(e1.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	e1.HTTPResponse = 200

	// update with nil capability statement
	capStat := e1.CapabilityStatement
	e1.CapabilityStatement = nil

	err = store.UpdateFHIREndpointInfo(ctx, e1)
	if err != nil {
		t.Errorf("Error updating fhir endpointInfo: %s", err.Error())
	}

	e1, err = store.GetFHIREndpointInfo(ctx, endpointInfo1.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if e1.CapabilityStatement != nil {
		t.Errorf("Expected capability statement to be nil")
	}

	e1.CapabilityStatement = capStat

	// check cascade delete

	err = store.DeleteFHIREndpoint(ctx, endpoint1) // delete the endpoint
	if err != nil {
		t.Errorf("Error deleting fhir endpointInfo: %s", err.Error())
	}

	_, err = store.GetFHIREndpointInfo(ctx, endpointInfo1.ID) // ensure we deleted the corresponding endpoint info enetry
	if err == nil {
		t.Errorf("endpointInfo1 was not deleted: %s", err.Error())
	}

	// delete endpointInfos

	_, err = store.GetFHIREndpointInfo(ctx, endpointInfo2.ID) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("error retrieving endpointInfo2 after deleting endpointInfo1: %s", err.Error())
	}

	err = store.DeleteFHIREndpointInfo(ctx, endpointInfo2)
	if err != nil {
		t.Errorf("Error deleting fhir endpointInfo: %s", err.Error())
	}

	_, err = store.GetFHIREndpointInfo(ctx, endpointInfo2.ID) // ensure we deleted the entry
	if err == nil {
		t.Errorf("expected error retrieving endpoint2 after deleting endpoint2")
	}

	// check history table
	var count int
	var response int
	var capStatJson []byte

	// check insertions
	rows := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history WHERE id=$1 AND operation='I';", endpointInfo1.ID)
	err = rows.Scan(&count)
	if err != nil {
		t.Errorf("history count for insertions: %s", err.Error())
	}
	if count != 1 {
		t.Errorf("expected 1 insertion for endpointInfo1. Got %d.", count)
	}

	// check the value
	rows = store.DB.QueryRow("SELECT http_response, capability_statement FROM fhir_endpoints_info_history WHERE id=$1 AND operation='I';", endpointInfo1.ID)
	err = rows.Scan(&response, &capStatJson)
	if err != nil {
		t.Errorf("get values for insertion: %s", err.Error())
	}
	if response != 200 {
		t.Errorf("expected http_response to be 200 for endpointInfo1 insert. Got %d.", response)
	}
	if bytes.Equal(capStatJson, []byte("null")) {
		t.Errorf("expected capability_statement to be present for endpointInfo1 insert. Got nil.")
	}

	// check updates

	// check that there are two
	rows = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history WHERE id=$1 AND operation='U';", endpointInfo1.ID)
	err = rows.Scan(&count)
	if err != nil {
		t.Errorf("history count for insertions: %s", err.Error())
	}
	if count != 2 {
		t.Errorf("expected 2 updates for endpointInfo1. Got %d.", count)
	}

	// get the first update and check its value
	rows = store.DB.QueryRow("SELECT http_response, capability_statement FROM fhir_endpoints_info_history WHERE id=$1 AND operation='U' ORDER BY entered_at ASC LIMIT 1;", endpointInfo1.ID)
	err = rows.Scan(&response, &capStatJson)
	if err != nil {
		t.Errorf("history count for insertions: %s", err.Error())
	}
	if response != 700 {
		t.Errorf("expected http_response to be 700 for update value for endpointInfo1. Got %d.", response)
	}
	if bytes.Equal(capStatJson, []byte("null")) {
		t.Errorf("expected capability_statement to be present for endpointInfo1 insert. Got nil.")
	}

	// get the second update and check its value
	rows = store.DB.QueryRow("SELECT http_response, capability_statement FROM fhir_endpoints_info_history WHERE id=$1 AND operation='U' ORDER BY entered_at DESC LIMIT 1;", endpointInfo1.ID)
	err = rows.Scan(&response, &capStatJson)
	if err != nil {
		t.Errorf("history count for insertions: %s", err.Error())
	}
	if response != 200 {
		t.Errorf("expected http_response to be 200 for update value for endpointInfo1. Got %d.", response)
	}
	if !bytes.Equal(capStatJson, []byte("null")) {
		t.Errorf("did not expect the capability statement to be present.")
	}

	// check deletes
	rows = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history WHERE id=$1 AND operation='D';", endpointInfo1.ID)
	err = rows.Scan(&count)
	if err != nil {
		t.Errorf("history count for deletions: %s", err.Error())
	}
	if count != 1 {
		t.Errorf("expected 1 deletion for endpointInfo1. Got %d.", count)
	}
}
