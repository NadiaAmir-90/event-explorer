package main

import (
	"os"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"

	"event-explorer/models"
)

func main() {
	baseURL, err := web.AppConfig.String("ticketmaster_base_url")
	if err != nil || baseURL == "" {
		logs.Error("ticketmaster_base_url is not configured")
		os.Exit(1)
	}

	apiKey, err := web.AppConfig.String("ticketmaster_api_key")
	if err != nil || apiKey == "" {
		logs.Error("ticketmaster_api_key is not configured")
		os.Exit(1)
	}

	cache := models.NewCache()

	apiClient := models.NewAPIClientWithConfig(
		nil,
		cache,
		baseURL,
		apiKey,
	)

	// The API client will be passed to your service/controller.
	_ = apiClient

	web.Run()
}