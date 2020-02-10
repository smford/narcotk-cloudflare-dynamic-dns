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
	"strings"
	"time"
)

const (
	// format generated from https://golang.org/src/time/format.go
	layoutCF = "2006-01-02 15:04:05.000000 -0700 MST"
)

var (
	user         string
	dodebug      bool
	domain       string
	apiKey       string
	dnsname      string
	ipstring     string
	newdnsrecord cloudflare.DNSRecord

	ipproviderlist = map[string]string{
		"aws":      "https://checkip.amazonaws.com",
		"ipify":    "https://api.ipify.org",
		"my-ip.io": "https://api.my-ip.io/ip",
	}
)

func init() {
	flag.Bool("cfproxy", false, "Make Cloudflare proxy the record, default = false")
	flag.Bool("debug", false, "Display debug information")
	flag.Bool("displayconfig", false, "Display configuration")
	flag.Bool("updatedns", false, "Update DNS")
	flag.Bool("getip", false, "Get external IPS, can be used with --ipprovider, or \"all\" for all providers")
	flag.Bool("help", false, "Display Help")
	flag.String("domain", "narco.tk", "DNS Domain, default = narco.tk")
	flag.String("host", "test1", "Hostname, default = test1")
	flag.String("ipprovider", "aws", "Provider of your external IP, \"aws\", \"ipify\" or \"my-ip.io\", default = aws")
	flag.Int("wait", 300, "Seconds to wait since last modificaiton")
	viper.SetEnvPrefix("CF")
	viper.BindEnv("API_EMAIL")
	viper.BindEnv("API_KEY")
	viper.BindEnv("HOST")
	viper.BindEnv("DOMAIN")
	//flag.String("appid", "", "appid")

	//user = viper.GetString("API_EMAIL")
	//user = os.Getenv("CF_API_EMAIL")
	//domain = os.Getenv("CF_DOMAIN")
	//apiKey = os.Getenv("CF_API_KEY")
	//dnsname = os.Getenv("CF_DNSNAME")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	dodebug = viper.GetBool("debug")

	if viper.GetBool("help") {
		displayHelp()
		os.Exit(0)
	}

	user = viper.GetString("API_EMAIL")
	apiKey = viper.GetString("API_KEY")
	domain = viper.GetString("DOMAIN")
	dnsname = viper.GetString("HOST")
}

func displayHelp() {
	fmt.Println("\ncf-ddns - Dynamic DNS updater for Cloudflare\n")
	fmt.Println("    --help                  Help")
	fmt.Println("    --displayconfig         Display configurtation")
	fmt.Println("    --domain                Domain")
	fmt.Println("    --getip                 Get external IPS, can be used with --ipprovider, or \"all\" for all providers")
	fmt.Println("    --host                  Host")
	fmt.Println("    --ipprovider            Provider of your external IP, \"aws\", \"ipify\" or \"ip.io\", default = aws")
	fmt.Println("    --updatedns             Should I update the dns?")
	fmt.Println("    --wait                  Seconds to wait since last modification")
	fmt.Println("    --cfproxy               Make Cloudflare proxy the record, default = false")
}

func main() {
	if viper.GetBool("displayconfig") {
		displayConfig()
		os.Exit(0)
	}
	//res, _ := http.Get("https://api.ipify.org")
	//ip, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(ip[:len(ip)]))
	//ipstring := string(ip[:len(ip)])

	if viper.GetBool("getip") {
		if strings.ToLower(viper.GetString("ipprovider")) == "all" {
			for k, v := range ipproviderlist {
				fmt.Printf("%s [%s] %s\n", k, v, getIP(k))
			}
		} else {
			fmt.Println(getIP(viper.GetString("ipprovider")))
			//fmt.Println("this is the else")
		}
		os.Exit(0)
	}

	ipstring = getIP(viper.GetString("ipprovider"))

	newdnsrecord.Type = "A"
	newdnsrecord.Name = dnsname
	newdnsrecord.Content = ipstring
	newdnsrecord.TTL = 300
	newdnsrecord.Proxied = viper.GetBool("cfproxy")

	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Printf("apitype=%R\n", api)

	// Fetch the zone ID for zone example.org
	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("zoneidtype=%T\n", zoneID)

	//findhost := cloudflare.DNSRecord{Content: "82.34.44.205"}

	findhost := cloudflare.DNSRecord{Name: dnsname + "." + domain}

	// Fetch all DNS records for example.org
	recs, err := api.DNSRecords(zoneID, findhost)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(recs) == 0 {
		fmt.Printf("No record found for %s.%s, creating.\n", dnsname, domain)
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
					//recs, err := api.CreateDNSRecord(zoneID, newdnsrecord)
					//if err != nil {
					//	fmt.Println(err)
					//	return
					//}
					//fmt.Println(recs)
					fmt.Println("newdnsrecord=", newdnsrecord)
					updatednsrecord(*api, zoneID, newdnsrecord)
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

func getIP(ipprovider string) string {
	ipprovider = strings.ToLower(ipprovider)

	res, err := http.Get(ipproviderlist[ipprovider])
	ip, _ := ioutil.ReadAll(res.Body)

	returnip := string(ip[:len(ip)])
	returnip = strings.TrimSuffix(returnip, "\n")

	if dodebug == true {
		fmt.Println("using: ", ipprovider)
		fmt.Println("ip:", returnip)
	}

	if err != nil {
		fmt.Printf("Cannot discern public IP using: %s", ipprovider)
		fmt.Println(err)
		os.Exit(2)
	}

	//return string(ip[:len(ip)])
	//return string(ip)

	return returnip
}

func updatednsrecord(myapi cloudflare.API, zoneID string, newdnsrecord cloudflare.DNSRecord) {
	//func updatednsrecord(myapi cloudflare.API, host string, domain string) {
	if viper.GetBool("updatedns") {
		recs, err := myapi.CreateDNSRecord(zoneID, newdnsrecord)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(recs)
	} else {
		fmt.Println("updatedns=false")
	}
}
