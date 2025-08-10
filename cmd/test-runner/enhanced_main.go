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
		username   = flag.String("username", "", "Steam username")
		password   = flag.String("password", "", "Steam password")
		authCode   = flag.String("authcode", "", "Steam Guard email code")
		enhanced   = flag.Bool("enhanced", false, "Run enhanced diagnostics")
		verbose    = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	// Get credentials - simplified for testing
	if *username == "" || *password == "" {
		fmt.Println("Steam Protocol Test Runner")
		fmt.Println("=========================")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  --username <steam_username>")
		fmt.Println("  --password <steam_password>")
		fmt.Println("  --authcode <steam_guard_code> (if required)")
		fmt.Println("  --enhanced (run enhanced diagnostics)")
		fmt.Println("  --verbose (detailed output)")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  ./steam-test-enhanced --username myuser --password mypass --enhanced")
		fmt.Println("  ./steam-test-enhanced --username myuser --password mypass --authcode ABC123 --enhanced")
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Testing Steam account: %s\n", *username)
		fmt.Println()
	}

	if *enhanced {
		// Run enhanced diagnostics
		suite := testing.NewEnhancedTestSuite(*username, *password, *authCode)
		suite.RunEnhancedTests()
		suite.Disconnect()
	} else {
		// Run standard tests
		suite := testing.NewTestSuite(*username, *password)
		
		if err := suite.RunAllTests(); err != nil {
			log.Fatalf("Test suite failed: %v", err)
		}

		suite.PrintResults()
	}
}