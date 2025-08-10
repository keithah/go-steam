package testing

import (
	"context"
	"fmt"
	"sync"
	"time"

	steam "github.com/Philipp15b/go-steam/v3"
	"github.com/Philipp15b/go-steam/v3/protocol/steamlang"
)

// EnhancedTestSuite with better Steam Guard handling and diagnostics
type EnhancedTestSuite struct {
	client   *steam.Client
	results  []TestResult
	mutex    sync.Mutex
	username string
	password string
	authCode string
}

// NewEnhancedTestSuite creates a new enhanced test suite
func NewEnhancedTestSuite(username, password, authCode string) *EnhancedTestSuite {
	return &EnhancedTestSuite{
		username: username,
		password: password,
		authCode: authCode,
		results:  make([]TestResult, 0),
	}
}

// DiagnosticConnectAndAuth with detailed Steam Guard and authentication analysis
func (ts *EnhancedTestSuite) DiagnosticConnectAndAuth() error {
	fmt.Println("🔍 Starting enhanced Steam authentication diagnostics...")
	
	ts.client = steam.NewClient()
	
	// Test connection first
	fmt.Println("⏳ Connecting to Steam servers...")
	ts.client.Connect()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	connected := make(chan bool)
	authResult := make(chan error)
	authEvents := make(chan string, 10) // Buffer for auth event details

	go func() {
		for event := range ts.client.Events() {
			switch e := event.(type) {
			case *steam.ConnectedEvent:
				fmt.Println("✅ Successfully connected to Steam network")
				authEvents <- "Connected to Steam servers"
				connected <- true
				
				// Attempt login
				logOnDetails := &steam.LogOnDetails{
					Username: ts.username,
					Password: ts.password,
				}
				
				if ts.authCode != "" {
					fmt.Printf("🔐 Attempting authentication with Steam Guard code: %s...\n", ts.authCode)
					logOnDetails.AuthCode = ts.authCode
				} else {
					fmt.Println("🔐 Attempting authentication...")
				}
				
				ts.client.Auth.LogOn(logOnDetails)
			
			case *steam.LoggedOnEvent:
				fmt.Println("✅ Authentication successful!")
				authEvents <- fmt.Sprintf("Logged in successfully as %s", ts.username)
				authResult <- nil
				return

			case *steam.LogOnFailedEvent:
				fmt.Printf("❌ Authentication failed: %v\n", e.Result)
				
				// Detailed error analysis
				errorMsg := ts.analyzeAuthError(e.Result)
				authEvents <- errorMsg
				authResult <- fmt.Errorf("authentication failed: %v - %s", e.Result, errorMsg)
				return

			case *steam.DisconnectedEvent:
				fmt.Println("❌ Disconnected from Steam during authentication")
				authEvents <- "Disconnected during authentication"
				authResult <- fmt.Errorf("disconnected during auth")
				return

			case *steam.MachineAuthUpdateEvent:
				fmt.Println("🔐 Steam Guard machine authentication event received")
				authEvents <- "Steam Guard machine auth update"
				
			case *steam.AccountInfoEvent:
				fmt.Printf("ℹ️  Account info received: %+v\n", e)
				authEvents <- "Account info received"
				
			default:
				// Log all events for diagnostic purposes
				fmt.Printf("🔍 Event received: %T\n", e)
			}
		}
	}()

	// Wait for connection
	select {
	case <-connected:
		fmt.Println("⏳ Waiting for authentication response...")
		// Wait for auth result
		select {
		case err := <-authResult:
			ts.printAuthDiagnostics(authEvents)
			return err
		case <-ctx.Done():
			ts.printAuthDiagnostics(authEvents)
			return fmt.Errorf("authentication timeout - this may indicate Steam Guard is required or servers are slow")
		}
	case <-ctx.Done():
		ts.printAuthDiagnostics(authEvents)
		return fmt.Errorf("connection timeout - Steam servers may be unreachable")
	}
}

