package chplquerier

import (
	"net/url"

	"github.com/spf13/viper"
)

const chplDomain string = "https://chpl.healthit.gov"
const chplAPIPath string = "/rest"

func makeCHPLURL(path string, queryArgs map[string]string) (*url.URL, error) {
	queryArgsToSend := url.Values{}
	chplURL, err := url.Parse(chplDomain)
	if err != nil {
		return nil, err
	}

	apiKey := viper.GetString("chplapikey")
	queryArgsToSend.Set("api_key", apiKey)
	for k, v := range queryArgs {
		queryArgsToSend.Set(k, v)
	}

	chplURL.RawQuery = queryArgsToSend.Encode()
	chplURL.Path = chplAPIPath + path

	return chplURL, nil
}
