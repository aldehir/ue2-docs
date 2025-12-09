package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "scrape":
		runScrape(os.Args[2:])
	case "convert":
		runConvert(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("UE2 Docs - Scrape and convert UE2 documentation")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ue2-docs <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  scrape    Scrape documentation from a website")
	fmt.Println("  convert   Convert scraped HTML to Markdown")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("Run 'ue2-docs <command> --help' for command-specific options.")
}
