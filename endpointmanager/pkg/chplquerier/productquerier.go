package chplquerier

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/httpclient"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

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

func GetCHPLProducts(ctx context.Context, store endpointmanager.HealthITProductStore, cli *httpclient.Client) error {
	storeWContext := endpointmanager.HealthITProductStoreWithContext{store}

	fmt.Printf("requesting products\n")

	prodJSON, err := getProductJSON(ctx, cli)
	if err != nil {
		return errors.Wrap(err, "getting health IT product JSON failed")
	}
	fmt.Printf("done requesting products\n")

	fmt.Printf("converting products")
	prodList, err := convertProductJSONToObj(ctx, prodJSON)
	if err != nil {
		return errors.Wrap(err, "converting health IT product JSON into a 'chplCertifiedProductList' object failed")
	}
	fmt.Printf("done converting products")

	fmt.Printf("persisting products")
	err = persistProducts(ctx, storeWContext, prodList)
	fmt.Printf("done persisting products")
	return errors.Wrap(err, "persisting the list of retrieved health IT products failed")
}

func makeCHPLProductURL() (*url.URL, error) {
	queryArgs := make(map[string]string)
	fieldStr := strings.Join(fields[:], ",")
	queryArgs["fields"] = fieldStr

	chplURL, err := makeCHPLURL(chplAPICertProdListPath, queryArgs)
	if err != nil {
		return nil, errors.Wrap(err, "creating the URL to query CHPL failed")
	}

	return chplURL, nil
}

func getProductJSON(ctx context.Context, client *httpclient.Client) ([]byte, error) {
	chplURL, err := makeCHPLProductURL()
	if err != nil {
		return nil, errors.Wrap(err, "creating the URL to query CHPL failed")
	}

	// request ceritified products list
	req, err := http.NewRequest("GET", chplURL.String(), nil)
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "making the GET request to the CHPL server failed")
	}
	defer resp.Body.Close()

	if resp == nil {
		return nil, errors.New("CHPL certified products request had nil response")
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New("CHPL certified products request responded with status: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Reading the CHPL response body failed")
	}

	return body, nil
}

func convertProductJSONToObj(ctx context.Context, prodJSON []byte) (*chplCertifiedProductList, error) {
	var prodList chplCertifiedProductList

	// don't unmarshal the JSON if the context has ended
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "Unable to convert product JSON to objects - context ended")
	}

	err := json.Unmarshal(prodJSON, &prodList)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling the JSON into a chplCertifiedProductList object failed.")
	}

	return &prodList, nil
}

func parseHITProd(prod *chplCertifiedProduct) (*endpointmanager.HealthITProduct, error) {
	dbProd := endpointmanager.HealthITProduct{
		Name:                  prod.Product,
		Version:               prod.Version,
		Developer:             prod.Developer,
		CertificationStatus:   prod.CertificationStatus,
		CertificationDate:     time.Unix(prod.CertificationDate/1000, 0),
		CertificationEdition:  prod.Edition,
		CHPLID:                prod.ChplProductNumber,
		CertificationCriteria: strings.Split(prod.CriteriaMet, delimiter1),
	}

	apiURL, err := getAPIURL(prod.APIDocumentation)
	if err != nil {
		return nil, errors.Wrap(err, "retreiving the API URL from the health IT product API documentation list failed")
	}
	dbProd.APIURL = apiURL

	return &dbProd, nil
}

func persistProducts(ctx context.Context, store endpointmanager.HealthITProductStoreWithContext, prodList *chplCertifiedProductList) error {
	for i, prod := range prodList.Results {

		if i%100 == 0 {
			log.Infof("Processing product %d/%d", i, len(prodList.Results))
		}

		ch := make(chan error)
		go func() { ch <- persistProduct(ctx, store, &prod) }()

		select {
		case <-ctx.Done():
			<-ch
			return errors.Wrapf(ctx.Err(), "persisted %d out of %d products before context ended", i, len(prodList.Results))
		case err := <-ch:
			if err != nil {
				log.Warn(err)
				continue
			}
		}
	}

	log.Info("Done processing products")
	return nil
}

func persistProduct(ctx context.Context,
	store endpointmanager.HealthITProductStoreWithContext,
	prod *chplCertifiedProduct) error {

	newDbProd, err := parseHITProd(prod)
	if err != nil {
		return err
	}
	existingDbProd, err := store.GetHealthITProductUsingNameAndVersionWithContext(ctx, prod.Product, prod.Version)

	if err == sql.ErrNoRows { // need to add new entry
		err = store.AddHealthITProductWithContext(ctx, newDbProd)
		if err != nil {
			return errors.Wrap(err, "adding health IT product to store failed")
		}
	} else if err != nil { // error thrown other than no entry exists
		return errors.Wrap(err, "getting health IT product from store failed")
	} else if !existingDbProd.Equal(newDbProd) {
		// changes exist. these may be due to products that have been certified multiple times w/in chpl.
		// we only care about the latest certification information for a particular version of software.

		needsUpdate, err := prodNeedsUpdate(existingDbProd, newDbProd)
		if err != nil {
			return errors.Wrap(err, "determining if a health IT product needs updating within the store failed")
		}

		if needsUpdate {
			existingDbProd.Update(newDbProd)
			err = store.UpdateHealthITProductWithContext(ctx, existingDbProd)
			if err != nil {
				return errors.Wrap(err, "updating health IT product to store failed")
			}
		}
	}
	return nil
}

// assumes that criteria/url chunks are delimited by '☺' and that criteria and url are separated by '☹'.
func getAPIURL(apiDocStr string) (string, error) {
	if len(apiDocStr) == 0 {
		return "", nil
	}

	apiDocStrs := strings.Split(apiDocStr, delimiter1)
	apiCritAndURL := strings.Split(apiDocStrs[0], delimiter2)
	if len(apiCritAndURL) == 2 {
		apiURL := apiCritAndURL[1]
		// check that it's a valid URL
		_, err := url.ParseRequestURI(apiURL)
		if err != nil {
			return "", errors.Wrap(err, "the URL in the health IT product API documentation string is not valid")
		}
		return apiURL, nil
	}

	return "", errors.New("unexpected format for api doc string")
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
		return false, errors.Wrap(err, "unable to make certification edition into an integer - expect certification edition to be a year")
	}
	newCertEdition, err := strconv.Atoi(newDbProd.CertificationEdition)
	if err != nil {
		return false, errors.Wrap(err, "unable to make certification edition into an integer - expect certification edition to be a year")
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
