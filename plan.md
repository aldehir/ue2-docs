# UE2 Docs Scraper - Implementation Plan

## Project Overview

A Go-based web scraper to download and archive the UE2 (Unreal Engine 2) documentation from https://docs.unrealengine.com/udk/Two/SiteMap.html.

**Module**: `github.com/aldehir/ue2-docs`

## Requirements

1. ✅ Written in Go
2. ✅ Parallel scraping with worker queue
3. ✅ Scrape HTML, CSS, JS, and image assets
4. ✅ Stay within root URL unless whitelisted
5. ✅ Handle cycles gracefully (visited URL tracking)
6. ✅ Modify HTML paths to be relative

## Architecture

### Directory Structure

```
ue2-docs/
├── cmd/
│   └── ue2-docs/          # Main CLI application
│       ├── main.go        # Entry point with subcommand routing
│       ├── scrape.go      # 'scrape' subcommand
│       └── convert.go     # 'convert' subcommand
├── internal/
│   ├── scraper/           # Core scraping logic
│   │   ├── scraper.go     # Main scraper orchestrator
│   │   ├── worker.go      # Worker pool implementation
│   │   └── queue.go       # URL queue management
│   ├── parser/            # HTML/CSS parsing & rewriting
│   │   ├── html.go        # HTML parser and path rewriter
│   │   ├── css.go         # CSS parser and URL rewriter
│   │   └── paths.go       # Path resolution utilities
│   ├── converter/         # HTML to Markdown conversion
│   │   ├── converter.go   # Main conversion logic
│   │   └── elements.go    # Element-specific converters
│   ├── fetcher/           # HTTP fetching logic
│   │   └── fetcher.go     # HTTP client with retry/timeout
│   ├── storage/           # File system operations
│   │   └── storage.go     # Save files with proper structure
│   └── urlutil/           # URL utilities
│       ├── filter.go      # URL filtering and validation
│       └── normalize.go   # URL normalization
├── pkg/                   # Public packages (if needed)
├── go.mod
├── go.sum
├── README.md
└── plan.md               # This file
```

## Core Components

### 1. Main Entry Point (`cmd/ue2-docs/main.go`)
- Route to subcommands: `scrape` or `convert`
- Handle global flags and help
- Graceful shutdown handling

### 1a. Scrape Command (`cmd/ue2-docs/scrape.go`)
- Parse scrape-specific flags (root URL, output, workers, whitelist)
- Initialize and run scraper
- Save HTML, CSS, JS, and assets with rewritten paths

### 1b. Convert Command (`cmd/ue2-docs/convert.go`)
- Parse convert-specific flags (input directory, output directory)
- Walk HTML files in input directory
- Convert each to markdown and save to output directory

### 2. Scraper Orchestrator (`internal/scraper/scraper.go`)
- Coordinate worker pool
- Manage URL queue
- Track visited URLs (cycle detection)
- Maintain domain whitelist
- Progress reporting

### 3. Worker Pool (`internal/scraper/worker.go`)
- Spawn N concurrent workers
- Pull URLs from queue
- Fetch and process resources
- Discover new URLs
- Handle errors and retries

### 4. URL Queue (`internal/scraper/queue.go`)
- Thread-safe queue implementation
- Priority handling (HTML > CSS > JS > Images)
- Deduplication

### 5. HTML Parser (`internal/parser/html.go`)
- Parse HTML using `golang.org/x/net/html`
- Extract links, scripts, stylesheets, images
- Rewrite paths to relative
- Preserve document structure

### 6. CSS Parser (`internal/parser/css.go`)
- Parse CSS files
- Extract `url()` references
- Rewrite paths to relative
- Handle `@import` statements

### 7. Fetcher (`internal/fetcher/fetcher.go`)
- HTTP client with timeout
- User-Agent header
- Retry logic with exponential backoff
- Respect robots.txt (optional)

### 8. Storage (`internal/storage/storage.go`)
- Create directory structure
- Save files with proper extensions
- Handle filename conflicts
- Preserve file metadata

### 9. URL Utilities (`internal/urlutil/`)
- Normalize URLs (remove fragments, resolve relative paths)
- Filter URLs (whitelist check, same-origin policy)
- Detect resource types by extension/content-type

### 10. Markdown Converter (`internal/converter/`)
- Convert HTML to Markdown using `golang.org/x/net/html`
- Element-specific conversion logic (headings, links, images, code blocks)
- Handle UE2-specific formatting
- Preserve code examples and special content
- Generate clean, readable markdown output

## Implementation Phases

### Phase 1: Project Setup ✓
- [x] Initialize Go module
- [x] Create directory structure
- [x] Set up basic main.go

### Phase 2: Core Infrastructure ✓
- [x] Implement URL queue (thread-safe)
- [x] Implement visited URL tracker
- [x] Create URL normalization utilities
- [x] Implement URL filtering/whitelist logic

### Phase 3: HTTP Fetching
- [ ] Create HTTP fetcher with timeout
- [ ] Add retry logic
- [ ] Implement content-type detection
- [ ] Add rate limiting (optional)

