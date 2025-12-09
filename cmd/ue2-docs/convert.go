package main

import (
	"flag"
	"fmt"
	"os"
)

func runConvert(args []string) {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)

	inputDir := fs.String("input", "./output", "Input directory containing scraped HTML")
	outputDir := fs.String("output", "./markdown", "Output directory for markdown files")
	preserveStructure := fs.Bool("preserve-structure", true, "Keep original directory structure")

	fs.Usage = func() {
		fmt.Println("Usage: ue2-docs convert [flags]")
		fmt.Println()
		fmt.Println("Convert scraped HTML documentation to Markdown.")
		fmt.Println()
		fmt.Println("Flags:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  ue2-docs convert --input ./scraped --output ./docs")
	}

	fs.Parse(args)

	fmt.Println("UE2 Docs - Convert to Markdown")
	fmt.Println("===============================")
	fmt.Println()
	fmt.Printf("Input Dir:           %s\n", *inputDir)
	fmt.Printf("Output Dir:          %s\n", *outputDir)
	fmt.Printf("Preserve Structure:  %t\n", *preserveStructure)
	fmt.Println()

	// TODO: Initialize and start converter
	fmt.Println("Converter not yet implemented. See plan.md for implementation roadmap.")
	os.Exit(0)
}
