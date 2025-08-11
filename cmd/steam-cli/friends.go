package main

import (
	"fmt"
	"time"
	
	"github.com/Philipp15b/go-steam/v3/protocol/steamlang"
	"github.com/Philipp15b/go-steam/v3/steamid"
)

func handleFriends(args []string) {
	if len(args) == 0 {
		fmt.Println("Friends commands:")
		fmt.Println("  list              - List all friends and pending requests")
		fmt.Println("  add <CODE>        - Send friend request by friend code")
		fmt.Println("  accept <STEAM_ID> - Accept a pending friend request")
		fmt.Println("  remove <STEAM_ID> - Remove friend")
		fmt.Println("  search <NAME>     - Search friends by name")
		return
	}

	subcommand := args[0]
	subargs := args[1:]

	switch subcommand {
	case "list":
		handleFriendsList()
	case "add":
		if len(subargs) == 0 {
			fmt.Println("Usage: steam friends add <FRIEND_CODE_OR_NAME>")
			fmt.Println("Example: steam friends add 1926659806")
			fmt.Println("Example: steam friends add username")
			return
		}
		if err := handleFriendsAdd(subargs[0]); err != nil {
			fmt.Printf("❌ Failed to add friend: %v\n", err)
		}
	case "remove":
		if len(subargs) == 0 {
			fmt.Println("Usage: steam friends remove <STEAM_ID>")
			return
		}
		if err := handleFriendsRemove(subargs[0]); err != nil {
			fmt.Printf("❌ Failed to remove friend: %v\n", err)
		}
	case "accept":
		if len(subargs) == 0 {
			fmt.Println("Usage: steam friends accept <STEAM_ID>")
			fmt.Println("Use 'steam friends list' to see pending requests")
			return
		}
		if err := handleFriendsAccept(subargs[0]); err != nil {
			fmt.Printf("❌ Failed to accept friend request: %v\n", err)
		}
	case "search":
		if len(subargs) == 0 {
			fmt.Println("Usage: steam friends search <NAME>")
			return
		}
		if err := handleFriendsSearch(subargs[0]); err != nil {
			fmt.Printf("❌ Failed to search friends: %v\n", err)
		}
	default:
		fmt.Printf("Unknown friends command: %s\n", subcommand)
	}
}

func handleFriendsList() {
	// Ensure we have an active connection
	if err := ensureConnection(); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	// Request fresh friend info to get current online status
	friends := globalClient.Social.Friends.GetCopy()
	var friendIds []steamid.SteamId
	for steamId, friend := range friends {
		if friend.Relationship == 3 { // EFriendRelationship_Friend
			friendIds = append(friendIds, steamid.SteamId(steamId))
		}
	}
	
	// Request updated persona state for all friends
	if len(friendIds) > 0 {
		fmt.Printf("🔄 Requesting fresh persona state for %d friends...\n", len(friendIds))
		flags := steamlang.EClientPersonaStateFlag_PlayerName | 
				steamlang.EClientPersonaStateFlag_Presence | 
				steamlang.EClientPersonaStateFlag_SourceID |
				steamlang.EClientPersonaStateFlag_GameDataBlob
		globalClient.Social.RequestFriendListInfo(friendIds, flags)
		
		// Give Steam more time to respond with updated info
		time.Sleep(2 * time.Second)
		friends = globalClient.Social.Friends.GetCopy()
		fmt.Printf("✅ Persona state refresh complete\n")
	}
	
	// Count different relationship types
	pendingCount := 0
	friendCount := 0
	
	// Show pending requests first
	fmt.Println("📨 Pending Friend Requests:")
	for steamId, friend := range friends {
		if friend.Relationship == 2 { // EFriendRelationship_RequestRecipient
			pendingCount++
			status := getPersonaStateString(friend.PersonaState)
			fmt.Printf("   %d: %s (%s) - PENDING REQUEST\n", steamId, friend.Name, status)
		}
	}
	if pendingCount == 0 {
		fmt.Println("   No pending requests")
	}
	
	fmt.Println()
	
	// Show actual friends
	fmt.Println("👥 Friends:")
	for steamId, friend := range friends {
		if friend.Relationship == 3 { // EFriendRelationship_Friend
			friendCount++
			status := getPersonaStateString(friend.PersonaState)
			// Debug: show raw persona state value
			fmt.Printf("   %d: %s (%s) [raw state: %d]\n", steamId, friend.Name, status, int(friend.PersonaState))
		}
	}
	if friendCount == 0 {
		fmt.Println("   No friends found")
	}
	
	fmt.Printf("\nTotal: %d pending requests, %d friends\n", pendingCount, friendCount)
}