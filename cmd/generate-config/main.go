package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"golang.org/x/mod/sumdb/note"
)

type Config struct {
	DatabaseUrl string            `json:"database_url"`
	Keys        map[string]string `json:"keys"`
}

var (
	db_url = flag.String("db-url", "", "Database connection string")
	suffix = flag.String("suffix", "", "Name suffix")
)

var regions = []string{
	"africa-south1",
	"asia-east1",
	"asia-east2",
	"asia-northeast1",
	"asia-northeast2",
	"asia-northeast3",
	"asia-south1",
	"asia-south2",
	"asia-southeast1",
	"asia-southeast2",
	"australia-southeast1",
	"australia-southeast2",
	"europe-central2",
	"europe-north1",
	"europe-southwest1",
	"europe-west1",
	"europe-west10",
	"europe-west12",
	"europe-west2",
	"europe-west3",
	"europe-west4",
	"europe-west6",
	"europe-west8",
	"europe-west9",
	"me-central1",
	"me-central2",
	"me-west1",
	"northamerica-northeast1",
	"northamerica-northeast2",
	"southamerica-east1",
	"southamerica-west1",
	"us-central1",
	"us-east1",
	"us-east4",
	"us-east5",
	"us-south1",
	"us-west1",
	"us-west2",
	"us-west3",
	"us-west4",
}

func main() {
	flag.Parse()

	if *db_url == "" {
		log.Fatalf("Database connection string is required")
	}

	config := Config{
		DatabaseUrl: *db_url,
		Keys:        make(map[string]string),
	}

	fmt.Println("REGION=REPLACE")
	fmt.Println()

	// iterate over the regions
	for _, region := range regions {
		skey, vkey, err := note.GenerateKey(nil, region+*suffix)
		if err != nil {
			log.Fatalf("Failed to generate keys: %v", err)
		}

		_, err = note.NewSigner(skey)
		if err != nil {
			log.Fatalf("Validation error forming a signer: %v", err)
		}

		config.Keys[region] = skey
		fmt.Println("#", vkey)
	}

	// Serialize the config
	configString, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Failed to serialize config: %v", err)
	}

	fmt.Printf("\nCONFIG=%s\n", base64.StdEncoding.EncodeToString(configString))
}
