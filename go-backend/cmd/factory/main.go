package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"tid/go-backend/internal/api"
	"tid/go-backend/internal/config"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)

	ingest := flag.NewFlagSet("ingest", flag.ExitOnError)
	analyze := flag.NewFlagSet("analyze", flag.ExitOnError)
	clip := flag.NewFlagSet("clip", flag.ExitOnError)
	schedule := flag.NewFlagSet("schedule", flag.ExitOnError)
	tick := flag.NewFlagSet("tick", flag.ExitOnError)

	url := ingest.String("url", "", "YouTube URL")
	title := ingest.String("title", "", "Speaker or episode title")
	podcast := ingest.String("podcast", "", "Podcast name")
	sourceID := flag.String("source", "", "Source ID")
	candidateID := flag.String("candidate", "", "Candidate ID")
	at := schedule.String("at", "", "Schedule time RFC3339")

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cfg := config.Load()
	logger := log.New(os.Stdout, "tid-factory ", log.LstdFlags)
	app, err := api.NewApp(cfg, logger)
	if err != nil {
		log.Fatalf("init: %v", err)
	}

	switch os.Args[1] {
	case "ingest":
		_ = ingest.Parse(os.Args[2:])
		if *url == "" {
			log.Fatal("--url required")
		}
		src, err := app.IngestSource(*url, *title, *podcast)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ingested: %s\n", src.ID)
	case "analyze":
		_ = analyze.Parse(os.Args[2:])
		if *sourceID == "" {
			log.Fatal("--source required")
		}
		result, err := app.AnalyzeSourceCLI(*sourceID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("analyzed: %d candidates\n", len(result))
	case "clip":
		_ = clip.Parse(os.Args[2:])
		if *candidateID == "" {
			log.Fatal("--candidate required")
		}
		if err := app.ClipCandidateCLI(*candidateID); err != nil {
			log.Fatal(err)
		}
		fmt.Println("clipped")
	case "schedule":
		_ = schedule.Parse(os.Args[2:])
		if *candidateID == "" || *at == "" {
			log.Fatal("--candidate and --at required")
		}
		when, err := time.Parse(time.RFC3339, *at)
		if err != nil {
			log.Fatal(err)
		}
		sp, err := app.ScheduleCandidateCLI(*candidateID, when)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("scheduled: %s at %s\n", sp.ID, sp.ScheduledAt)
	case "tick":
		_ = tick.Parse(os.Args[2:])
		prepared, err := app.SchedulerTickCLI()
		if err != nil {
			log.Fatal(err)
		}
		if len(prepared) == 0 {
			fmt.Println("no due posts")
		} else {
			fmt.Printf("prepared: %v\n", prepared)
		}
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`Post Factory CLI

Usage:
  factory ingest  --url URL [--title NAME] [--podcast NAME]
  factory analyze --source ID
  factory clip    --candidate ID
  factory schedule --candidate ID --at RFC3339
  factory tick    # run due scheduled posts (prepare-post flow)`)
}