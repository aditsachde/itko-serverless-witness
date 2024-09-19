package main

import (
	"context"
	"flag"
	"log"

	"github.com/jackc/pgx/v5"
)

var (
	url = flag.String("db-url", "", "DB connection URL")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	config, err := pgx.ParseConfig(*url)
	if err != nil {
		log.Fatalln("Failed to parse config:", err)
	}
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalln("Failed to connect:", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx,
		`CREATE TABLE chkpts (
    		region VARCHAR NOT NULL,
    		log_id VARCHAR NOT NULL,
    		chkpt BYTEA,
    		range BYTEA,
    		PRIMARY KEY (region, log_id)
		)`)

	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	log.Println("Table created")
}
