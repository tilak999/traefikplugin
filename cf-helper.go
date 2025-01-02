package traefikplugin

import (
	"io"
	"log"
	"net/http"
	"strings"
)

func PurgeCache(config *Config, detectedHeaders map[string]string, logger *log.Logger) {

	url_template := "https://api.cloudflare.com/client/v4/zones/$zoneId/purge_cache"
	url := strings.Replace(url_template, "$zoneId", config.CloudflareZone, 1)

	payload := strings.NewReader("{\n  \"purge_everything\": true\n}")

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		logger.Println(err)
		return
	}

	authToken := strings.Replace("Bearer $token", "$token", config.CloudflareToken, 1)
	req.Header.Add("Authorization", authToken)
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Println(err)
		return
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Println(err)
		return
	}

	if res.StatusCode == 200 {
		logger.Print("Cloudflare cache purged: OK")
		if config.DryRun {
			logger.Println(string(body))
		}
		return
	}

	logger.Printf("Request completed with status != 200: actual status [%d]", res.StatusCode)
	logger.Println(string(body))

}
