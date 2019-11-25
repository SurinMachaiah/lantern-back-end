package chplquerier

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/mock"

	logtest "github.com/sirupsen/logrus/hooks/test"

	"github.com/spf13/viper"
)

type mockStore struct {
	data []*endpointmanager.HealthITProduct
	mock.Store
}

var testCHPLProd chplCertifiedProduct = chplCertifiedProduct{
	ID:                  7849,
	ChplProductNumber:   "15.04.04.2657.Care.01.00.0.160701",
	Edition:             "2014",
	Developer:           "Carefluence",
	Product:             "Carefluence Open API",
	Version:             "1",
	CertificationDate:   1467331200000,
	CertificationStatus: "Active",
	CriteriaMet:         "170.315 (d)(1)☺170.315 (d)(10)☺170.315 (d)(9)☺170.315 (g)(4)☺170.315 (g)(5)☺170.315 (g)(6)☺170.315 (g)(7)☺170.315 (g)(8)☺170.315 (g)(9)",
	APIDocumentation:    "170.315 (g)(7)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
}

var testHITP endpointmanager.HealthITProduct = endpointmanager.HealthITProduct{
	Name:                  "Carefluence Open API",
	Version:               "1",
	Developer:             "Carefluence",
	CertificationStatus:   "Active",
	CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
	CertificationEdition:  "2014",
	CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
	CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"},
	APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
}

func Test_makeCHPLProductURLBasic(t *testing.T) {
	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	expected := "https://chpl.healthit.gov/rest/collections/certified_products?api_key=tmp_api_key&fields=id%2Cedition%2Cdeveloper%2Cproduct%2Cversion%2CchplProductNumber%2CcertificationStatus%2CcriteriaMet%2CapiDocumentation%2CcertificationDate%2CpracticeType"

	actualURL, err := makeCHPLProductURL()
	Assert(t, err == nil, err)

	actual := actualURL.String()
	Assert(t, expected == actual, fmt.Sprintf("Expected %s to equal %s.", actual, expected))
}

func Test_makeCHPLProductURLError(t *testing.T) {
	chplDomainOrig := chplDomain
	chplDomain = "http://%41:8080/" // invalid domain
	defer func() { chplDomain = chplDomainOrig }()

	_, err := makeCHPLProductURL()
	Assert(t, err != nil, "Expected error due to invalid domain name")
}

