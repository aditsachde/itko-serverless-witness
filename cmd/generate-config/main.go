package main

import (
	"encoding/json"
	"fmt"
	"log"

	"golang.org/x/mod/sumdb/note"
)

type Config struct {
	DatabaseUrl string            `json:"database_url"`
	Keys        map[string]string `json:"keys"`
}

var regions = []string{
	"northamerica-northeast1",
	"northamerica-northeast2",

	"us-central1",

	"us-east1",
	"us-east4",
	"us-east5",

	"us-west1",
	"us-west2",
	"us-west3",
	"us-west4",

	"us-south1",
}

func main() {
	config := Config{
		DatabaseUrl: "REPLACE",
		Keys:        make(map[string]string),
	}

	fmt.Println("REGION=REPLACE")
	fmt.Println()

	// iterate over the regions
	for _, region := range regions {
		skey, vkey, err := note.GenerateKey(nil, region)
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

	fmt.Printf("\nCONFIG='%s'\n", string(configString))
}
