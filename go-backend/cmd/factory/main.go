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
	trim := flag.NewFlagSet("trim", flag.ExitOnError)
	schedule := flag.NewFlagSet("schedule", flag.ExitOnError)
	tick := flag.NewFlagSet("tick", flag.ExitOnError)
	postNow := flag.NewFlagSet("post-now", flag.ExitOnError)

	url := ingest.String("url", "", "YouTube URL")
	podcast := ingest.String("podcast", "", "Podcast name")
	sourceID := analyze.String("source", "", "Source ID")
	clipCandidateID := clip.String("candidate", "", "Candidate ID")
	trimCandidateID := trim.String("candidate", "", "Candidate ID")
	startTime := trim.String("start", "", "New start time HH:MM:SS")
	endTime := trim.String("end", "", "New end time HH:MM:SS")
	scheduleCandidateID := schedule.String("candidate", "", "Candidate ID")
	at := schedule.String("at", "", "Schedule time RFC3339")
	postNowCandidateID := postNow.String("candidate", "", "Candidate ID")

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
		if *podcast == "" {
			log.Fatal("--podcast required")
		}
		src, err := app.IngestSource(*url, *podcast)
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
		if *clipCandidateID == "" {
			log.Fatal("--candidate required")
		}
		if err := app.ClipCandidateCLI(*clipCandidateID); err != nil {
			log.Fatal(err)
		}
		fmt.Println("clipped")
	case "trim":
		_ = trim.Parse(os.Args[2:])
		if *trimCandidateID == "" {
			log.Fatal("--candidate required")
		}
		if *startTime == "" && *endTime == "" {
			log.Fatal("--start and/or --end required")
		}
		c, err := app.TrimCandidateCLI(*trimCandidateID, *startTime, *endTime)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("trimmed: %s (%s → %s)\n", c.ID, c.StartTime, c.EndTime)
	case "schedule":
		_ = schedule.Parse(os.Args[2:])
		if *scheduleCandidateID == "" || *at == "" {
			log.Fatal("--candidate and --at required")
		}
		when, err := time.Parse(time.RFC3339, *at)
		if err != nil {
			log.Fatal(err)
		}
		sp, err := app.ScheduleCandidateCLI(*scheduleCandidateID, when)
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
	case "post-now":
		_ = postNow.Parse(os.Args[2:])
		if *postNowCandidateID == "" {
			log.Fatal("--candidate required")
		}
		c, err := app.PostNowCLI(*postNowCandidateID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("prepared: %s (clip: %s)\n", c.ID, c.ClipPath)
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`Post Factory CLI

Usage:
  factory ingest  --url URL --podcast NAME
  factory analyze --source ID
  factory clip    --candidate ID
  factory trim    --candidate ID [--start HH:MM:SS] [--end HH:MM:SS]
  factory schedule --candidate ID --at RFC3339
  factory post-now --candidate ID
  factory tick    # run due scheduled posts (prepare-post flow)`)
}