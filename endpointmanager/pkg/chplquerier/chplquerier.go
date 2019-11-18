package chplquerier

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/spf13/viper"

	"errors"
)

type CHPLQuerier interface {
}

type CHPLCertifiedProduct struct {
	ID                       int         `json:"id"`
	ChplProductNumber        string      `json:"chplProductNumber"`
	Edition                  string      `json:"edition"`
	Acb                      string      `json:"acb"`
	AcbCertificationID       string      `json:"acbCertificationId"`
	PracticeType             string      `json:"practiceType"`
	Developer                string      `json:"developer"`
	DeveloperStatus          string      `json:"developerStatus"`
	Product                  string      `json:"product"`
	Version                  string      `json:"version"`
	CertificationDate        int64       `json:"certificationDate"`
	CertificationStatus      string      `json:"certificationStatus"`
	SurveillanceCount        int         `json:"surveillanceCount"`
	OpenNonconformityCount   int         `json:"openNonconformityCount"`
	ClosedNonconformityCount int         `json:"closedNonconformityCount"`
	PreviousDevelopers       interface{} `json:"previousDevelopers"`
	CriteriaMet              string      `json:"criteriaMet"`
	CqmsMet                  interface{} `json:"cqmsMet"`
}

type CHPLCertifiedProductList struct {
	Results []CHPLCertifiedProduct `json:"results"`
}

var chplDomain string = "https://chpl.healthit.gov"
var chplAPIPath string = "/rest"
var chplAPICertProdList string = "/collections/certified_products"

func GetCHPLProducts() error {
	err := getCertifiedProductCollection()

	return err
}

func makeCHPLURL() (*url.URL, error) {
	queryArgs := make(url.Values)
	chplURL, err := url.Parse(chplDomain)
	if err != nil {
		return nil, err
	}
	apiKey := viper.GetString("chplapikey")
	queryArgs.Set("api_key", apiKey)
	chplURL.RawQuery = queryArgs.Encode()

	return chplURL, nil
}

func getCertifiedProductCollection() error {
	chplURL, err := makeCHPLURL()
	if err != nil {
		return err
	}

	// request ceritified products list
	chplURL.Path = chplAPIPath + chplAPICertProdList
	chplURLStr := chplURL.String()
	println(chplURLStr)
	println(chplURL)
	resp, err := http.Get(chplURL.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp == nil {
		return errors.New("CHPL certified products request had nil response")
	} else if resp.StatusCode != http.StatusOK {
		return errors.New("CHPL certified products request responded with status: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var prodList CHPLCertifiedProductList

	err = json.Unmarshal(body, &prodList)
	if err != nil {
		return err
	}

	return nil
}
