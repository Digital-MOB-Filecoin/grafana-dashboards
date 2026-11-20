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

const (
	defaultEndHeight  = 2383680
	finalityEpochs    = 900
)

func main() {
	lotusURL := flag.String("lotus-url", envOrDefault("LOTUS_RPC_URL", "http://localhost:1234/rpc/v1"), "Lotus RPC endpoint")
	lotusToken := flag.String("lotus-token", os.Getenv("LOTUS_API_TOKEN"), "Lotus API bearer token")
	dbDSN := flag.String("db", os.Getenv("DATABASE_URL"), "Postgres DSN")
	from := flag.Int64("from", 0, "start height (must be finalised); defaults to current chain head")
	to := flag.Int64("to", 0, "stop ingesting at this height (default: latest stored height, or 2383680 if db is empty)")
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

	// Always fetch head — needed for finality check and as default start.
	head, err := lotusAPI.ChainGetHead(ctx)
	if err != nil {
		log.Fatalf("ChainGetHead: %v", err)
	}

	// Resolve start tipset.
	start := head
	if *from != 0 {
		finalised := head.Height - finalityEpochs
		if *from > finalised {
			log.Fatalf("-from %d is not yet finalised (head=%d, finality=%d, earliest finalised=%d)",
				*from, head.Height, finalityEpochs, finalised)
		}
		start, err = lotusAPI.ChainGetTipSetByHeight(ctx, *from, nil)
		if err != nil {
			log.Fatalf("ChainGetTipSetByHeight(%d): %v", *from, err)
		}
	}

	// Resolve end height.
	endHeight := *to
	if endHeight == 0 {
		latest, found, err := st.LatestBlockHeight(ctx)
		if err != nil {
			log.Fatalf("get latest stored height: %v", err)
		}
		if found {
			endHeight = latest
		} else {
			endHeight = defaultEndHeight
		}
	}

	if endHeight >= start.Height {
		log.Fatalf("-to %d must be lower than start height %d", endHeight, start.Height)
	}

	job := scraper.NewIngestJob(lotusAPI, st)

	log.Printf("ingest: walking from height %d down to %d", start.Height, endHeight)
	if err := job.RunFrom(ctx, start, endHeight); err != nil {
		log.Fatalf("ingest: %v", err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