func Test_convertProductJSONToObj(t *testing.T) {
	var ctx context.Context
	var err error

	// test standard case
	prodListJSON := `{
		"results": [
		{
			"id": 7849,
			"chplProductNumber": "15.04.04.2657.Care.01.00.0.160701",
			"edition": "2014",
			"developer": "Carefluence",
			"product": "Carefluence Open API",
			"version": "1",
			"certificationDate": 1467331200000,
			"certificationStatus": "Active",
			"criteriaMet": "170.315 (d)(1)☺170.315 (d)(10)☺170.315 (d)(9)☺170.315 (g)(4)☺170.315 (g)(5)☺170.315 (g)(6)☺170.315 (g)(7)☺170.315 (g)(8)☺170.315 (g)(9)",
			"apiDocumentation": "170.315 (g)(7)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
		},
		{
			"id": 7850,
			"chplProductNumber": "15.04.04.2657.Care.01.00.0.160703",
			"edition": "2014",
			"developer": "Carefluence",
			"product": "Carefluence Open API",
			"version": "0.3",
			"certificationDate": 1467320000000,
			"certificationStatus": "Active",
			"criteriaMet": "170.315 (d)(1)☺170.315 (d)(10)☺170.315 (d)(9)☺170.315 (g)(4)☺170.315 (g)(5)☺170.315 (g)(6)☺170.315 (g)(7)☺170.315 (g)(8)☺170.315 (g)(9)",
			"apiDocumentation": "170.315 (g)(7)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
		}]}
		`

	expectedProd1 := testCHPLProd

	expectedProd2 := chplCertifiedProduct{
		ID:                  7850,
		ChplProductNumber:   "15.04.04.2657.Care.01.00.0.160703",
		Edition:             "2014",
		Developer:           "Carefluence",
		Product:             "Carefluence Open API",
		Version:             "0.3",
		CertificationDate:   1467320000000,
		CertificationStatus: "Active",
		CriteriaMet:         "170.315 (d)(1)☺170.315 (d)(10)☺170.315 (d)(9)☺170.315 (g)(4)☺170.315 (g)(5)☺170.315 (g)(6)☺170.315 (g)(7)☺170.315 (g)(8)☺170.315 (g)(9)",
		APIDocumentation:    "170.315 (g)(7)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}

	expectedProdList := chplCertifiedProductList{
		Results: []chplCertifiedProduct{expectedProd1, expectedProd2},
	}

	ctx = context.Background()
	prodList, err := convertProductJSONToObj(ctx, []byte(prodListJSON))
	Assert(t, err == nil, err)
	Assert(t, prodList.Results != nil, "Expected results field to be filled out for  product list.")
	Assert(t, len(prodList.Results) == len(expectedProdList.Results), fmt.Sprintf("Number of products is %d. Should be %d.", len(prodList.Results), len(expectedProdList.Results)))

	for i, prod := range prodList.Results {
		Assert(t, prod == expectedProdList.Results[i], "Expected parsed products to equal expected products.")
	}

	// test with done context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = convertProductJSONToObj(ctx, []byte(prodListJSON))
	Assert(t, errors.Cause(err) == context.Canceled, "Expected malformed JSON error")

	// test with malformed JSON
	ctx = context.Background()
	malformedJSON := `
		"asdf": [
		{}]}
		`

	_, err = convertProductJSONToObj(ctx, []byte(malformedJSON))
	switch errors.Cause(err).(type) {
	case *json.SyntaxError:
		// ok
	default:
		t.Fatal("Expected JSON syntax error")
	}
}

func Test_convertProductJSONToObjError(t *testing.T) {
}

func Test_parseHITProd(t *testing.T) {
	prod := testCHPLProd
	expectedHITProd := testHITP

	hitProd, err := parseHITProd(&prod)
	Assert(t, err == nil, err)
	Assert(t, hitProd.Equal(&expectedHITProd), "CHPL Product did not parse into HealthITProduct as expected.")

	// provide bad api doc string to cause error
	prod.APIDocumentation = "170.315 (g)(7)☹.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
	_, err = parseHITProd(&prod)
	Assert(t, err != nil, "Expected api doc parsing error")
}

func Test_getAPIURL(t *testing.T) {
	apiDocString := "170.315 (g)(7)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
	expectedURL := "http://carefluence.com/Carefluence-OpenAPI-Documentation.html"

	actualURL, err := getAPIURL(apiDocString)
	Assert(t, err == nil, err)
	Assert(t, expectedURL == actualURL, fmt.Sprintf("Expected '%s'. Got '%s'.", expectedURL, actualURL))

	// provide bad string - unexpected delimeter
	apiDocString = "170.315 (g)(7),http://carefluence.com/Carefluence-OpenAPI-Documentation.html"

	actualURL, err = getAPIURL(apiDocString)
	Assert(t, err != nil, "Expected error due to malformed api doc string")

	// provide empty string
	apiDocString = ""
	expectedURL = ""

	actualURL, err = getAPIURL(apiDocString)
	Assert(t, err == nil, err)
	Assert(t, expectedURL == actualURL, fmt.Sprintf("Expected an empty string"))

}

func Test_prodNeedsUpdate(t *testing.T) {

	type expectedResult struct {
		name        string
		hitProd     endpointmanager.HealthITProduct
		needsUpdate bool
		err         error
	}

	expectedResults := []expectedResult{}

	base := testHITP

	same := testHITP
	expectedResults = append(expectedResults, expectedResult{name: "same", hitProd: same, needsUpdate: false, err: nil})

	badEdition := testHITP
	badEdition.CertificationEdition = "asdf"
	expectedResults = append(expectedResults, expectedResult{name: "badEdition", hitProd: badEdition, needsUpdate: false, err: errors.New(`strconv.Atoi: parsing "asdf": invalid syntax`)})

	editionAfter := testHITP
	editionAfter.CertificationEdition = "2015"
	expectedResults = append(expectedResults, expectedResult{name: "editionAfter", hitProd: editionAfter, needsUpdate: true, err: nil})

	dateAfter := testHITP
	dateAfter.CertificationDate = time.Date(2016, 9, 1, 0, 0, 0, 0, time.UTC)
	expectedResults = append(expectedResults, expectedResult{name: "dateAfter", hitProd: dateAfter, needsUpdate: true, err: nil})

	critListLonger := testHITP
	critListLonger.CertificationCriteria = []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)", "170.315 (g)(10)"}
	expectedResults = append(expectedResults, expectedResult{name: "critListLonger", hitProd: critListLonger, needsUpdate: true, err: nil})

	editionBefore := testHITP
	editionBefore.CertificationEdition = "2011"
	expectedResults = append(expectedResults, expectedResult{name: "editionBefore", hitProd: editionBefore, needsUpdate: false, err: nil})

	dateBefore := testHITP
	dateBefore.CertificationDate = time.Date(2016, 5, 1, 0, 0, 0, 0, time.UTC)
	expectedResults = append(expectedResults, expectedResult{name: "dateBefore", hitProd: dateBefore, needsUpdate: false, err: nil})

	critListShorter := testHITP
	critListShorter.CertificationCriteria = []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)"}
	expectedResults = append(expectedResults, expectedResult{name: "critListShorter", hitProd: critListShorter, needsUpdate: false, err: nil})

	critListDiff := testHITP
	critListDiff.CertificationCriteria = []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(10)"}
	expectedResults = append(expectedResults, expectedResult{name: "critListDiff", hitProd: critListDiff, needsUpdate: false, err: errors.New("HealthITProducts certification edition and date are equal, but has same number but unequal certification criteria - unknown precendence for updates")})

	chplID := testHITP
	chplID.CHPLID = "15.04.04.2657.Care.01.00.0.160733"
	expectedResults = append(expectedResults, expectedResult{name: "chplID", hitProd: chplID, needsUpdate: false, err: nil})

	certStatus := testHITP
	certStatus.CertificationStatus = "Retired"
	expectedResults = append(expectedResults, expectedResult{name: "certStatus", hitProd: certStatus, needsUpdate: false, err: errors.New("HealthITProducts certification edition, date, and criteria lists are equal - unknown precendence for updates")})

	for _, expRes := range expectedResults {
		needsUpdate, err := prodNeedsUpdate(&base, &(expRes.hitProd))
		Assert(t, needsUpdate == expRes.needsUpdate, fmt.Sprintf("For 'prodNeedsUpdate' using %s, expected %t and got %t.", expRes.name, expRes.needsUpdate, needsUpdate))
		if err != nil && expRes.err == nil {
			t.Fatalf("For 'prodNeedsUpdate' using %s, did not expect error but got error\n%v", expRes.name, err)
		}
		if err == nil && expRes.err != nil {
			t.Fatalf("For 'prodNeedsUpdate' using %s, did not receive error but expected error\n%v", expRes.name, expRes.err)
		}
		if err != nil && expRes.err != nil {
			origErr := errors.Cause(err)
			if origErr.Error() != expRes.err.Error() {
				t.Fatalf("For 'prodNeedsUpdate' using %s, expected error\n%v\nAnd got error\n%v", expRes.name, expRes.err, origErr)
			}
		}
	}

	baseBadEdition := testHITP
	baseBadEdition.CertificationEdition = "asdf"
	name := "baseBadEdition"
	expectedNeedsUpdate := false
	expectedErrorStr := `strconv.Atoi: parsing "asdf": invalid syntax`

	needsUpdate, err := prodNeedsUpdate(&baseBadEdition, &base)
	Assert(t, needsUpdate == expectedNeedsUpdate, fmt.Sprintf("For 'prodNeedsUpdate' using %s, expected %t and got %t.", name, expectedNeedsUpdate, needsUpdate))
	Assert(t, err != nil, "Expected an error")
	origErr := errors.Cause(err)
	Assert(t, origErr.Error() == expectedErrorStr, fmt.Sprintf("For 'prodNeedsUpdate' using %s, expected error\n%v\nAnd got error\n%v", name, expectedErrorStr, origErr))
}

