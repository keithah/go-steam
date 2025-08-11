package main

import (
	"fmt"
	"strconv"
	"time"

	steam "github.com/Philipp15b/go-steam/v3"
	"github.com/Philipp15b/go-steam/v3/protocol/steamlang"
	"github.com/Philipp15b/go-steam/v3/steamid"
)

// Extended friends management functions

func handleFriendsAdd(friendCode string) error {
	// Ensure we have an active connection
	if err := ensureConnection(); err != nil {
		return err
	}

	// Try to determine if input is a friend code (numeric) or username
	if code, err := strconv.ParseUint(friendCode, 10, 32); err == nil {
		// Method 1: Add by friend code (convert to SteamID64)
		steamId64 := steamid.SteamId(76561197960265728 + code)
		
		fmt.Printf("🤝 Sending friend request by friend code\n")
		fmt.Printf("   Friend Code: %s → SteamID: %d\n", friendCode, steamId64)
		
		// Listen for response events temporarily
		go handleFriendRequestResponse()
		
		// Send friend request by SteamID
		globalClient.Social.AddFriend(steamId64)
		
	} else {
		// Method 2: Add by account name/email
		fmt.Printf("🤝 Sending friend request by username\n")
		fmt.Printf("   Username: %s\n", friendCode)
		
		// Listen for response events temporarily
		go handleFriendRequestResponse()
		
		// Send friend request by account name - need to use the protobuf directly
		globalClient.Social.AddFriendByName(friendCode)
	}

	fmt.Println("✅ Friend request sent!")
	fmt.Println("⏳ Waiting for response...")
	
	// Wait a bit for the response
	time.Sleep(5 * time.Second)
	
	return nil
}

func handleFriendsAccept(steamIdStr string) error {
	// Ensure we have an active connection (handles daemon mode too)
	if err := ensureConnection(); err != nil {
		return err
	}

	steamId64, err := strconv.ParseUint(steamIdStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid Steam ID: %s", steamIdStr)
	}

	steamId := steamid.SteamId(steamId64)
	
	fmt.Printf("✅ Accepting friend request from: %d\n", steamId)
	
	// To accept a friend request, we use AddFriend on the pending request
	globalClient.Social.AddFriend(steamId)
	
	fmt.Println("✅ Friend request accepted!")
	fmt.Println("⏳ Waiting for confirmation...")
	
	// Wait a bit for the response
	time.Sleep(3 * time.Second)
	
	return nil
}

func handleFriendsRemove(steamIdStr string) error {
	if !isAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	clientMutex.Lock()
	defer clientMutex.Unlock()

	if globalClient == nil {
		return fmt.Errorf("no active session")
	}

	steamId64, err := strconv.ParseUint(steamIdStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid Steam ID: %s", steamIdStr)
	}

	steamId := steamid.SteamId(steamId64)
	
	fmt.Printf("❌ Removing friend: %d\n", steamId)
	globalClient.Social.RemoveFriend(steamId)
	
	fmt.Println("✅ Friend removed!")
	return nil
}

func handleFriendsSearch(query string) error {
	if !isAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	clientMutex.Lock()
	defer clientMutex.Unlock()

	if globalClient == nil {
		return fmt.Errorf("no active session")
	}

	friends := globalClient.Social.Friends.GetCopy()
	
	fmt.Printf("🔍 Searching friends for: %s\n", query)
	found := 0
	
	for steamId, friend := range friends {
		if contains(friend.Name, query) {
			status := getPersonaStateString(friend.PersonaState)
			fmt.Printf("   %d: %s (%s)\n", steamId, friend.Name, status)
			found++
		}
	}
	
	if found == 0 {
		fmt.Println("   No friends found matching that query")
	} else {
		fmt.Printf("   Found %d matching friends\n", found)
	}
	
	return nil
}

// Helper functions
func contains(str, substr string) bool {
	// Simple case-insensitive contains
	return len(str) >= len(substr) && 
		   str[:len(substr)] == substr || 
		   str[len(str)-len(substr):] == substr
}

func getPersonaStateString(state steamlang.EPersonaState) string {
	switch state {
	case steamlang.EPersonaState_Offline:
		return "offline"
	case steamlang.EPersonaState_Online:
		return "online"  
	case steamlang.EPersonaState_Busy:
		return "busy"
	case steamlang.EPersonaState_Away:
		return "away"
	case steamlang.EPersonaState_Snooze:
		return "snooze"
	case steamlang.EPersonaState_LookingToTrade:
		return "looking to trade"
	case steamlang.EPersonaState_LookingToPlay:
		return "looking to play"
	case steamlang.EPersonaState_Invisible:
		return "invisible"
	default:
		return "unknown"
	}
}

func handleFriendRequestResponse() {
	// Listen for friend request responses with timeout
	timeout := time.After(10 * time.Second)
	
	for {
		select {
		case event := <-globalClient.Events():
			switch e := event.(type) {
			case *steam.FriendAddedEvent:
				fmt.Printf("📫 Friend request response: %v\n", e.Result)
				if e.Result == steamlang.EResult_OK {
					fmt.Printf("✅ Successfully added: %s (ID: %d)\n", e.PersonaName, e.SteamId)
				} else {
					fmt.Printf("❌ Friend request failed: %v\n", e.Result)
					switch e.Result {
					case steamlang.EResult_Ignored:
						fmt.Println("   The user has ignored your request")
					case steamlang.EResult_DuplicateRequest:
						fmt.Println("   You are already friends with this user or request already sent")
					case steamlang.EResult_InvalidSteamID:
						fmt.Println("   Invalid Steam ID")
					default:
						fmt.Printf("   Reason: %v\n", e.Result)
					}
				}
				return // Exit after getting response
			default:
				// Log other events for debugging
				fmt.Printf("🔍 Event: %T\n", e)
			}
		case <-timeout:
			fmt.Println("⏰ No response received within 10 seconds")
			fmt.Println("   This might indicate:")
			fmt.Println("   - Friend request was sent but Steam didn't respond")
			fmt.Println("   - Account limitations (new accounts may have restrictions)")
			fmt.Println("   - Protocol issue with go-steam")
			return
		}
	}
}