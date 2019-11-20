package chplquerier

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/spf13/viper"

	"errors"
	"fmt"
	"strings"
)

type CHPLQuerier interface {
}

type CHPLCertifiedProduct struct {
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

type CHPLCertifiedProductList struct {
	Results []CHPLCertifiedProduct `json:"results"`
}

var chplDomain string = "https://chpl.healthit.gov"
var chplAPIPath string = "/rest"
var chplAPICertProdList string = "/collections/certified_products"
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
var delimiter1 string = "☺"
var delimiter2 string = "☹"

func GetCHPLProducts(store endpointmanager.HealthITProductStore) error {
	err := getCertifiedProductCollection(store)

	return err
}

func makeCHPLURL() (*url.URL, error) {
	queryArgs := make(url.Values)
	chplURL, err := url.Parse(chplDomain)
	if err != nil {
		return nil, err
	}
	apiKey := viper.GetString("chplapikey")
	fieldStr := strings.Join(fields[:], ",")
	queryArgs.Set("api_key", apiKey)
	queryArgs.Set("fields", fieldStr)
	chplURL.RawQuery = queryArgs.Encode()
	chplURL.Path = chplAPIPath + chplAPICertProdList

	return chplURL, nil
}

func requestProducts() (*CHPLCertifiedProductList, error) {
	chplURL, err := makeCHPLURL()
	if err != nil {
		return nil, err
	}

	// TODO: use 'fields' in query
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

	var prodList CHPLCertifiedProductList

	fmt.Printf("unmarshalling json\n")

	err = json.Unmarshal(body, &prodList)
	if err != nil {
		return nil, err
	}
	fmt.Printf("finished unmarshalling json\n")

	return &prodList, nil
}

func parseHITProd(prod *CHPLCertifiedProduct) *endpointmanager.HealthITProduct {
	dbProd := new(endpointmanager.HealthITProduct)

	dbProd.Name = prod.Product
	dbProd.Version = prod.Version
	dbProd.Developer = prod.Developer
	dbProd.CertificationStatus = prod.CertificationStatus
	dbProd.CertificationDate = time.Unix(prod.CertificationDate/1000, 0)
	dbProd.CertificationEdition = prod.Edition
	dbProd.CHPLID = prod.ChplProductNumber
	dbProd.CertificationCriteria = strings.Split(prod.CriteriaMet, delimiter1)
	dbProd.APIURL = getAPIURL(prod.APIDocumentation)

	return dbProd
}

func persistProducts(store endpointmanager.HealthITProductStore, prodList *CHPLCertifiedProductList) error {
	fmt.Printf("%d\n", len(prodList.Results))
	prod1 := prodList.Results[7958]
	prod2 := prodList.Results[7959]

	fmt.Printf("prod 1 id: %d", prod1.ID)
	fmt.Printf("prod 2 id: %d", prod2.ID)

	for i, prod := range prodList.Results {

		if i%100 == 0 {
			fmt.Printf("Processing product #%d\n", i)
		}

		newDbProd := parseHITProd(&prod)
		existingDbProd, err := store.GetHealthITProductUsingNameAndVersion(prod.Product, prod.Version)
		// need to add new entry
		if err == sql.ErrNoRows {
			store.AddHealthITProduct(newDbProd)
		} else if !existingDbProd.Equal(newDbProd) {
			// changes exist. these may be due to products that have been certified multiple times w/in chpl.
			// we only care about the latest certification information for a particular version of software.

			// begin by comparing certification editions. Repalce with latest. Assumes certification editions
			// are years, which is the case as of 11/20/19.
			existingCertEdition, err := strconv.Atoi(existingDbProd.CertificationEdition)
			if err != nil {
				// TODO: figure out global logging
				fmt.Println(err)
				continue
			}
			newCertEdition, err := strconv.Atoi(newDbProd.CertificationEdition)
			if err != nil {
				// TODO: figure out global logging
				fmt.Println(err)
				continue
			}
			if newCertEdition > existingCertEdition {
				existingDbProd.Update(newDbProd)
			} else if newCertEdition == existingCertEdition {
				// compare certification dates
				if existingDbProd.CertificationDate.Before(newDbProd.CertificationDate) {
					existingDbProd.Update(newDbProd)
				} else if existingDbProd.CertificationDate.Equal(newDbProd.CertificationDate) {
					if len(existingDbProd.CertificationCriteria) < len(newDbProd.CertificationCriteria) {
						existingDbProd.Update(newDbProd)
					} else if len(existingDbProd.CertificationCriteria) == len(newDbProd.CertificationCriteria) {
						if !cmp.Equal(existingDbProd.CertificationCriteria, newDbProd.CertificationCriteria) {
							// TODO: figure out global logging
							fmt.Println("Criteria lists are not equal. Unable to determine precedence. Making no updates.")
						} else {
							// TODO: figure out global logging
							fmt.Println("Unknown changes. Unable to determine precedence. Making no updates.")
						}
					}
				}
			}
			store.UpdateHealthITProduct(existingDbProd)
		}
	}

	fmt.Printf("Done processing products\n")
	return nil
}

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

func getCertifiedProductCollection(store endpointmanager.HealthITProductStore) error {
	fmt.Printf("requesting products\n")
	prodList, err := requestProducts()
	if err != nil {
		return err
	}
	fmt.Printf("done requestion products\n")

	err = persistProducts(store, prodList)

	return err
}
