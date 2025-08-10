package main

import (
	"fmt"
)

func handleFriends(args []string) {
	if len(args) == 0 {
		fmt.Println("Friends commands:")
		fmt.Println("  list              - List all friends")
		fmt.Println("  add <CODE>        - Send friend request by friend code")
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

	friends := globalClient.Social.Friends.GetCopy()
	
	fmt.Printf("👥 Friends (%d):\n", len(friends))
	if len(friends) == 0 {
		fmt.Println("   No friends found")
		return
	}

	for steamId, friend := range friends {
		status := getPersonaStateString(friend.PersonaState)
		fmt.Printf("   %d: %s (%s)\n", steamId, friend.Name, status)
	}
}