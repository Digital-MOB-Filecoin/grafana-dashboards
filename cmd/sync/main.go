package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/Digital-MOB-Filecoin/grafana-dashboards/rpc"
	"github.com/Digital-MOB-Filecoin/grafana-dashboards/scraper"
	"github.com/Digital-MOB-Filecoin/grafana-dashboards/store/postgres"
)

func main() {
	lotusURL := flag.String("lotus-url", envOrDefault("LOTUS_RPC_URL", "http://localhost:1234/rpc/v1"), "Lotus RPC endpoint")
	lotusToken := flag.String("lotus-token", os.Getenv("LOTUS_API_TOKEN"), "Lotus API bearer token")
	dbDSN := flag.String("db", os.Getenv("DATABASE_URL"), "Postgres DSN")
	flag.Parse()

	if *dbDSN == "" {
		log.Fatal("-db or DATABASE_URL is required")
	}

	ctx := context.Background()

	lotusAPI := rpc.NewLotusAPI(rpc.NewHTTPClient(*lotusURL, *lotusToken))

	st, err := postgres.New(ctx, *dbDSN)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer st.Close()

	job := scraper.NewChainSyncJob(lotusAPI, st)
	if err := job.Run(ctx); err != nil {
		log.Fatalf("sync: %v", err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
