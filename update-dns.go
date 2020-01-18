package main

import (
	"fmt"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	user         string
	domain       string
	apiKey       string
	newdnsrecord cloudflare.DNSRecord
)

func init() {
	user = os.Getenv("CF_API_EMAIL")
	domain = os.Getenv("CF_DOMAIN")
	apiKey = os.Getenv("CF_API_KEY")
}

func main() {
	res, _ := http.Get("https://api.ipify.org")
	ip, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(ip[:len(ip)]))

	newdnsrecord.Type = "A"
	newdnsrecord.Name = "test1"
	newdnsrecord.Content = string(ip[:len(ip)])
	newdnsrecord.TTL = 300

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

	// localhost := cloudflare.DNSRecord{Content: "127.0.0.1"}

	// Fetch all DNS records for example.org
	recs, err := api.CreateDNSRecord(zoneID, newdnsrecord)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(recs)

	//for _, r := range recs {
	//	fmt.Printf("%s: %s\n", r.Name, r.Content)
	//}

	//fmt.Printf("%s        result.Re

	//
	//for _, r := range records {
	//	fmt.Printf("%s: %s %s %d %s/%s\n", r.Name, r.Type, r.Content, r.TTL, r.CreatedOn, r.ModifiedOn)
	//}
}
