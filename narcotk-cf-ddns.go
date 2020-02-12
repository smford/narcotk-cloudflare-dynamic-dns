package main

import (
	"flag"
	"fmt"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// format generated from https://golang.org/src/time/format.go
	layoutCF = "2006-01-02 15:04:05.000000 -0700 MST"
)

var (
	apiKey       string
	dnsname      string
	dodebug      bool
	domain       string
	ipstring     string
	newdnsrecord cloudflare.DNSRecord
	recordtype   string
	ttl          string
	user         string

	ipproviderlist = map[string]string{
		"aws":      "https://checkip.amazonaws.com",
		"ipify":    "https://api.ipify.org",
		"my-ip.io": "https://api.my-ip.io/ip",
	}

	recordtypes = []string{"A", "AAAA", "CAA", "CERT", "CNAME", "DNSKEY", "DS", "LOC", "MX", "NAPTR", "NS", "PTR", "SMIMEA", "SPF", "SRV", "SSHFP", "TLSA", "TXT", "URI"}
)

func init() {
	flag.Bool("cfproxy", false, "Make Cloudflare proxy the record, default = false")
	flag.Bool("debug", false, "Display debug information")
	flag.Bool("displayconfig", false, "Display configuration")
	flag.Bool("doit", false, "Disable dry-run and make changes")
	flag.String("domain", "narco.tk", "DNS Domain, default = narco.tk")
	flag.Bool("force", false, "Force update")
	flag.Bool("getip", false, "Get external IPS, can be used with --ipprovider, or \"all\" for all providers")
	flag.Bool("help", false, "Display Help")
	flag.String("host", "test1", "Hostname, default = test1")
	flag.String("ipv4", "", "IPv4 address to use, rather than auto detecting it")
	flag.String("ipprovider", "aws", "Provider of your external IP, \"aws\", \"ipify\" or \"my-ip.io\", default = aws")
	flag.String("ttl", "300", "TTL in seconds (30-600, or auto) for DNS record, default = 300")
	flag.String("type", "A", "Record type, default = \"A\"")
	flag.Bool("typelist", false, "List record types")
	flag.Int("wait", 300, "Seconds to wait since last modificaiton, default = 300")

	viper.SetEnvPrefix("CF")
	viper.BindEnv("API_EMAIL")
	viper.BindEnv("API_KEY")
	viper.BindEnv("DOMAIN")
	viper.BindEnv("HOST")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	dodebug = viper.GetBool("debug")

	if viper.GetBool("help") {
		displayHelp()
		os.Exit(0)
	}

	apiKey = viper.GetString("API_KEY")
	dnsname = viper.GetString("HOST")
	domain = viper.GetString("DOMAIN")
	user = viper.GetString("API_EMAIL")
}

func displayHelp() {
	fmt.Println("")
	fmt.Println("cf-ddns - Dynamic DNS updater for Cloudflare")
	fmt.Println("")
	fmt.Println("    --cfproxy               Make Cloudflare proxy the record, default = false")
	fmt.Println("    --displayconfig         Display configurtation")
	fmt.Println("    --doit                  Disable dry-run and make changes")
	fmt.Println("    --domain                Domain")
	fmt.Println("    --force                 Force update")
	fmt.Println("    --getip                 Get external IPS, can be used with --ipprovider, or \"all\" for all providers")
	fmt.Println("    --help                  Help")
	fmt.Println("    --host                  Host")
	fmt.Println("    --ipv4                  IPv4 address to use, rather than auto detecting it")
	fmt.Println("    --ipprovider            Provider of your external IP, \"aws\", \"ipify\" or \"my-ip.io\", default = aws")
	fmt.Println("    --ttl                   TTL in seconds (30-600, or auto) for DNS record, default = 300")
	fmt.Println("    --type                  Record type, default = \"A\"")
	fmt.Println("    --typelist              List record types")
	fmt.Println("    --wait                  Seconds to wait since last modification, default = 300")
	fmt.Println("")
}