func Test_persistProduct(t *testing.T) {
	store, err := createStore()
	if err != nil {
		t.Fatalf("create mock store error\n%v", err)
	}
	storeWContext := endpointmanager.HealthITProductStoreWithContext{store}

	var ctx context.Context
	var cancel context.CancelFunc

	prod := testCHPLProd
	hitp := testHITP

	// check that ended context when no element in store fails as expected
	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	err = persistProduct(ctx, storeWContext, &prod)
	Assert(t, len(store.(*mockStore).data) == 0, "should not have stored data")
	Assert(t, errors.Cause(err) == context.Canceled, "should have errored out with root cause that the context was canceled")

	// reset context
	ctx = context.Background()

	// check that new item is stored
	err = persistProduct(ctx, storeWContext, &prod)
	Assert(t, err == nil, err)
	Assert(t, len(store.(*mockStore).data) == 1, "did not store data as expected")
	Assert(t, hitp.Equal(store.(*mockStore).data[0]), "stored data does not equal expected store data")

	// check that newer updated item replaces item
	prod.Edition = "2015"
	hitp.CertificationEdition = "2015"
	err = persistProduct(ctx, storeWContext, &prod)
	Assert(t, err == nil, err)
	Assert(t, len(store.(*mockStore).data) == 1, "did not store data as expected")
	Assert(t, hitp.Equal(store.(*mockStore).data[0]), "stored data does not equal expected store data")

	// check that older updated item does not replace item
	prod.Edition = "2014"
	hitp.CertificationEdition = "2015" // keeping 2015
	err = persistProduct(ctx, storeWContext, &prod)
	Assert(t, err == nil, err)
	Assert(t, len(store.(*mockStore).data) == 1, "did not store data as expected")
	Assert(t, hitp.Equal(store.(*mockStore).data[0]), "stored data does not equal expected store data")

	// check that malformed product throws error
	prod.APIDocumentation = "170.315 (g)(7),http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
	err = persistProduct(ctx, storeWContext, &prod)
	Assert(t, err != nil, "expected error parsing product")

	// check that ambiguous update throws error
	prod = testCHPLProd
	prod.Edition = "2015" // same date as what is in store
	prod.CertificationStatus = "Retired"
	err = persistProduct(ctx, storeWContext, &prod)
	Assert(t, err != nil, "expected error updating product")
}

