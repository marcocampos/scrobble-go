// Command webauth demonstrates the Last.fm web authentication flow.
//
// This is the recommended flow for desktop and web applications where you
// don't want to handle the user's password directly.
//
// Usage:
//
//	export LASTFM_API_KEY=...
//	export LASTFM_API_SECRET=...
//	go run ./examples/webauth
//
// The program will print an authorisation URL. Open it in a browser, approve
// the request, then press Enter. The session key is printed to stdout — store
// it and pass it to WithSessionKey on subsequent runs.
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/marcocampos/scrobble-go/lastfm"
)

func main() {
	apiKey := mustEnv("LASTFM_API_KEY")
	apiSecret := mustEnv("LASTFM_API_SECRET")

	client := lastfm.NewLastFMClient(apiKey, apiSecret)
	gen := lastfm.NewSessionKeyGenerator(client)

	ctx := context.Background()

	// Step 1: get a one-time authorisation URL.
	fmt.Println("Requesting authorisation URL from Last.fm...")
	authURL, err := gen.GetWebAuthURL(ctx)
	if err != nil {
		log.Fatalf("GetWebAuthURL: %v", err)
	}

	fmt.Printf("\nOpen this URL in your browser and click \"Allow\":\n\n  %s\n\n", authURL)
	fmt.Print("Press Enter once you have approved the request...")
	if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
		log.Fatalf("reading stdin: %v", err)
	}

	// Step 2: poll until the user approves (or until we time out).
	fmt.Println("Retrieving session key...")
	var sessionKey, username string
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		sessionKey, username, err = gen.GetWebAuthSessionKeyAndUsername(ctx, authURL, "")
		if err == nil {
			break
		}

		var wsErr *lastfm.WSError
		if errors.As(err, &wsErr) && wsErr.Status == "14" {
			// StatusTokenUnauthorized — user hasn't approved yet; keep polling.
			fmt.Print(".")
			time.Sleep(2 * time.Second)
			continue
		}
		log.Fatalf("GetWebAuthSessionKeyAndUsername: %v", err)
	}
	if sessionKey == "" {
		log.Fatal("timed out waiting for user to approve the request")
	}

	fmt.Printf("\n\nAuthenticated as: %s\n", username)
	fmt.Printf("Session key:      %s\n\n", sessionKey)
	fmt.Println("Store this session key and pass it to WithSessionKey on future runs:")
	fmt.Printf("  client := lastfm.NewLastFMClient(apiKey, apiSecret,\n")
	fmt.Printf("      lastfm.WithSessionKey(%q),\n", sessionKey)
	fmt.Printf("  )\n")
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}
