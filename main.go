package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/proxy"
)

func main() {

	// Create output directory if it doesn't exist
	outputDir := "scans"
	reportFile, err := os.OpenFile("scan_report.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open report file: %v", err)
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	// Check for command-line argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go targets.yaml")
		os.Exit(1)
	}

	// Read targets from the specified file
	fileName := os.Args[1]

	targets, err := readLines(fileName)
	if err != nil {
		log.Fatalf("Failed to read targets file: %v", err)
	}
	fmt.Printf("Targets:\n%v\n", targets)

	// Tor proxy settings
	torProxy := "127.0.0.1:9150"

	var fileMutex sync.Mutex

	// Create a SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", torProxy, nil, proxy.Direct)
	if err != nil {
		log.Fatal("Proxy dialer error", err)
	}

	// Create an HTTP client that uses the SOCKS5 dialer
	tr := &http.Transport{Dial: dialer.Dial}
	client := &http.Client{Transport: tr, Timeout: time.Second * 20}

	var wg sync.WaitGroup

	// Make a request to a .onion addresses an example
	for i, url := range targets {

		wg.Add(1)

		go func(targetIndex int, targetUrl string) {

			now := time.Now().Format("2006-01-02 15:04:05")

			defer wg.Done()

			fmt.Printf("[%d-%d]Scanning %s\n", targetIndex+1, len(targets), targetUrl)
			resp, err := client.Get(targetUrl)
			if err != nil {
				fmt.Println("Request error: ", err)
				return
			}
			defer resp.Body.Close()

			// Read the response body
			// Previously we used ioutil.ReadAll, but it's deprecated
			bodyBytes, err := io.ReadAll(resp.Body)

			if err != nil {
				fmt.Println("Error reading response body: ", err)
				resp.Body.Close()
				return
			}
			fmt.Printf("Success (Size: %d byte) -> ", len(bodyBytes))

			htmlString := string(bodyBytes)
			pageTitle := extractTitle(htmlString)
			safeTitle := sanitizeFilename(pageTitle)

			timestamp := time.Now().Format("20060102_150405")
			folderName := fmt.Sprintf("%s/%s_%s", outputDir, safeTitle, timestamp)

			err = os.Mkdir(folderName, 0755)
			if err != nil {
				fmt.Println("Error creating folder: ", err)
				resp.Body.Close()
				return
			}

			fmt.Printf("Screenshotting...%s\n", targetUrl)
			ssData, ssErr := takeScreenshot(targetUrl, torProxy)

			if ssErr != nil {
				fmt.Println("Screenshot error: ", ssErr)
				fileMutex.Lock()
				reportFile.WriteString(fmt.Sprintf("[%s] WARNING  %s  Screenshot Failed: %v\n", now, targetUrl, ssErr))
				fileMutex.Unlock()
				return
			} else {
				ssFileName := fmt.Sprintf("%s/screenshot.png", folderName)
				err = os.WriteFile(ssFileName, ssData, 0644)
				fmt.Printf("Screenshot saved: %s\n", ssFileName)
			}

			htmlFileName := fmt.Sprintf("%s/index.html", folderName)
			err = os.WriteFile(htmlFileName, bodyBytes, 0644)

			fileMutex.Lock()
			// Log success to report file
			reportFile.WriteString(fmt.Sprintf("[%s] SUCCESS  %s  Title: %s  Size: %d\n", now, targetUrl, pageTitle, len(bodyBytes)))
			fileMutex.Unlock()

			fmt.Printf("Response Status for %s: %s\n", url, resp.Status)
			resp.Body.Close()

		}(i, url)
	}
	fmt.Println("Is waited all scans...")
	wg.Wait()
	fmt.Println("All scans completed.")
}

func takeScreenshot(url string, proxy string) ([]byte, error) {

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("proxy-server", "socks5://"+proxy),
		chromedp.Flag("ignore-certificate-errors", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := context.WithTimeout(allocCtx, 45*time.Second)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var buf []byte

	// Run chromedp tasks
	err := chromedp.Run(ctx, chromedp.Navigate(url), chromedp.Sleep(2*time.Second), chromedp.CaptureScreenshot(&buf))

	if err != nil {
		return nil, err
	}

	return buf, nil
}

func sanitizeFilename(name string) string {
	// Replace any characters that are not allowed in filenames
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Limit length to 50 characters
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

func extractTitle(htmlContent string) string {
	// Simple extraction of the <title> tag content
	startTag := "<title>"
	endTag := "</title>"

	// Case insensitive search
	lowerContent := strings.ToLower(htmlContent)
	startIndex := strings.Index(lowerContent, startTag)
	if startIndex == -1 {
		return "No Title Found"
	}

	// Move index to the end of the start tag
	startIndex += len(startTag)
	endIndex := strings.Index(lowerContent[startIndex:], endTag)
	if endIndex == -1 {
		return "No Title Found"
	}

	// Extract and return the title
	title := htmlContent[startIndex : startIndex+endIndex]
	return strings.TrimSpace(title)

}

// readLines reads a file and returns a slice of its lines
func readLines(path string) ([]string, error) {

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read lines from the file
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}
