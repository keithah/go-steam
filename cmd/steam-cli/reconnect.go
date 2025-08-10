package main

import (
	"fmt"
	"time"

	steam "github.com/Philipp15b/go-steam/v3"
	"github.com/Philipp15b/go-steam/v3/protocol/steamlang"
)

// ensureConnection makes sure we have an active authenticated connection
func ensureConnection() error {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	// Check if we already have a working connection
	if globalClient != nil && isSessionActive() {
		return nil
	}

	fmt.Println("🔄 Reconnecting to Steam...")

	session := getCurrentSession()
	if !session.Authenticated || session.Username == "" {
		return fmt.Errorf("not authenticated - use 'steam auth login' first")
	}

	// Clean up any existing session
	if globalClient != nil {
		globalClient.Disconnect()
	}

	// Create new client
	globalClient = steam.NewClient()

	// Skip auto-login in event handler 
	skipAutoLogin = true

	// Start simplified event handling for reconnection
	go handleReconnectEvents()

	// Connect
	globalClient.Connect()

	// Wait for connection
	time.Sleep(3 * time.Second)

	// Log back in with stored credentials
	globalClient.Auth.LogOn(&steam.LogOnDetails{
		Username: session.Username,
		Password: session.Password,
	})

	// Wait for authentication
	time.Sleep(5 * time.Second)

	// Re-enable auto-login
	skipAutoLogin = false

	// Set online status
	if globalClient.SteamId() != 0 {
		globalClient.Social.SetPersonaState(steamlang.EPersonaState_Online)
		fmt.Println("✅ Reconnected successfully")
		return nil
	}

	return fmt.Errorf("failed to reconnect")
}

func handleReconnectEvents() {
	for event := range globalClient.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			// Connection established - login will be handled manually
			
		case *steam.LoggedOnEvent:
			// Authentication successful
			updateSessionState(func(state *SessionState) {
				state.Connected = true
				state.Authenticated = true
				state.SteamID = uint64(globalClient.SteamId())
			})

		case *steam.LogOnFailedEvent:
			fmt.Printf("❌ Reconnection failed: %v\n", e.Result)
			
		case *steam.DisconnectedEvent:
			// Handle disconnection
			
		case *steam.ChatMsgEvent:
			fmt.Printf("📨 Message from %d: %s\n", e.ChatterId, e.Message)
		}
	}
}

func isSessionActive() bool {
	if globalClient == nil {
		return false
	}
	
	// Check if we're connected and have a valid Steam ID
	return globalClient.Connected() && globalClient.SteamId() != 0
}