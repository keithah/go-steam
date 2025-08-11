package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Philipp15b/go-steam/v3/protocol/steamlang"
	"github.com/Philipp15b/go-steam/v3/steamid"
)

func handleMessage(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: steam msg <STEAM_ID> <MESSAGE>")
		fmt.Println("Example: steam msg 76561198000000000 \"Hello there!\"")
		return
	}

	// Ensure we have an active connection (handles daemon mode too)
	if err := ensureConnection(); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	steamIdStr := args[0]
	message := strings.Join(args[1:], " ")

	// Parse Steam ID
	steamId64, err := strconv.ParseUint(steamIdStr, 10, 64)
	if err != nil {
		fmt.Printf("❌ Invalid Steam ID: %s\n", steamIdStr)
		return
	}

	// Convert to go-steam SteamId type
	steamId := steamid.SteamId(steamId64)
	
	// Check if target is a friend and refresh their online status
	friend, err := globalClient.Social.Friends.ById(steamId)
	if err != nil {
		fmt.Printf("❌ %d is not in your friends list\n", steamId)
		return
	}
	
	// Request fresh persona state for the target friend
	flags := steamlang.EClientPersonaStateFlag_PlayerName | 
			steamlang.EClientPersonaStateFlag_Presence | 
			steamlang.EClientPersonaStateFlag_SourceID
	globalClient.Social.RequestFriendInfo(steamId, flags)
	time.Sleep(time.Second) // Give Steam time to respond
	
	// Get updated friend info
	friend, err = globalClient.Social.Friends.ById(steamId)
	if err != nil {
		fmt.Printf("❌ %d is not in your friends list\n", steamId)
		return
	}
	
	fmt.Printf("📤 Sending message to %d (%s): %s\n", steamId, friend.Name, message)
	
	// Show warning if friend is offline (messages may not be delivered)
	if friend.PersonaState == steamlang.EPersonaState_Offline {
		fmt.Printf("⚠️  Warning: %s appears to be offline - message may not be delivered\n", friend.Name)
	} else {
		fmt.Printf("✅ %s is online - message should be delivered\n", friend.Name)
	}
	
	// Ensure we're set to online (offline messages already requested during login)
	fmt.Printf("🐛 DEBUG: Ensuring persona state is Online\n")
	globalClient.Social.SetPersonaState(steamlang.EPersonaState_Online)
	fmt.Printf("🐛 DEBUG: Ready to send message\n")
	
	// Send message using SteamKit-style implementation
	// Uses EChatEntryType_ChatMsg like SteamKit's default
	globalClient.Social.SendMessage(steamId, steamlang.EChatEntryType_ChatMsg, message)
	
	fmt.Println("✅ Message sent to Steam servers!")
	fmt.Println("💡 Note: For offline friends, messages may not be delivered until they come online")
}

func handleStatus(args []string) {
	fmt.Println("📊 Steam CLI Status")
	fmt.Println("==================")

	status := getAuthStatus()
	
	if status.Connected {
		fmt.Println("🔗 Connection: ✅ Connected")
	} else {
		fmt.Println("🔗 Connection: ❌ Disconnected")
	}

	if status.Authenticated {
		fmt.Println("🔐 Authentication: ✅ Authenticated")
		fmt.Printf("   Steam ID: %d\n", status.SteamID)
		fmt.Printf("   Username: %s\n", status.Username)
	} else if status.Connected {
		fmt.Println("🔐 Authentication: ⏳ In progress")
		if status.NeedsCode {
			fmt.Println("   📧 Steam Guard code required")
		}
	} else {
		fmt.Println("🔐 Authentication: ❌ Not authenticated")
	}

	if status.LastError != "" {
		fmt.Printf("⚠️  Last Error: %s\n", status.LastError)
	}

	fmt.Println()
	fmt.Println("💡 Next steps:")
	if !status.Connected {
		fmt.Println("   steam auth login")
	} else if status.NeedsCode {
		fmt.Println("   steam auth code <CODE>")
	} else if status.Authenticated {
		fmt.Println("   steam friends list")
		fmt.Println("   steam msg <ID> <MESSAGE>")
	}
}