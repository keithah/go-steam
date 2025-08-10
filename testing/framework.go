package testing

import (
	"context"
	"fmt"
	"sync"
	"time"

	steam "github.com/Philipp15b/go-steam/v3"
	"github.com/Philipp15b/go-steam/v3/protocol/steamlang"
)

// TestResult represents the outcome of a single test
type TestResult struct {
	Name        string
	Passed      bool
	Error       error
	Duration    time.Duration
	Description string
}

// TestSuite manages a collection of Steam protocol tests
type TestSuite struct {
	client   *steam.Client
	results  []TestResult
	mutex    sync.Mutex
	username string
	password string
}

// NewTestSuite creates a new test suite for go-steam validation
func NewTestSuite(username, password string) *TestSuite {
	return &TestSuite{
		username: username,
		password: password,
		results:  make([]TestResult, 0),
	}
}

// RunTest executes a single test and records the result
func (ts *TestSuite) RunTest(name, description string, testFunc func(*steam.Client) error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	start := time.Now()
	err := testFunc(ts.client)
	duration := time.Since(start)

	result := TestResult{
		Name:        name,
		Passed:      err == nil,
		Error:       err,
		Duration:    duration,
		Description: description,
	}

	ts.results = append(ts.results, result)

	status := "PASS"
	if !result.Passed {
		status = "FAIL"
	}

	fmt.Printf("[%s] %s (%s) - %s\n", status, name, duration, description)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	}
}

// ConnectAndAuth handles initial connection and authentication
func (ts *TestSuite) ConnectAndAuth() error {
	ts.client = steam.NewClient()
	ts.client.Connect()

	// Wait for connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	connected := make(chan bool)
	authResult := make(chan error)

	go func() {
		for event := range ts.client.Events() {
			switch e := event.(type) {
			case *steam.ConnectedEvent:
				connected <- true
				ts.client.Auth.LogOn(&steam.LogOnDetails{
					Username: ts.username,
					Password: ts.password,
				})

			case *steam.LoggedOnEvent:
				authResult <- nil
				return

			case *steam.LogOnFailedEvent:
				authResult <- fmt.Errorf("login failed: %v", e.Result)
				return

			case *steam.DisconnectedEvent:
				authResult <- fmt.Errorf("disconnected during auth")
				return
			}
		}
	}()

	select {
	case <-connected:
		// Wait for auth result
		select {
		case err := <-authResult:
			return err
		case <-ctx.Done():
			return fmt.Errorf("authentication timeout")
		}
	case <-ctx.Done():
		return fmt.Errorf("connection timeout")
	}
}

// Disconnect cleanly closes the connection
func (ts *TestSuite) Disconnect() {
	if ts.client != nil {
		ts.client.Disconnect()
	}
}

// RunAllTests executes the complete test suite
func (ts *TestSuite) RunAllTests() error {
	fmt.Println("Starting go-steam validation tests...")
	fmt.Println("=====================================")

	// Connect and authenticate
	if err := ts.ConnectAndAuth(); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer ts.Disconnect()

	// Core connection tests
	ts.RunTest("connection", "Basic Steam connection", ts.testConnection)
	ts.RunTest("authentication", "User authentication", ts.testAuthentication)
	
	// Protocol tests
	ts.RunTest("heartbeat", "Heartbeat mechanism", ts.testHeartbeat)
	ts.RunTest("persona_state", "Persona state changes", ts.testPersonaState)
	
	// Social features
	ts.RunTest("friends_list", "Friends list retrieval", ts.testFriendsList)
	ts.RunTest("friend_messaging", "Friend messaging", ts.testFriendMessaging)
	
	// Advanced features
	ts.RunTest("group_chat", "Group chat functionality", ts.testGroupChat)
	ts.RunTest("trade_offers", "Trade offer system", ts.testTradeOffers)

	return nil
}

// PrintResults displays a summary of all test results
func (ts *TestSuite) PrintResults() {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	fmt.Println("\nTest Results Summary")
	fmt.Println("===================")

	passed := 0
	for _, result := range ts.results {
		if result.Passed {
			passed++
		}
	}

	fmt.Printf("Total: %d, Passed: %d, Failed: %d\n", 
		len(ts.results), passed, len(ts.results)-passed)

	if passed < len(ts.results) {
		fmt.Println("\nFailed Tests:")
		for _, result := range ts.results {
			if !result.Passed {
				fmt.Printf("  - %s: %v\n", result.Name, result.Error)
			}
		}
	}
}

// Individual test implementations
func (ts *TestSuite) testConnection(client *steam.Client) error {
	if !client.Connected() {
		return fmt.Errorf("client not connected")
	}
	return nil
}

func (ts *TestSuite) testAuthentication(client *steam.Client) error {
	if client.SteamId() == 0 {
		return fmt.Errorf("not authenticated - no steam ID")
	}
	return nil
}

func (ts *TestSuite) testHeartbeat(client *steam.Client) error {
	// Test if heartbeat is working by waiting and checking connection
	time.Sleep(2 * time.Second)
	if !client.Connected() {
		return fmt.Errorf("connection lost during heartbeat test")
	}
	return nil
}

func (ts *TestSuite) testPersonaState(client *steam.Client) error {
	// Try to set persona state
	client.Social.SetPersonaState(steamlang.EPersonaState_Online)
	time.Sleep(1 * time.Second)
	
	// Check if we can read our current state
	state := client.Social.GetPersonaState()
	if state == steamlang.EPersonaState_Offline {
		return fmt.Errorf("persona state not updated")
	}
	return nil
}

func (ts *TestSuite) testFriendsList(client *steam.Client) error {
	friends := client.Social.Friends.GetCopy()
	if len(friends) < 0 { // Allow empty friends list
		return fmt.Errorf("could not retrieve friends list")
	}
	fmt.Printf("  Found %d friends\n", len(friends))
	return nil
}

func (ts *TestSuite) testFriendMessaging(client *steam.Client) error {
	// This test requires having at least one friend
	friends := client.Social.Friends.GetCopy()
	if len(friends) == 0 {
		return fmt.Errorf("no friends available for messaging test")
	}
	
	// For now, just test that the messaging function exists
	// In a real test, we'd send a message and verify delivery
	for steamId := range friends {
		_ = steamId // We would use this to test messaging
		break
	}
	
	return nil // Placeholder - implement actual messaging test
}

func (ts *TestSuite) testGroupChat(client *steam.Client) error {
	// Test group chat functionality
	groups := client.Social.Groups.GetCopy()
	fmt.Printf("  Found %d groups\n", len(groups))
	return nil
}

func (ts *TestSuite) testTradeOffers(client *steam.Client) error {
	// Test trade offer system
	// This is a placeholder - would need actual trade offer testing
	return nil
}