package main

import (
	"flag"
	"fmt"
	"os"
)

func runScrape(args []string) {
	fs := flag.NewFlagSet("scrape", flag.ExitOnError)

	rootURL := fs.String("root-url", "https://docs.unrealengine.com/udk/Two/SiteMap.html", "Starting URL to scrape")
	outputDir := fs.String("output", "./output", "Output directory for scraped content")
	workers := fs.Int("workers", 10, "Number of concurrent workers")
	whitelist := fs.String("whitelist", "", "Comma-separated list of additional domains to allow")
	maxDepth := fs.Int("max-depth", 0, "Maximum link depth (0 = unlimited)")

	fs.Usage = func() {
		fmt.Println("Usage: ue2-docs scrape [flags]")
		fmt.Println()
		fmt.Println("Scrape documentation from a website and save locally with rewritten paths.")
		fmt.Println()
		fmt.Println("Flags:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  ue2-docs scrape --root-url https://docs.unrealengine.com/udk/Two/SiteMap.html --output ./scraped")
	}

	fs.Parse(args)

	fmt.Println("UE2 Docs - Scrape")
	fmt.Println("=================")
	fmt.Println()
	fmt.Printf("Root URL:     %s\n", *rootURL)
	fmt.Printf("Output Dir:   %s\n", *outputDir)
	fmt.Printf("Workers:      %d\n", *workers)
	if *whitelist != "" {
		fmt.Printf("Whitelist:    %s\n", *whitelist)
	}
	if *maxDepth > 0 {
		fmt.Printf("Max Depth:    %d\n", *maxDepth)
	}
	fmt.Println()

	// TODO: Initialize and start scraper
	fmt.Println("Scraper not yet implemented. See plan.md for implementation roadmap.")
	os.Exit(0)
}
