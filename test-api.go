package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudflare/cloudflare-go"
)

func main() {
	// Construct a new API object
	api, err := cloudflare.New(os.Getenv("CF_API_KEY"), os.Getenv("CF_API_EMAIL"))
	if err != nil {
		log.Fatal(err)
	}

	// Fetch user details on the account
	u, err := api.UserDetails()
	if err != nil {
		log.Fatal(err)
	}
	// Print user details
	fmt.Println(u)

	// Fetch the zone ID
	id, err := api.ZoneIDByName("narco.tk") // Assuming example.com exists in your Cloudflare account already
	if err != nil {
		log.Fatal(err)
	}

	// Fetch zone details
	zone, err := api.ZoneDetails(id)
	if err != nil {
		log.Fatal(err)
	}
	// Print zone details
	fmt.Println(zone)
}
