package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// CLI flags
	rootURL := flag.String("root-url", "https://docs.unrealengine.com/udk/Two/SiteMap.html", "Starting URL to scrape")
	outputDir := flag.String("output", "./output", "Output directory for scraped content")
	workers := flag.Int("workers", 10, "Number of concurrent workers")
	whitelist := flag.String("whitelist", "", "Comma-separated list of additional domains to allow")

	flag.Parse()

	fmt.Printf("UE2 Docs Scraper\n")
	fmt.Printf("================\n\n")
	fmt.Printf("Root URL:     %s\n", *rootURL)
	fmt.Printf("Output Dir:   %s\n", *outputDir)
	fmt.Printf("Workers:      %d\n", *workers)
	if *whitelist != "" {
		fmt.Printf("Whitelist:    %s\n", *whitelist)
	}
	fmt.Println()

	// TODO: Initialize and start scraper
	fmt.Println("Scraper not yet implemented. See plan.md for implementation roadmap.")
	os.Exit(0)
}
