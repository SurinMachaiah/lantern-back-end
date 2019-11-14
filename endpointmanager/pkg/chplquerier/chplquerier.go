package chplquerier

import (
	"context"
	"time"

	"github.com/spf13/viper"
)

type CHPLQuerier interface {
}

var chpl_url string = "https://chpl.healthit.gov/rest/"

func GetCHPLProducts() {
	ctx := context.Background()

	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancelFunc()

	api_key := viper.GetString("chplapikey")

}