func Test_persistProducts(t *testing.T) {
	store, err := createStore()
	if err != nil {
		t.Fatalf("create mock store error\n%v", err)
	}
	storeWContext := endpointmanager.HealthITProductStoreWithContext{store}

	ctx := context.Background()

	// standard persist

	prod1 := testCHPLProd
	prod2 := testCHPLProd
	prod2.Product = "another prod"

	prodList := chplCertifiedProductList{Results: []chplCertifiedProduct{prod1, prod2}}

	err = persistProducts(ctx, storeWContext, &prodList)
	Assert(t, err == nil, err)

	Assert(t, len(store.(*mockStore).data) == 2, "did not persist two products as expected")
	Assert(t, store.(*mockStore).data[0].Name == testCHPLProd.Product, "Did not store first product as expected")
	Assert(t, store.(*mockStore).data[1].Name == "another prod", "Did not store second product as expected")

	// persist with errors

	hook := logtest.NewGlobal()
	store, err = createStore()
	if err != nil {
		t.Fatalf("create mock store error\n%v", err)
	}
	storeWContext = endpointmanager.HealthITProductStoreWithContext{store}

	prod2.APIDocumentation = "170.315 (g)(7),http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
	expectedErr := "retreiving the API URL from the health IT product API documentation list failed: unexpected format for api doc string"
	prodList = chplCertifiedProductList{Results: []chplCertifiedProduct{prod1, prod2}}

	err = persistProducts(ctx, storeWContext, &prodList)
	// don't expect the function to return with errors
	Assert(t, err == nil, err)
	// only expect one item to be stored
	Assert(t, len(store.(*mockStore).data) == 1, "did not persist one product as expected")
	Assert(t, store.(*mockStore).data[0].Name == testCHPLProd.Product, "Did not store first product as expected")
	// expect presence of a log message
	found := false
	for i := range hook.Entries {
		if hook.Entries[i].Message == expectedErr {
			found = true
			break
		}
	}
	Assert(t, found, "expected an error to be logged")

	// persist when context has ended
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	store, err = createStore()
	if err != nil {
		t.Fatalf("create mock store error\n%v", err)
	}
	storeWContext = endpointmanager.HealthITProductStoreWithContext{store}

	prod2 = testCHPLProd
	prod2.Product = "another prod"

	err = persistProducts(ctx, storeWContext, &prodList)
	Assert(t, errors.Cause(err) == context.Canceled, "expected persistProducts to error out due to context ending")
}

func Assert(t *testing.T, boolStatement bool, errorValue interface{}) {
	if !boolStatement {
		t.Fatalf("%s: %v", t.Name(), errorValue)
	}
}

func createStore() (endpointmanager.HealthITProductStore, error) {

	store := mockStore{}

	store.AddHealthITProductFn = func(hitp *endpointmanager.HealthITProduct) error {
		for _, existingHitp := range store.data {
			if existingHitp.ID == hitp.ID {
				return errors.New("HealthITProduct with that ID already exists")
			}
		}
		// want to store a copy
		newHitp := *hitp
		newHitp.ID = len(store.data) + 1
		store.data = append(store.data, &newHitp)
		return nil
	}

	store.GetHealthITProductUsingNameAndVersionFn = func(name string, version string) (*endpointmanager.HealthITProduct, error) {
		for _, existingHitp := range store.data {
			if existingHitp.Name == name && existingHitp.Version == version {
				// want to return a copy
				hitp := *existingHitp
				return &hitp, nil
			}
		}
		return nil, sql.ErrNoRows
	}

	store.UpdateHealthITProductFn = func(hitp *endpointmanager.HealthITProduct) error {
		var existingHitp *endpointmanager.HealthITProduct
		var i int
		replace := false
		for i, existingHitp = range store.data {
			if existingHitp.ID == hitp.ID {
				replace = true
				break
			}
		}
		if replace {
			// replacing with copy
			updatedHitp := *hitp
			store.data[i] = &updatedHitp
		} else {
			return errors.New("No existing entry exists")
		}

		return nil
	}

	return &store, nil
}
