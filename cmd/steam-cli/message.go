package main

import (
	"fmt"
	"strconv"
	"strings"
)

func handleMessage(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: steam msg <STEAM_ID> <MESSAGE>")
		fmt.Println("Example: steam msg 76561198000000000 \"Hello there!\"")
		return
	}

	if !isAuthenticated() {
		fmt.Println("❌ Not authenticated")
		fmt.Println("   Use: steam auth login")
		return
	}

	steamIdStr := args[0]
	message := strings.Join(args[1:], " ")

	// Parse Steam ID
	steamId, err := strconv.ParseUint(steamIdStr, 10, 64)
	if err != nil {
		fmt.Printf("❌ Invalid Steam ID: %s\n", steamIdStr)
		return
	}

	fmt.Printf("📤 Sending message to %d: %s\n", steamId, message)
	
	// TODO: Implement actual message sending once we have working auth
	fmt.Println("   Message sending not yet implemented")
	fmt.Println("   This will be added once authentication is fully working")
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