### Phase 4: Storage Layer
- [ ] Implement file storage with directory structure
- [ ] Create path mapping (URL -> filesystem)
- [ ] Handle filename sanitization

### Phase 5: HTML Processing
- [ ] Parse HTML documents
- [ ] Extract all resource references (links, images, scripts, styles)
- [ ] Rewrite paths to relative
- [ ] Queue discovered URLs

### Phase 6: Asset Processing
- [ ] Handle CSS files (parse and rewrite url())
- [ ] Handle JavaScript files (download as-is)
- [ ] Handle images (download binary)
- [ ] Handle other assets (fonts, etc.)

### Phase 7: Worker Pool
- [ ] Implement worker pool
- [ ] Add work distribution
- [ ] Implement graceful shutdown
- [ ] Add progress tracking

### Phase 8: Main Orchestrator
- [ ] Wire all components together
- [ ] Add CLI flags and configuration
- [ ] Implement main scraping loop
- [ ] Add logging and error handling

### Phase 9: Testing & Refinement
- [ ] Test with sample pages
- [ ] Test full UE2 docs scrape
- [ ] Fix path rewriting issues
- [ ] Optimize performance
- [ ] Add unit tests

### Phase 10: Markdown Conversion
- [ ] Implement HTML node walker
- [ ] Create element-to-markdown converters (h1-h6, p, a, img, code, pre, ul, ol, table)
- [ ] Handle nested elements and text formatting (bold, italic, code)
- [ ] Convert scraped HTML files to markdown
- [ ] Preserve code blocks and UE2-specific content
- [ ] Generate index/navigation for markdown docs
- [ ] Validate markdown output

## Technical Details

### Cycle Detection
- Use `sync.Map` or map with mutex to track visited URLs
- Store normalized URLs to handle duplicates
- Check before adding to queue

### Parallel Processing
- Use worker pool pattern with channels
- Worker count configurable (default: 10)
- Use `sync.WaitGroup` for coordination
- Buffered channels for queue

### URL Whitelisting
- Root domain: `docs.unrealengine.com`
- Allowed paths: `/udk/Two/*`
- Whitelist for external resources (e.g., CDN domains)
- Reject external links by default

### Path Rewriting Strategy
1. Parse base URL from document
2. Resolve all relative URLs to absolute
3. Determine if URL should be scraped
4. Calculate relative path from current document to target
5. Rewrite reference in HTML/CSS

### Resource Type Detection
- By extension: `.html`, `.css`, `.js`, `.png`, `.jpg`, `.gif`, etc.
- By Content-Type header
- Default to binary for unknown types

### Error Handling
- Retry HTTP errors (3 attempts)
- Log and skip broken links
- Continue on parse errors
- Graceful shutdown on interrupt

## Dependencies

```go
require (
    golang.org/x/net v0.x.x  // HTML parsing and manipulation
    // Potentially:
    // - github.com/tdewolff/parse/v2 for CSS parsing (if needed)
)
```

**Why golang.org/x/net/html?**
- Official Go extended library (15,509+ packages use it)
- HTML5-compliant parser with tokenizer and node tree APIs
- Perfect for both scraping AND markdown conversion
- Fine-grained control for custom HTML-to-Markdown logic
- Minimal dependencies, maximum flexibility
- Single library for entire pipeline: scrape → rewrite → convert

## CLI Commands

### `ue2-docs scrape`
Scrape documentation from a website and save locally with rewritten paths.

**Flags:**
- `--root-url`: Starting URL (default: https://docs.unrealengine.com/udk/Two/SiteMap.html)
- `--output`: Output directory for scraped HTML (default: ./output)
- `--workers`: Number of concurrent workers (default: 10)
- `--whitelist`: Additional domains to allow (comma-separated)
- `--max-depth`: Maximum link depth (optional)

**Example:**
```bash
ue2-docs scrape --root-url https://docs.unrealengine.com/udk/Two/SiteMap.html --output ./scraped
```

### `ue2-docs convert`
Convert scraped HTML documentation to Markdown.

**Flags:**
- `--input`: Input directory containing scraped HTML (default: ./output)
- `--output`: Output directory for markdown files (default: ./markdown)
- `--preserve-structure`: Keep original directory structure (default: true)

**Example:**
```bash
ue2-docs convert --input ./scraped --output ./docs
```

## Success Criteria

### Phase 1-9: HTML Scraping
1. Complete scrape of UE2 documentation
2. All HTML pages render correctly offline
3. All CSS styles applied correctly
4. All images display correctly
5. No broken internal links
6. Reasonable performance (minutes, not hours)

### Phase 10: Markdown Conversion
7. Clean, readable markdown files generated from HTML
8. All links work correctly in markdown (relative paths preserved)
9. Images embedded properly with correct paths
10. Code blocks and formatting preserved
11. Documentation navigable as markdown (e.g., via GitHub, static site generators)
12. Markdown files suitable for version control and collaboration
