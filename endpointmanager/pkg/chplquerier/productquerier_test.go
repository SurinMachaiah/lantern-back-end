package chplquerier

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/mock"

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
	if err != nil {
		t.Fatal(err)
	}

	actual := actualURL.String()

	if expected != actual {
		t.Fatalf("Expected %s to equal %s.", actual, expected)
	}
}

func Test_makeCHPLProductURLError(t *testing.T) {
	chplDomainOrig := chplDomain
	chplDomain = "http://%41:8080/" // invalid domain
	defer func() { chplDomain = chplDomainOrig }()

	_, err := makeCHPLProductURL()
	if err == nil {
		t.Fatal("Expected error due to invalid domain name")
	}
}

func Test_convertProductJSONToObj(t *testing.T) {
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

	prodList, err := convertProductJSONToObj([]byte(prodListJSON))
	if err != nil {
		t.Fatal(err)
	}

	if prodList.Results == nil {
		t.Fatalf("Expected results field to be filled out for  product list.")
	}
	if len(prodList.Results) != len(expectedProdList.Results) {
		t.Fatalf("Number of products is %d. Should be %d.", len(prodList.Results), len(expectedProdList.Results))
	}
	for i, prod := range prodList.Results {
		if prod != expectedProdList.Results[i] {
			t.Fatalf("Expected parsed products to equal expected products.")
		}
	}
}

func Test_convertProductJSONToObjError(t *testing.T) {
	malformedJSON := `
		"asdf": [
		{}]}
		`

	_, err := convertProductJSONToObj([]byte(malformedJSON))
	if err == nil {
		t.Fatalf("Expected malformed JSON error")
	}
}

func Test_parseHITProd(t *testing.T) {
	prod := testCHPLProd

	expectedHITProd := testHITP

	hitProd := parseHITProd(&prod)

	if !hitProd.Equal(&expectedHITProd) {
		t.Fatalf("CHPL Product did not parse into HealthITProduct as expected.")
	}
}

func Test_getAPIURL(t *testing.T) {
	apiDocString := "170.315 (g)(7)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
	expectedURL := "http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
	actualURL := getAPIURL(apiDocString)

	if expectedURL != actualURL {
		t.Fatalf("Expected '%s'. Got '%s'.", expectedURL, actualURL)
	}
}

