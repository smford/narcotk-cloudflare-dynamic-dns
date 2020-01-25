package main

import (
	"flag"
	"fmt"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"time"
)

const (
	// format generated from https://golang.org/src/time/format.go
	layoutCF = "2006-01-02 15:04:05.000000 -0700 MST"
)

var (
	user         string
	domain       string
	apiKey       string
	dnsname      string
	newdnsrecord cloudflare.DNSRecord
)

func init() {
	flag.Bool("displayconfig", false, "Display configuration")
	flag.Bool("updatedns", false, "Update DNS")
	flag.Bool("help", false, "Display Help")
	flag.String("domain", "narco.tk", "DNS Domain, default = narco.tk")
	flag.String("host", "test1", "Hostname, default = test1")
	flag.Int("wait", 300, "Seconds to wait since last modificaiton")
	viper.SetEnvPrefix("CF")
	viper.BindEnv("API_EMAIL")
	viper.BindEnv("API_KEY")
	//flag.String("appid", "", "appid")

	//user = viper.GetString("API_EMAIL")
	//user = os.Getenv("CF_API_EMAIL")
	//domain = os.Getenv("CF_DOMAIN")
	//apiKey = os.Getenv("CF_API_KEY")
	//dnsname = os.Getenv("CF_DNSNAME")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if viper.GetBool("help") {
		displayHelp()
		os.Exit(0)
	}

	user = viper.GetString("API_EMAIL")
	apiKey = viper.GetString("API_KEY")
	domain = viper.GetString("domain")
	dnsname = viper.GetString("host")
}

func displayHelp() {
	fmt.Println("\ncf-ddns - Dynamic DNS updater for Cloudflare\n")
	fmt.Println("    --help                  Help")
	fmt.Println("    --displayconfig         Display configurtation")
	fmt.Println("    --domain                Domain")
	fmt.Println("    --host                  Host")
	fmt.Println("    --updatedns             Should I update the dns?")
	fmt.Println("    --wait                  Seconds to wait since last modification")
}

func main() {
	if viper.GetBool("displayconfig") {
		displayConfig()
		os.Exit(0)
	}
	res, _ := http.Get("https://api.ipify.org")
	ip, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(ip[:len(ip)]))

	ipstring := string(ip[:len(ip)])

	newdnsrecord.Type = "A"
	newdnsrecord.Name = dnsname
	newdnsrecord.Content = ipstring
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

	//findhost := cloudflare.DNSRecord{Content: "82.34.44.205"}

	findhost := cloudflare.DNSRecord{Name: dnsname + "." + domain}

	// Fetch all DNS records for example.org
	recs, err := api.DNSRecords(zoneID, findhost)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, r := range recs {
		fmt.Printf("%s: %s %s %d %s/%s\n", r.Name, r.Type, r.Content, r.TTL, r.CreatedOn, r.ModifiedOn)
		fmt.Printf("last modified: %s\n", r.ModifiedOn)
		// temportarily over ride dns record to be the real one for my home
		//newdnsrecord.Content = "82.34.44.205"

		if r.Content == newdnsrecord.Content {
			fmt.Println("current = new: ignoring")
		} else {
			fmt.Println("needs updating")

			lastmodified, _ := time.Parse(layoutCF, r.ModifiedOn.String())
			timenow := time.Now().UTC()
			timediff := timenow.Sub(lastmodified).Round(time.Second).Seconds()

			fmt.Println("       now:", timenow)
			fmt.Println("  modified:", lastmodified)
			fmt.Println("difference:", timediff)
			fmt.Println("      wait:", viper.GetInt("wait"))

			if int64(timediff) >= int64(viper.GetInt("wait")) {
				fmt.Printf("updating dns because it was last updated more than %d seconds ago and wait time set to %d seconds\n", int64(timediff), int64(viper.GetInt("wait")))
				if viper.GetBool("updatedns") {
					recs, err := api.CreateDNSRecord(zoneID, newdnsrecord)
					if err != nil {
						fmt.Println(err)
						return
					}
					fmt.Println(recs)
				} else {
					fmt.Println("updatedns=false")
				}

			} else {
				fmt.Printf("not updating dns as it was only updated %d seconds ago\n", int64(timediff))
			}

		}

	}
}

func displayConfig() {
	allmysettings := viper.AllSettings()
	var keys []string
	for k := range allmysettings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println("CONFIG:", k, ":", allmysettings[k])
	}
}
