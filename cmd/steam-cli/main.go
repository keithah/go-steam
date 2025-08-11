package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "auth":
		handleAuth(args)
	case "friends":
		handleFriends(args)
	case "msg":
		handleMessage(args)
	case "status":
		handleStatus(args)
	case "daemon":
		handleDaemon(args)
	case "version":
		fmt.Println("steam-cli v1.0.0 (go-steam testing)")
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Steam CLI - GitHub-style interface for Steam testing")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  steam <command> [arguments]")
	fmt.Println()
	fmt.Println("Auth Commands:")
	fmt.Println("  steam auth login                    # Start authentication")
	fmt.Println("  steam auth code <CODE>              # Submit Steam Guard code")
	fmt.Println("  steam auth logout                   # End session")
	fmt.Println("  steam auth status                   # Check authentication status")
	fmt.Println()
	fmt.Println("Social Commands:")
	fmt.Println("  steam friends list                  # List friends")
	fmt.Println("  steam friends add <CODE>            # Add friend by friend code")
	fmt.Println("  steam friends remove <STEAM_ID>     # Remove friend")
	fmt.Println("  steam friends search <NAME>         # Search friends")
	fmt.Println("  steam msg <STEAM_ID> <MESSAGE>      # Send message")
	fmt.Println()
	fmt.Println("General Commands:")
	fmt.Println("  steam status                        # Overall status")
	fmt.Println("  steam daemon start                  # Start persistent connection")
	fmt.Println("  steam daemon stop                   # Stop persistent connection") 
	fmt.Println("  steam daemon status                 # Check daemon status")
	fmt.Println("  steam version                       # Show version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  steam auth login")
	fmt.Println("  steam daemon start")
	fmt.Println("  steam friends list")
	fmt.Println()
	fmt.Println("State is maintained in ~/.steam-cli/")
}

func handleDaemon(args []string) {
	if len(args) == 0 {
		fmt.Println("Daemon commands:")
		fmt.Println("  start   - Start persistent Steam connection")
		fmt.Println("  stop    - Stop persistent connection") 
		fmt.Println("  status  - Check daemon status")
		fmt.Println("  run     - Run daemon (internal use)")
		return
	}

	subcommand := args[0]
	
	switch subcommand {
	case "start":
		if err := startDaemon(); err != nil {
			fmt.Printf("❌ Failed to start daemon: %v\n", err)
			os.Exit(1)
		}
	case "stop":
		if err := stopDaemon(); err != nil {
			fmt.Printf("❌ Failed to stop daemon: %v\n", err)
			os.Exit(1)
		}
	case "status":
		if err := getDaemonStatus(); err != nil {
			fmt.Printf("❌ Failed to get daemon status: %v\n", err)
			os.Exit(1)
		}
	case "run":
		// This is called internally to run the daemon
		runDaemon()
	default:
		fmt.Printf("Unknown daemon command: %s\n", subcommand)
	}
}