func Test_prodNeedsUpdate(t *testing.T) {

	type expectedResult struct {
		name        string
		hitProd     endpointmanager.HealthITProduct
		needsUpdate bool
		err         error
	}

	base := testHITP

	same := testHITP

	badEdition := testHITP
	badEdition.CertificationEdition = "asdf"

	// base := endpointmanager.HealthITProduct{
	// 	Name:                  "Carefluence Open API",
	// 	Version:               "1",
	// 	Developer:             "Carefluence",
	// 	CertificationStatus:   "Active",
	// 	CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
	// 	CertificationEdition:  "2014",
	// 	CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
	// 	CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"},
	// 	APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	// }

	// same := endpointmanager.HealthITProduct{
	// 	Name:                  "Carefluence Open API",
	// 	Version:               "1",
	// 	Developer:             "Carefluence",
	// 	CertificationStatus:   "Active",
	// 	CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
	// 	CertificationEdition:  "2014",
	// 	CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
	// 	CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"},
	// 	APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	// }

	// badEdition := endpointmanager.HealthITProduct{
	// 	Name:                  "Carefluence Open API",
	// 	Version:               "1",
	// 	Developer:             "Carefluence",
	// 	CertificationStatus:   "Active",
	// 	CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
	// 	CertificationEdition:  "asdf",
	// 	CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
	// 	CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"},
	// 	APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	// }

	editionAfter := endpointmanager.HealthITProduct{
		Name:                  "Carefluence Open API",
		Version:               "1",
		Developer:             "Carefluence",
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2015",
		CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
		CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"},
		APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}

	dateAfter := endpointmanager.HealthITProduct{
		Name:                  "Carefluence Open API",
		Version:               "1",
		Developer:             "Carefluence",
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2016, 9, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2014",
		CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
		CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"},
		APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}

	critListLonger := endpointmanager.HealthITProduct{
		Name:                  "Carefluence Open API",
		Version:               "1",
		Developer:             "Carefluence",
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2014",
		CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
		CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)", "170.315 (g)(10)"},
		APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}

	editionBefore := endpointmanager.HealthITProduct{
		Name:                  "Carefluence Open API",
		Version:               "1",
		Developer:             "Carefluence",
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2011",
		CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
		CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"},
		APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}

	dateBefore := endpointmanager.HealthITProduct{
		Name:                  "Carefluence Open API",
		Version:               "1",
		Developer:             "Carefluence",
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2016, 5, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2014",
		CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
		CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"},
		APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}

	critListShorter := endpointmanager.HealthITProduct{
		Name:                  "Carefluence Open API",
		Version:               "1",
		Developer:             "Carefluence",
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2014",
		CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
		CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)"},
		APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}

	critListDiff := endpointmanager.HealthITProduct{
		Name:                  "Carefluence Open API",
		Version:               "1",
		Developer:             "Carefluence",
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2014",
		CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
		CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(10)"},
		APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}

	chplID := endpointmanager.HealthITProduct{
		Name:                  "Carefluence Open API",
		Version:               "1",
		Developer:             "Carefluence",
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2014",
		CHPLID:                "15.04.04.2657.Care.01.00.0.160733",
		CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"},
		APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}

	certStatus := endpointmanager.HealthITProduct{
		Name:                  "Carefluence Open API",
		Version:               "1",
		Developer:             "Carefluence",
		CertificationStatus:   "Retired",
		CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2014",
		CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
		CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"},
		APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}

	expectedResults := []expectedResult{
		expectedResult{name: "same", hitProd: same, needsUpdate: false, err: nil},
		expectedResult{name: "badEdition", hitProd: badEdition, needsUpdate: false, err: errors.New(`strconv.Atoi: parsing "asdf": invalid syntax`)},
		expectedResult{name: "editionAfter", hitProd: editionAfter, needsUpdate: true, err: nil},
		expectedResult{name: "dateAfter", hitProd: dateAfter, needsUpdate: true, err: nil},
		expectedResult{name: "critListLonger", hitProd: critListLonger, needsUpdate: true, err: nil},
		expectedResult{name: "editionBefore", hitProd: editionBefore, needsUpdate: false, err: nil},
		expectedResult{name: "dateBefore", hitProd: dateBefore, needsUpdate: false, err: nil},
		expectedResult{name: "critListShorter", hitProd: critListShorter, needsUpdate: false, err: nil},
		expectedResult{name: "critListDiff", hitProd: critListDiff, needsUpdate: false, err: errors.New("HealthITProducts certification edition and date are equal, but has same number but unequal certification criteria - unknown precendence for updates")},
		expectedResult{name: "chplID", hitProd: chplID, needsUpdate: false, err: nil},
		expectedResult{name: "certStatus", hitProd: certStatus, needsUpdate: false, err: errors.New("HealthITProducts certification edition, date, and criteria lists are equal - unknown precendence for updates")},
	}

	for _, expRes := range expectedResults {
		needsUpdate, err := prodNeedsUpdate(&base, &(expRes.hitProd))
		if needsUpdate != expRes.needsUpdate {
			t.Fatalf("For 'prodNeedsUpdate' using %s, expected %t and got %t.", expRes.name, expRes.needsUpdate, needsUpdate)
		}
		if err != nil && expRes.err == nil {
			t.Fatalf("For 'prodNeedsUpdate' using %s, did not expect error but got error\n%v", expRes.name, err)
		}
		if err == nil && expRes.err != nil {
			t.Fatalf("For 'prodNeedsUpdate' using %s, did not receive error but expected error\n%v", expRes.name, expRes.err)
		}
		if err != nil && expRes.err != nil && err.Error() != expRes.err.Error() {
			t.Fatalf("For 'prodNeedsUpdate' using %s, expected error\n%v\nAnd got error\n%v", expRes.name, expRes.err, err)
		}
	}

	baseBadEdition := endpointmanager.HealthITProduct{
		Name:                  "Carefluence Open API",
		Version:               "1",
		Developer:             "Carefluence",
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "asdf",
		CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
		CertificationCriteria: []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"},
		APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}
	name := "baseBadEdition"
	expectedNeedsUpdate := false
	expectedErrorStr := `strconv.Atoi: parsing "asdf": invalid syntax`

	needsUpdate, err := prodNeedsUpdate(&baseBadEdition, &base)
	if needsUpdate != expectedNeedsUpdate {
		t.Fatalf("For 'prodNeedsUpdate' using %s, expected %t and got %t.", name, expectedNeedsUpdate, needsUpdate)
	}
	if err == nil || err.Error() != expectedErrorStr {
		t.Fatalf("For 'prodNeedsUpdate' using %s, expected error\n%v\nAnd got error\n%v", name, expectedErrorStr, err)
	}
}

func Test_persistProduct(t *testing.T) {
	store, err := createStore()
	if err != nil {
		t.Fatalf("create mock store error\n%v", err)
	}

	prod := testCHPLProd
	hitp := testHITP

	// check that new item is stored
	err = persistProduct(store, &prod)
	AssertTrue(t, err == nil, err)
	AssertTrue(t, len(store.(*mockStore).data) == 1, "did not store data as expected")
	AssertTrue(t, hitp.Equal(store.(*mockStore).data[0]), "stored data does not equal expected store data")

	// check that newer updated item replaces item
	prod.Edition = "2015"
	hitp.CertificationEdition = "2015"
	err = persistProduct(store, &prod)
	AssertTrue(t, err == nil, err)
	AssertTrue(t, len(store.(*mockStore).data) == 1, "did not store data as expected")
	AssertTrue(t, hitp.Equal(store.(*mockStore).data[0]), "stored data does not equal expected store data")

	// check that older updated item does not replace item
	prod.Edition = "2014"
	hitp.CertificationEdition = "2015" // keeping 2015
	err = persistProduct(store, &prod)
	AssertTrue(t, err == nil, err)
	AssertTrue(t, len(store.(*mockStore).data) == 1, "did not store data as expected")
	AssertTrue(t, hitp.Equal(store.(*mockStore).data[0]), "stored data does not equal expected store data")

}

func AssertTrue(t *testing.T, boolStatement bool, errorValue interface{}) {
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
