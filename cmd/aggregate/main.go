package main

import (
	"flag"
	"log"
	"time"
	"web_backend_v2/config"
	"web_backend_v2/db"
	"web_backend_v2/models"
)

func main() {
	dayStr := flag.String("day", "", "YYYY-MM-DD day to aggregate (default: yesterday, UTC)")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := db.InitDB(cfg, true); err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			log.Printf("close db: %v", err)
		}
	}()

	targetDay := time.Now().UTC().AddDate(0, 0, -1) // default: yesterday
	if *dayStr != "" {
		parsed, err := time.Parse("2006-01-02", *dayStr)
		if err != nil {
			log.Fatalf("invalid -day value (want YYYY-MM-DD): %v", err)
		}
		targetDay = parsed.UTC()
	}

	if err := models.AggregateDailyOccupancy(targetDay); err != nil {
		log.Fatalf("aggregate daily occupancy: %v", err)
	}

	log.Printf("aggregated occupancy for %s", targetDay.Format("2006-01-02"))
}

