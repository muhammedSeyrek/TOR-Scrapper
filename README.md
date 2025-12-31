
# TOR Scraper

A fast, concurrent tool written in Go to scrape and take screenshots of `.onion` websites via the Tor network.

I built this project to automate the analysis of Tor hidden services. Instead of checking links manually one by one, this tool spins up multiple goroutines to handle the workload efficiently.

## Features

* **Tor Proxy Integration:** Routes all traffic through a local SOCKS5 proxy (default `127.0.0.1:9150`) to ensure anonymity.
* **Headless Screenshots:** Uses `chromedp` to capture full screenshots of target sites invisibly.
* **Data Collection:** Downloads and saves the raw HTML source code.
* **Concurrency:** Scans multiple targets simultaneously for speed.
* **Smart Logging:** Generates a detailed `scan_report.log` and handles timeouts/errors gracefully without crashing.

## Prerequisites

Before running the tool, ensure you have the following:

1.  **Go (Golang):** Installed on your machine.
2.  **Tor Service:** You must have the Tor Browser or Tor service running in the background.
    * *Note: The code defaults to port `9150` (Tor Browser default). If you are using the system Tor service, you might need to change it to `9050` in `main.go`.*
3.  **Google Chrome / Chromium:** Required for the screenshot functionality.

## Installation

```bash
# Clone the repository
git clone [https://github.com/muhammedSeyrek/TOR-Scrapper.git](https://github.com/muhammedSeyrek/TOR-Scrapper.git)

# Navigate to the directory
cd TOR-Scrapper

# Install dependencies
go mod tidy
Usage
Prepare your targets: Create a file named targets.yaml in the root directory and add the onion URLs you want to scan (one per line).

YAML

[http://exampleonionaddress.onion](http://exampleonionaddress.onion)
[http://anotheraddress.onion](http://anotheraddress.onion)
Run the tool:

Bash

go run main.go targets.yaml
Or build it as an exe: go build -o tor_scraper.exe main.go

Output Structure
The tool creates a scans directory. Each URL gets its own folder named with the page title and a timestamp.

Plaintext

TOR-Scrapper/
├── scan_report.log       # Summary of success/fail status
├── scans/
│   ├── PageTitle_Timestamp/
│   │   ├── index.html        # Source code
│   │   └── screenshot.png    # Screenshot of the page
│   └── ...
├── targets.yaml
└── main.go
Screenshots & Demo
Here is the tool in action:

1. Terminal Output
Running the tool shows real-time logs of the concurrent scanning process.

2. File Organization
Every target gets a dedicated folder containing the HTML source and a screenshot.

3. Captured Evidence
The tool successfully captures screenshots of live onion sites using the Tor network.

4. Reporting
A clean log file (scan_report.txt) summarizes the status, title, and size of each scan.

Disclaimer
This tool is for educational and research purposes only. I am not responsible for how you use this tool or the content you access on the Tor network. Always obey your local laws.