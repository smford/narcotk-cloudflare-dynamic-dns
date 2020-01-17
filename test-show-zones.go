package main

import (
	"fmt"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"os"
)

var (
	user   string
	domain string
	apiKey string
)

func init() {
	user = os.Getenv("CF_API_EMAIL")
	domain = "narco.tk"
	apiKey = os.Getenv("CF_API_KEY")
}

func main() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Fetch the zone ID for zone example.org
	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Fetch all DNS records for example.org
	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{})
	if err != nil {
		fmt.Println(err)
		return
	}

	//
	for _, r := range records {
		fmt.Printf("%s: %s %s %d %s/%s\n", r.Name, r.Type, r.Content, r.TTL, r.CreatedOn, r.ModifiedOn)
	}
}
