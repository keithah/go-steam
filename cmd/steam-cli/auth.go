package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func handleAuth(args []string) {
	if len(args) == 0 {
		fmt.Println("Auth commands:")
		fmt.Println("  login             - Start authentication")
		fmt.Println("  code              - Submit Steam Guard code")
		fmt.Println("  logout            - End session")
		fmt.Println("  status            - Check authentication status")
		fmt.Println("  clear-rate-limit  - Clear rate limit history (for testing)")
		return
	}

	subcommand := args[0]
	subargs := args[1:]

	switch subcommand {
	case "login":
		handleAuthLogin(subargs)
	case "code":
		handleAuthCode(subargs)
	case "logout":
		handleAuthLogout(subargs)
	case "status":
		handleAuthStatus(subargs)
	case "clear-rate-limit":
		clearRateLimit()
	default:
		fmt.Printf("Unknown auth command: %s\n", subcommand)
	}
}

func handleAuthLogin(args []string) {
	fmt.Println("🔐 Steam Authentication")
	fmt.Println("======================")

	// Check if already authenticated
	if isAuthenticated() {
		fmt.Println("✅ Already authenticated!")
		fmt.Println("   Use 'steam auth logout' to start fresh")
		return
	}

	// Check for recent rate limiting
	if checkRateLimit() {
		return // Exit if rate limited
	}

	var username, password string
	
	// Check for command line arguments first
	if len(args) >= 2 {
		username = args[0]
		password = args[1]
		fmt.Printf("Using provided credentials for: %s\n", username)
	} else {
		// Interactive mode
		reader := bufio.NewReader(os.Stdin)
		
		fmt.Print("Steam username: ")
		input, _ := reader.ReadString('\n')
		username = strings.TrimSpace(input)
		
		fmt.Print("Steam password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			return
		}
		password = string(bytePassword)
		fmt.Println() // New line after hidden input
	}

	if username == "" || password == "" {
		fmt.Println("❌ Username and password are required")
		fmt.Println("Usage: steam auth login [username] [password]")
		return
	}

	// Start authentication session
	fmt.Println("\n⏳ Starting authentication session...")
	err := startAuthSession(username, password)
	if err != nil {
		fmt.Printf("❌ Authentication failed: %v\n", err)
		return
	}

	fmt.Println("✅ Authentication session started")
	fmt.Println("📧 If Steam Guard is required, you'll receive an email")
	fmt.Println("   Use: steam auth code <CODE>")
}

func handleAuthCode(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: steam auth code <STEAM_GUARD_CODE>")
		fmt.Println("Example: steam auth code 4D6XG")
		return
	}

	code := args[0]
	fmt.Printf("🔐 Submitting Steam Guard code: %s\n", code)

	err := submitSteamGuardCode(code)
	if err != nil {
		fmt.Printf("❌ Failed to submit code: %v\n", err)
		return
	}

	fmt.Println("✅ Code submitted successfully")
	fmt.Println("   Use 'steam auth status' to check authentication")
}

func handleAuthLogout(args []string) {
	fmt.Println("👋 Logging out...")
	err := endAuthSession()
	if err != nil {
		fmt.Printf("❌ Logout failed: %v\n", err)
		return
	}
	fmt.Println("✅ Logged out successfully")
}

func handleAuthStatus(args []string) {
	fmt.Println("📊 Authentication Status")
	fmt.Println("=======================")

	status := getAuthStatus()
	fmt.Printf("Connected: %v\n", status.Connected)
	fmt.Printf("Authenticated: %v\n", status.Authenticated)
	fmt.Printf("Needs Code: %v\n", status.NeedsCode)
	
	if status.Authenticated {
		fmt.Printf("Steam ID: %d\n", status.SteamID)
		fmt.Printf("Username: %s\n", status.Username)
	}

	if status.LastError != "" {
		fmt.Printf("Last Error: %s\n", status.LastError)
	}

	if status.Connected && !status.Authenticated {
		if status.NeedsCode {
			fmt.Println("\n💡 Steam Guard code required")
			fmt.Println("   Use: steam auth code <CODE>")
		} else {
			fmt.Println("\n⏳ Authentication in progress...")
		}
	}
}