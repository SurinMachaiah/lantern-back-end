package chplquerier

import (
	"testing"

	"github.com/spf13/viper"
)

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
	prodJSON := `
		{
			"id": 7849,
			"chplProductNumber": "15.04.04.2657.Care.01.00.0.160701",
			"edition": "2015",
			"developer": "Carefluence",
			"product": "Carefluence Open API",
			"version": "1",
			"certificationDate": 1467331200000,
			"certificationStatus": "Active",
			"criteriaMet": "170.315 (d)(1)☺170.315 (d)(10)☺170.315 (d)(9)☺170.315 (g)(4)☺170.315 (g)(5)☺170.315 (g)(6)☺170.315 (g)(7)☺170.315 (g)(8)☺170.315 (g)(9)",
			"apiDocumentation": "170.315 (g)(7)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
		}
		`

	expectedProd := CHPLCertifiedProduct{
		ID:                  7849,
		ChplProductNumber:   "15.04.04.2657.Care.01.00.0.160701",
		Edition:             "2015",
		Developer:           "Carefluence",
		Product:             "Carefluence Open API",
		Version:             "1",
		CertificationDate:   1467331200000,
		CertificationStatus: "Active",
		CriteriaMet:         "170.315 (d)(1)☺170.315 (d)(10)☺170.315 (d)(9)☺170.315 (g)(4)☺170.315 (g)(5)☺170.315 (g)(6)☺170.315 (g)(7)☺170.315 (g)(8)☺170.315 (g)(9)",
		APIDocumentation:    "170.315 (g)(7)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	}

	prod, err := convertProductJSONToObj([]byte(prodJSON))
	if err != nil {
		t.Fatal(err)
	}

	if prod != expectedProd {
		t.Fatalf("Expected and actual products are not equal.")
	}
}
