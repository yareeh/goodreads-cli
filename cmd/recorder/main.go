package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type RequestLog struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	PostData    string            `json:"post_data,omitempty"`
	RequestType string            `json:"type"`
}

func main() {
	outputFile := flag.String("o", "", "output file for request logs (default: stdout)")
	flag.Parse()

	var out *os.File
	if *outputFile != "" {
		var err error
		out, err = os.Create(*outputFile)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")

	// Launch a visible browser
	u := launcher.New().
		Headless(false).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("")

	// Enable network interception via CDP
	_ = proto.NetworkEnable{}.Call(page)

	go page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
		// Filter to goodreads.com domains only
		if !strings.Contains(e.Request.URL, "goodreads.com") {
			return
		}

		headers := make(map[string]string)
		for k, v := range e.Request.Headers {
			headers[k] = v.Str()
		}

		entry := RequestLog{
			Method:      e.Request.Method,
			URL:         e.Request.URL,
			Headers:     headers,
			PostData:    e.Request.PostData,
			RequestType: string(e.Type),
		}

		if err := encoder.Encode(entry); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding log: %v\n", err)
		}
	})()

	// Navigate to Goodreads
	page.MustNavigate("https://www.goodreads.com")

	fmt.Fprintln(os.Stderr, "=== Goodreads Request Recorder ===")
	fmt.Fprintln(os.Stderr, "Browser is open. Interact with Goodreads normally.")
	fmt.Fprintln(os.Stderr, "All requests to goodreads.com are being logged.")
	fmt.Fprintln(os.Stderr, "Press Ctrl+C to stop.")

	// Wait for Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Fprintln(os.Stderr, "\nRecorder stopped.")
}
