package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	steam "github.com/Philipp15b/go-steam/v3"
	"github.com/Philipp15b/go-steam/v3/protocol/steamlang"
)

// Session state that persists between commands
type SessionState struct {
	Username     string    `json:"username"`
	Password     string    `json:"password"` // Temporarily store for Steam Guard flow
	Connected    bool      `json:"connected"`
	Authenticated bool     `json:"authenticated"`
	NeedsCode    bool      `json:"needs_code"`
	SteamID      uint64    `json:"steam_id"`
	LastError    string    `json:"last_error"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AuthStatus for displaying current status
type AuthStatus struct {
	Connected     bool
	Authenticated bool
	NeedsCode     bool
	SteamID       uint64
	Username      string
	LastError     string
}

var (
	globalClient *steam.Client
	clientMutex  sync.Mutex
	sessionFile  string
	skipAutoLogin bool // Flag to prevent auto-login in event handler
)

func init() {
	// Create ~/.steam-cli directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Cannot get home directory: %v", err))
	}
	
	configDir := filepath.Join(homeDir, ".steam-cli")
	os.MkdirAll(configDir, 0700)
	
	sessionFile = filepath.Join(configDir, "session.json")
}

func startAuthSession(username, password string) error {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	// Clean up any existing session
	if globalClient != nil {
		globalClient.Disconnect()
	}

	// Create new client
	globalClient = steam.NewClient()
	
	// Store password temporarily for auth (not persisted)
	tempPassword = password
	
	// Save initial state (including password for Steam Guard flow)
	state := &SessionState{
		Username:      username,
		Password:      password, // Store temporarily for Steam Guard
		Connected:     false,
		Authenticated: false,
		NeedsCode:     false,
		UpdatedAt:     time.Now(),
	}
	saveSessionState(state)
	
	// Start event handling in background
	go handleSteamEvents()

	// Connect and wait briefly for initial connection
	globalClient.Connect()
	
	// Wait up to 10 seconds for connection events
	fmt.Println("⏳ Connecting to Steam...")
	time.Sleep(10 * time.Second)
	
	return nil
}

var tempPassword string // Temporary storage, not persisted

func handleSteamEvents() {
	for event := range globalClient.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			fmt.Println("✅ Connected to Steam servers")
			updateSessionState(func(state *SessionState) {
				state.Connected = true
				state.LastError = ""
			})
			
			// Only attempt auto-login if not skipping
			if !skipAutoLogin {
				session := getCurrentSession()
				globalClient.Auth.LogOn(&steam.LogOnDetails{
					Username: session.Username,
					Password: session.Password,
				})
			}

		case *steam.LoggedOnEvent:
			// Record successful auth
			recordAuthAttempt(steamlang.EResult_OK)
			
			fmt.Println("✅ Authentication successful!")
			updateSessionState(func(state *SessionState) {
				state.Authenticated = true
				state.NeedsCode = false
				state.SteamID = uint64(globalClient.SteamId())
				state.LastError = ""
			})
			
			// Clear temp password
			tempPassword = ""
			
			// Set online status
			globalClient.Social.SetPersonaState(steamlang.EPersonaState_Online)

		case *steam.LogOnFailedEvent:
			// Record the auth attempt for rate limiting
			recordAuthAttempt(e.Result)
			
			errorMsg := analyzeAuthError(e.Result)
			fmt.Printf("❌ %s\n", errorMsg)
			
			updateSessionState(func(state *SessionState) {
				state.Authenticated = false
				state.LastError = fmt.Sprintf("Authentication failed: %v", e.Result)
				
				switch e.Result {
				case steamlang.EResult_AccountLogonDenied:
					state.NeedsCode = true
					fmt.Println("📧 Use: steam auth code <CODE>")
					
				case steamlang.EResult_InvalidLoginAuthCode:
					state.NeedsCode = true
					fmt.Println("📧 Use: steam auth code <NEW_CODE>")
					
				case steamlang.EResult_RateLimitExceeded:
					fmt.Println("🚫 Account temporarily blocked - wait 15+ minutes")
					
				default:
					// For other errors, don't prompt for codes
				}
			})

		case *steam.DisconnectedEvent:
			fmt.Println("❌ Disconnected from Steam")
			updateSessionState(func(state *SessionState) {
				state.Connected = false
				state.Authenticated = false
			})
			
			// Auto-reconnect if we're waiting for a Steam Guard code
			session := getCurrentSession()
			if session.NeedsCode && !session.Authenticated {
				fmt.Println("🔄 Reconnecting for Steam Guard submission...")
				time.Sleep(2 * time.Second)
				globalClient.Connect()
			}

		case *steam.ChatMsgEvent:
			fmt.Printf("📨 Message from %d: %s\n", e.ChatterId, e.Message)
		}
	}
}

func submitSteamGuardCode(code string) error {
	session := getCurrentSession()
	if !session.NeedsCode {
		return fmt.Errorf("Steam Guard code not currently needed")
	}

	if session.Username == "" {
		return fmt.Errorf("no username stored - use 'steam auth login' first")
	}

	fmt.Println("🔄 Reconnecting with Steam Guard code...")
	
	clientMutex.Lock()
	defer clientMutex.Unlock()

	// Clean up any existing session
	if globalClient != nil {
		globalClient.Disconnect()
	}

	// Create new client
	globalClient = steam.NewClient()
	
	// Skip auto-login in event handler
	skipAutoLogin = true
	
	// Start event handling in background
	go handleSteamEvents()

	// Connect
	globalClient.Connect()
	
	// Wait for connection, then submit with code
	time.Sleep(3 * time.Second)
	
	// Attempt login with code
	globalClient.Auth.LogOn(&steam.LogOnDetails{
		Username: session.Username,
		Password: session.Password, // Use stored password
		AuthCode: code,
	})
	
	// Re-enable auto-login for future
	skipAutoLogin = false

	updateSessionState(func(state *SessionState) {
		state.NeedsCode = false
	})
	
	// Wait for auth result
	time.Sleep(5 * time.Second)

	return nil
}

func endAuthSession() error {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	if globalClient != nil {
		globalClient.Disconnect()
		globalClient = nil
	}

	// Clear session file
	return os.Remove(sessionFile)
}

func isAuthenticated() bool {
	session := getCurrentSession()
	return session.Connected && session.Authenticated
}

func getAuthStatus() AuthStatus {
	session := getCurrentSession()
	return AuthStatus{
		Connected:     session.Connected,
		Authenticated: session.Authenticated,
		NeedsCode:     session.NeedsCode,
		SteamID:       session.SteamID,
		Username:      session.Username,
		LastError:     session.LastError,
	}
}

func getCurrentSession() *SessionState {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		// Return empty session if file doesn't exist
		return &SessionState{}
	}

	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return &SessionState{}
	}

	return &state
}

func saveSessionState(state *SessionState) error {
	state.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionFile, data, 0600)
}

func updateSessionState(updateFunc func(*SessionState)) error {
	state := getCurrentSession()
	updateFunc(state)
	return saveSessionState(state)
}