func (ts *EnhancedTestSuite) analyzeAuthError(result steamlang.EResult) string {
	switch result {
	case steamlang.EResult_AccountLogonDenied:
		return "Account login denied - This usually means Steam Guard email verification is required. Check your email and approve the login, or try logging into the Steam client first to authorize this device."
		
	case steamlang.EResult_InvalidPassword:
		return "Invalid credentials - Double-check username and password"
		
	case steamlang.EResult_RateLimitExceeded:
		return "Too many login attempts - Wait 15+ minutes before trying again"
		
	case steamlang.EResult_AccountLoginDeniedNeedTwoFactor:
		return "Two-factor authentication required - Need to implement mobile authenticator code support"
		
	case steamlang.EResult_InvalidLoginAuthCode:
		return "Invalid Steam Guard code - The code may be expired (they expire in ~5 minutes) or go-steam may not be sending it correctly. Try getting a fresh code."
		
	case steamlang.EResult_AccountDisabled:
		return "Account is disabled - Contact Steam support"
		
	case steamlang.EResult_AccountNotFound:
		return "Account not found - Check username spelling"
		
	case steamlang.EResult_ServiceUnavailable:
		return "Steam service unavailable - Try again later"
		
	default:
		return fmt.Sprintf("Unknown authentication error - Check Steam status and go-steam library compatibility")
	}
}

func (ts *EnhancedTestSuite) printAuthDiagnostics(authEvents chan string) {
	fmt.Println("\n🔍 Authentication Diagnostics Summary:")
	fmt.Println("=====================================")
	
	close(authEvents)
	eventCount := 0
	for event := range authEvents {
		eventCount++
		fmt.Printf("  %d. %s\n", eventCount, event)
	}
	
	if eventCount == 0 {
		fmt.Println("  No authentication events recorded - possible connection issue")
	}
	
	fmt.Println("\n💡 Recommendations:")
	fmt.Println("  1. Ensure the Steam account exists and credentials are correct")
	fmt.Println("  2. Log into Steam client first to authorize this device")  
	fmt.Println("  3. Check email for Steam Guard verification requests")
	fmt.Println("  4. Disable Steam Guard temporarily for testing (not recommended for production)")
	fmt.Println("  5. Check Steam server status at https://steamstat.us")
	fmt.Println("  6. Consider updating go-steam library for newer auth protocols")
}

// RunEnhancedTests runs diagnostic tests with better error reporting
func (ts *EnhancedTestSuite) RunEnhancedTests() {
	fmt.Println("🧪 Enhanced go-steam Diagnostic Tests")
	fmt.Println("====================================")

	// Test 1: Connection diagnostics
	fmt.Println("\n1️⃣  Testing Steam connection and authentication...")
	if err := ts.DiagnosticConnectAndAuth(); err != nil {
		fmt.Printf("❌ Connection/Auth test failed: %v\n", err)
		fmt.Println("\n🔧 This indicates the primary issue that needs fixing in go-steam")
		return
	}
	
	fmt.Println("✅ Connection and authentication successful!")
	
	// If we get here, we can run additional tests
	ts.runProtocolTests()
}

func (ts *EnhancedTestSuite) runProtocolTests() {
	fmt.Println("\n2️⃣  Testing Steam protocol features...")
	
	// Test persona state
	fmt.Println("   • Testing persona state changes...")
	ts.client.Social.SetPersonaState(steamlang.EPersonaState_Online)
	time.Sleep(2 * time.Second)
	
	// Test friends list
	fmt.Println("   • Testing friends list retrieval...")
	friends := ts.client.Social.Friends.GetCopy()
	fmt.Printf("   ✅ Retrieved %d friends\n", len(friends))
	
	// Test basic messaging capability
	fmt.Println("   • Testing messaging system availability...")
	if len(friends) > 0 {
		fmt.Println("   ✅ Friends available for messaging tests")
	} else {
		fmt.Println("   ⚠️  No friends available for messaging tests")
	}
	
	fmt.Println("\n✅ Basic protocol tests completed!")
}

// Disconnect cleanly closes the connection
func (ts *EnhancedTestSuite) Disconnect() {
	if ts.client != nil {
		ts.client.Disconnect()
	}
}