package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type Config struct {
	DatabaseUrl string            `json:"database_url"`
	Keys        map[string]string `json:"keys"`
}

func ParseConfig() (db_url string, witness_key string, region string, err error) {
	// get the region from "REGION"
	region = os.Getenv("REGION")
	if region == "" {
		return "", "", "", fmt.Errorf("REGION must be set!")
	}

	// try to get a value from the environment var "CONFIG"
	config := os.Getenv("CONFIG")
	if config == "" {
		// try to instead fetch the value from GCP secret manager
		secretName := os.Getenv("CONFIG_SECRET")
		if secretName == "" {
			return "", "", "", fmt.Errorf("One of CONFIG or CONFIG_SECRET must be set!")
		}

		ctx := context.Background()

		// setup the client
		client, err := secretmanager.NewClient(ctx)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to create secretmanager client: %v", err)
		}
		defer client.Close()

		result, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: secretName,
		})
		if err != nil {
			return "", "", "", fmt.Errorf("failed to access secret version: %v", err)
		}

		config = string(result.Payload.Data)
	}

	// parse the config
	var c Config
	err = json.Unmarshal([]byte(config), &c)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse config: %v", err)
	}

	// get the database url
	db_url = c.DatabaseUrl
	if db_url == "" {
		return "", "", "", fmt.Errorf("database_url must be set!")
	}

	// get the witness key
	witness_key = c.Keys[region]
	if witness_key == "" {
		return "", "", "", fmt.Errorf("witness key for region %s must be set!", region)
	}

	return db_url, witness_key, region, nil
}
