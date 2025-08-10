package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	testing "github.com/Philipp15b/go-steam/v3/testing"
)

func main() {
	var (
		username = flag.String("username", "", "Steam username")
		password = flag.String("password", "", "Steam password (will prompt if not provided)")
		verbose  = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	// Get credentials - simplified for testing
	if *username == "" || *password == "" {
		fmt.Println("Please provide credentials with flags:")
		fmt.Println("  --username <steam_username>")
		fmt.Println("  --password <steam_password>")
		fmt.Println("Example: ./steam-test --username myuser --password mypass --verbose")
		os.Exit(1)
	}

	if *username == "" || *password == "" {
		log.Fatal("Username and password are required")
	}

	// Create and run test suite
	suite := testing.NewTestSuite(*username, *password)
	
	if *verbose {
		fmt.Printf("Running tests for user: %s\n\n", *username)
	}

	if err := suite.RunAllTests(); err != nil {
		log.Fatalf("Test suite failed: %v", err)
	}

	suite.PrintResults()
}