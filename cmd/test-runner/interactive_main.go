package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	testing "github.com/Philipp15b/go-steam/v3/testing"
)

func main() {
	var (
		username = flag.String("username", "", "Steam username")
		password = flag.String("password", "", "Steam password")
		verbose  = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Println("Interactive Steam Test Session")
		fmt.Println("=============================")
		fmt.Println()
		fmt.Println("This creates a persistent Steam connection that you can interact with.")
		fmt.Println("Perfect for testing the Steam Guard flow and messaging!")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  --username <steam_username>")
		fmt.Println("  --password <steam_password>")
		fmt.Println("  --verbose (detailed output)")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  ./steam-interactive --username myuser --password mypass")
		fmt.Println()
		fmt.Println("The session will:")
		fmt.Println("  1. Connect to Steam")
		fmt.Println("  2. Attempt authentication") 
		fmt.Println("  3. Ask for Steam Guard code if needed")
		fmt.Println("  4. Provide interactive commands for testing")
		fmt.Println("  5. Log everything to steam-test.log")
		os.Exit(1)
	}

	// Create interactive session
	session, err := testing.NewInteractiveTestSession(*username, *password)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n🛑 Shutting down gracefully...")
		session.Stop()
		os.Exit(0)
	}()

	if *verbose {
		fmt.Printf("Starting interactive session for: %s\n", *username)
		fmt.Println()
	}

	// Start the session
	session.Start()
	
	// Keep running until user quits
	select {}
}