func main() {
	if viper.GetBool("displayconfig") {
		displayConfig()
		os.Exit(0)
	}

	if viper.GetBool("typelist") {
		displaytypelist()
		os.Exit(0)
	}

	if !validateipprovider(viper.GetString("ipprovider")) {
		fmt.Printf("--ipprovider %s is not a valid provider\n", viper.GetString("ipprovider"))
		os.Exit(1)
	}

	if viper.GetBool("getip") {
		if strings.ToLower(viper.GetString("ipprovider")) == "all" {
			for k, v := range ipproviderlist {
				fmt.Printf("%s [%s] %s\n", k, v, getIP(k))
			}
		} else {
			fmt.Println(getIP(viper.GetString("ipprovider")))
		}
		os.Exit(0)
	}

	if len(viper.GetString("ipv4")) == 0 {
		ipstring = getIP(viper.GetString("ipprovider"))
	} else {
		if validateipv4(viper.GetString("ipv4")) {
			ipstring = viper.GetString("ipv4")
		} else {
			fmt.Printf("--ipv4 %s is not a valid ip\n", viper.GetString("ipv4"))
			os.Exit(1)
		}
	}

	if validaterecordtype(viper.GetString("type")) {
		recordtype = strings.ToUpper(viper.GetString("type"))
	} else {
		fmt.Printf("--type %s is not valid\n", viper.GetString("type"))
		os.Exit(1)
	}

	if validatettl(viper.GetString("ttl")) {
		ttl = viper.GetString("ttl")
	} else {
		fmt.Printf("--ttl %s is not valid, must be between 30 and 600, or \"auto\"\n", viper.GetString("ttl"))
		os.Exit(1)
	}

	newdnsrecord.Type = recordtype
	newdnsrecord.Name = dnsname
	newdnsrecord.Content = ipstring
	newdnsrecord.Proxied = viper.GetBool("cfproxy")

	if strings.ToLower(ttl) != "auto" {
		newdnsrecord.TTL, _ = strconv.Atoi(ttl)
	}

	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		fmt.Println(err)
		return
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		fmt.Println(err)
		return
	}

	findhost := cloudflare.DNSRecord{Name: dnsname + "." + domain}

	// Fetch all DNS records for example.org that match host
	recs, err := api.DNSRecords(zoneID, findhost)
	if err != nil {
		fmt.Println("Could not retrieve DNS records: ", err)
		os.Exit(1)
		return
	}

	if len(recs) == 0 {
		fmt.Printf("No record found for %s.%s, CREATING.\n", dnsname, domain)
		creatednsrecord(*api, zoneID, newdnsrecord)

	} else {
		if dodebug == true {
			fmt.Println("UPDATING DNS RECORD")
		}

		for _, r := range recs {
			if dodebug == true {
				fmt.Printf("ID: %s %s: %s %s %d %s/%s\n", r.ID, r.Name, r.Type, r.Content, r.TTL, r.CreatedOn, r.ModifiedOn)
				fmt.Printf("last modified: %s\n", r.ModifiedOn)
			}

			if r.Content == newdnsrecord.Content {
				if dodebug == true {
					fmt.Println("DNS record up to date, not updating")
				}
			} else {
				if dodebug == true {
					fmt.Println("DNS record needs updating")
				}

				lastmodified, _ := time.Parse(layoutCF, r.ModifiedOn.String())
				timenow := time.Now().UTC()
				timediff := timenow.Sub(lastmodified).Round(time.Second).Seconds()

				if dodebug == true {
					fmt.Println("Time difference information:")
					fmt.Println("       now:", timenow)
					fmt.Println("  modified:", lastmodified)
					fmt.Println("difference:", timediff)
					fmt.Println("      wait:", viper.GetInt("wait"))
				}

				if (int64(timediff) >= int64(viper.GetInt("wait"))) || viper.GetBool("force") {
					fmt.Printf("Updating DNS record because it was last updated more than %d seconds ago and wait time set to %d seconds\n", int64(timediff), int64(viper.GetInt("wait")))

					if dodebug == true {
						fmt.Println("newdnsrecord=", newdnsrecord)
					}

					updatednsrecord(*api, zoneID, r.ID, newdnsrecord)
				} else {
					fmt.Printf("Not updating dns as it was only updated %d seconds ago\n", int64(timediff))
				}

			}

		}

		// end updating record
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

	return returnip
}

func updatednsrecord(myapi cloudflare.API, zoneID string, recordID string, newdnsrecord cloudflare.DNSRecord) {
	if viper.GetBool("doit") {
		err := myapi.UpdateDNSRecord(zoneID, recordID, newdnsrecord)
		if err != nil {
			fmt.Println("Could not update DNS record:", err)
			os.Exit(1)
			return
		}
		fmt.Println("Updated DNS record: ", recordID)
	} else {
		fmt.Println("Dry run complete")
	}
}

func creatednsrecord(myapi cloudflare.API, zoneID string, newdnsrecord cloudflare.DNSRecord) {
	if viper.GetBool("doit") {
		recs, err := myapi.CreateDNSRecord(zoneID, newdnsrecord)
		if err != nil {
			fmt.Println("Could not create DNS record: ", err)
			os.Exit(1)
			return
		}

		if dodebug == true {
			fmt.Println(recs)
		}

		fmt.Println("Created dns record")
	} else {
		fmt.Println("Dry run complete")
	}
}

func validatettl(checkttl string) bool {
	if strings.ToLower(checkttl) == "auto" {
		return true
	}

	tempttl, err := strconv.Atoi(checkttl)
	if err == nil {
		if (tempttl >= 30) && (tempttl <= 600) {
			return true
		}
	}

	return false
}

func validaterecordtype(recordtype string) bool {
	recordtype = strings.ToUpper(recordtype)

	for _, item := range recordtypes {
		if item == recordtype {
			return true
		}
	}

	return false
}

func validateipprovider(ipname string) bool {
	ipname = strings.ToLower(ipname)

	if ipname == "all" {
		return true
	}

	for k := range ipproviderlist {
		if k == ipname {
			return true
		}
	}

	return false
}

func validateipv4(ipv4 string) bool {
	if dodebug == true {
		fmt.Println("Validating IPv4 address")
	}

	if net.ParseIP(ipv4) != nil {
		return true
	}

	return false
}

func displaytypelist() {
	sort.Strings(recordtypes)
	for i := 0; i < len(recordtypes); i++ {
		fmt.Println(recordtypes[i])
	}
}
