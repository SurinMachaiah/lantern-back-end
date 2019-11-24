package http

import (
	gohttp "net/http"
	"time"
)

// Need to define timeout or else it is infinite
var netClient = &gohttp.Client{
	Timeout: time.Second * 35,
}

func GetClient() *gohttp.Client {
	return netClient
}
