package chplquerier

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

var chplAPICertProdListPath string = "/collections/certified_products"
var delimiter1 string = "☺"
var delimiter2 string = "☹"

var fields [11]string = [11]string{
	"id",
	"edition",
	"developer",
	"product",
	"version",
	"chplProductNumber",
	"certificationStatus",
	"criteriaMet",
	"apiDocumentation",
	"certificationDate",
	"practiceType"}

type chplCertifiedProduct struct {
	ID                  int    `json:"id"`
	ChplProductNumber   string `json:"chplProductNumber"`
	Edition             string `json:"edition"`
	PracticeType        string `json:"practiceType"`
	Developer           string `json:"developer"`
	Product             string `json:"product"`
	Version             string `json:"version"`
	CertificationDate   int64  `json:"certificationDate"`
	CertificationStatus string `json:"certificationStatus"`
	CriteriaMet         string `json:"criteriaMet"`
	APIDocumentation    string `json:"apiDocumentation"`
}

type chplCertifiedProductList struct {
	Results []chplCertifiedProduct `json:"results"`
}

func GetCHPLProducts(store endpointmanager.HealthITProductStore) error {
	fmt.Printf("requesting products\n")
	prodJSON, err := getProductJSON()
	if err != nil {
		return err
	}
	prodList, err := convertProductJSONToObj(prodJSON)
	if err != nil {
		return err
	}
	fmt.Printf("done requestion products\n")

	err = persistProducts(store, prodList)

	return err
}

func makeCHPLProductURL() (*url.URL, error) {
	queryArgs := make(map[string]string)
	fieldStr := strings.Join(fields[:], ",")
	queryArgs["fields"] = fieldStr

	chplURL, err := makeCHPLURL(chplAPICertProdListPath, queryArgs)
	if err != nil {
		return nil, err
	}

	return chplURL, nil
}

func getProductJSON() ([]byte, error) {
	chplURL, err := makeCHPLProductURL()
	if err != nil {
		return nil, err
	}

	// request ceritified products list
	resp, err := http.Get(chplURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp == nil {
		return nil, errors.New("CHPL certified products request had nil response")
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New("CHPL certified products request responded with status: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func convertProductJSONToObj(prodJSON []byte) (*chplCertifiedProductList, error) {
	var prodList chplCertifiedProductList

	fmt.Printf("unmarshalling json\n")

	err := json.Unmarshal(prodJSON, &prodList)
	if err != nil {
		return nil, err
	}
	fmt.Printf("finished unmarshalling json\n")

	return &prodList, nil
}

func parseHITProd(prod *chplCertifiedProduct) *endpointmanager.HealthITProduct {
	dbProd := endpointmanager.HealthITProduct{
		Name:                  prod.Product,
		Version:               prod.Version,
		Developer:             prod.Developer,
		CertificationStatus:   prod.CertificationStatus,
		CertificationDate:     time.Unix(prod.CertificationDate/1000, 0),
		CertificationEdition:  prod.Edition,
		CHPLID:                prod.ChplProductNumber,
		CertificationCriteria: strings.Split(prod.CriteriaMet, delimiter1),
		APIURL:                getAPIURL(prod.APIDocumentation),
	}

	return &dbProd
}

func persistProducts(store endpointmanager.HealthITProductStore, prodList *chplCertifiedProductList) error {
	for i, prod := range prodList.Results {

		if i%100 == 0 {
			fmt.Printf("Processing product #%d\n", i)
		}

		err := persistProduct(store, &prod)
		if err != nil {
			// TODO: figure out global logging
			fmt.Println(err)
			continue
		}
	}

	fmt.Printf("Done processing products\n")
	return nil
}

func persistProduct(store endpointmanager.HealthITProductStore, prod *chplCertifiedProduct) error {

	newDbProd := parseHITProd(prod)
	existingDbProd, err := store.GetHealthITProductUsingNameAndVersion(prod.Product, prod.Version)
	// need to add new entry
	if err == sql.ErrNoRows {
		err = store.AddHealthITProduct(newDbProd)
		if err != nil {
			return err
		}
	} else if !existingDbProd.Equal(newDbProd) {
		// changes exist. these may be due to products that have been certified multiple times w/in chpl.
		// we only care about the latest certification information for a particular version of software.

		needsUpdate, err := prodNeedsUpdate(existingDbProd, newDbProd)
		if err != nil {
			return err
		}

		if needsUpdate {
			existingDbProd.Update(newDbProd)
			err = store.UpdateHealthITProduct(existingDbProd)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// assumes that criteria/url chunks are delimited by '☺' and that criteria and url are separated by '☹'.
func getAPIURL(apiDocStr string) string {
	apiDocStrs := strings.Split(apiDocStr, delimiter1)
	if len(apiDocStrs) >= 1 {
		apiCritAndURL := strings.Split(apiDocStrs[0], delimiter2)
		if len(apiCritAndURL) == 2 {
			return apiCritAndURL[1]
		}
	}
	return ""
}

func prodNeedsUpdate(existingDbProd *endpointmanager.HealthITProduct, newDbProd *endpointmanager.HealthITProduct) (bool, error) {
	// check if the two are equal.
	if existingDbProd.Equal(newDbProd) {
		return false, nil
	}

	// begin by comparing certification editions.
	// Assumes certification editions are years, which is the case as of 11/20/19.
	existingCertEdition, err := strconv.Atoi(existingDbProd.CertificationEdition)
	if err != nil {
		return false, err
	}
	newCertEdition, err := strconv.Atoi(newDbProd.CertificationEdition)
	if err != nil {
		return false, err
	}

	// if new prod has more recent cert edition, should update.
	if newCertEdition > existingCertEdition {
		return true, nil
	} else if newCertEdition < existingCertEdition {
		return false, nil
	}

	// cert editions are the same. if new prod has more recent cert date, should update.
	if existingDbProd.CertificationDate.Before(newDbProd.CertificationDate) {
		return true, nil
	} else if existingDbProd.CertificationDate.After(newDbProd.CertificationDate) {
		return false, nil
	}

	// cert dates are the same. checking certification criteria lists. if new prod has more criteria, should update.
	if len(existingDbProd.CertificationCriteria) < len(newDbProd.CertificationCriteria) {
		return true, nil
	} else if len(existingDbProd.CertificationCriteria) > len(newDbProd.CertificationCriteria) {
		return false, nil
	}

	// certification criteria lists are the same lengths. unknown precedent for updates. return error.
	if !cmp.Equal(existingDbProd.CertificationCriteria, newDbProd.CertificationCriteria) {
		return false, errors.New("HealthITProducts certification edition and date are equal, but has same number but unequal certification criteria - unknown precendence for updates")
	}
	return false, errors.New("HealthITProducts certification edition, date, and criteria lists are equal - unknown precendence for updates")
}
