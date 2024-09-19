package main

import (
	"context"
	"fmt"
	"log"
	// "math"
	"net"
	"os"
	"time"

	"github.com/transparency-dev/witness/monitoring"
	"github.com/transparency-dev/witness/omniwitness"

	"net/http"
)

type witness struct {
	handler http.Handler
}

func main() {
	// Context
	var o_ctx = context.Background()

	// Config
	db_url, witness_key, region, err := ParseConfig()
	if err != nil {
		log.Fatalln("Failed to parse config:", err)
	}
	log.Println("Region:", region)

	var o_operatorConfig omniwitness.OperatorConfig = omniwitness.OperatorConfig{
		WitnessKey: witness_key,

		// constrain the feeder and distributor to only run once on initialization
		FeedInterval:       time.Second,
		DistributeInterval: time.Second,
		// FeedInterval:       time.Duration(math.MaxInt64),
		// DistributeInterval: time.Duration(math.MaxInt64),
	}

	// Persistence
	var o_p *PgPersistence
	o_p, err = NewPgPersistence(o_ctx, db_url, region)
	if err != nil {
		log.Fatalln("Failed to open persistence DB:", err)
	}

	// Listener
	port := os.Getenv("PORT")
	if port == "" {
		port = "60606"
		log.Printf("defaulting to port %s", port)
	}
	addr := fmt.Sprintf(":%s", port)

	var o_httpListener net.Listener
	o_httpListener, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to start listener:", err)
	}

	// Outbound
	var o_httpClient *http.Client = &http.Client{}

	// Metrics
	monitoring.SetMetricFactory(monitoring.InertMetricFactory{})

	// Start
	log.Println("starting server...")
	err = omniwitness.Main(o_ctx, o_operatorConfig, o_p, o_httpListener, o_httpClient)
	log.Fatalln("Omniwitness exited:", err)
}
