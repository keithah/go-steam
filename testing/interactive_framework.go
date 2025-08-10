package testing

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	steam "github.com/Philipp15b/go-steam/v3"
	"github.com/Philipp15b/go-steam/v3/protocol/steamlang"
)

// InteractiveTestSession manages a persistent Steam connection with interactive auth
type InteractiveTestSession struct {
	client      *steam.Client
	username    string
	password    string
	logFile     *os.File
	logger      *log.Logger
	mutex       sync.Mutex
	
	// State tracking
	connected   bool
	needsCode   bool
	authenticated bool
	
	// Channels for interaction
	authCodeChan chan string
	statusChan   chan string
}

// NewInteractiveTestSession creates a new interactive test session
func NewInteractiveTestSession(username, password string) (*InteractiveTestSession, error) {
	// Create log file
	logFile, err := os.OpenFile("steam-test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %v", err)
	}
	
	logger := log.New(logFile, "", log.LstdFlags)
	
	session := &InteractiveTestSession{
		username:     username,
		password:     password,
		logFile:      logFile,
		logger:       logger,
		authCodeChan: make(chan string),
		statusChan:   make(chan string, 10),
	}
	
	return session, nil
}

// Start begins the interactive Steam session
func (s *InteractiveTestSession) Start() {
	s.log("=== Starting Interactive Steam Session ===")
	s.log(fmt.Sprintf("Username: %s", s.username))
	
	fmt.Println("🚀 Starting interactive Steam test session...")
	fmt.Println("📝 Logging to: steam-test.log")
	fmt.Println()
	
	s.client = steam.NewClient()
	
	// Start event handler
	go s.handleEvents()
	
	// Connect to Steam
	s.log("Connecting to Steam servers...")
	fmt.Println("⏳ Connecting to Steam servers...")
	s.client.Connect()
	
	// Wait for connection and handle authentication flow
	s.waitForConnection()
}

func (s *InteractiveTestSession) handleEvents() {
	for event := range s.client.Events() {
		s.mutex.Lock()
		
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			s.connected = true
			s.log("✅ Connected to Steam servers")
			fmt.Println("✅ Connected to Steam servers")
			
			// Attempt initial authentication
			s.log("Attempting authentication...")
			fmt.Println("🔐 Attempting authentication...")
			
			s.client.Auth.LogOn(&steam.LogOnDetails{
				Username: s.username,
				Password: s.password,
			})
			
		case *steam.LoggedOnEvent:
			s.authenticated = true
			s.log("✅ Authentication successful!")
			fmt.Println("✅ Authentication successful!")
			fmt.Printf("   Steam ID: %d\n", s.client.SteamId())
			
			// Set online status
			s.client.Social.SetPersonaState(steamlang.EPersonaState_Online)
			
			// Start messaging tests
			go s.runMessagingTests()
			
		case *steam.LogOnFailedEvent:
			s.log(fmt.Sprintf("❌ Authentication failed: %v", e.Result))
			
			switch e.Result {
			case steamlang.EResult_AccountLogonDenied:
				s.needsCode = true
				s.log("🔐 Steam Guard email verification required")
				fmt.Println("🔐 Steam Guard email verification required!")
				fmt.Println("   Check your email for the verification code.")
				fmt.Println("   When you have the code, type: code <YOUR_CODE>")
				
			case steamlang.EResult_InvalidLoginAuthCode:
				s.needsCode = true
				s.log("❌ Invalid Steam Guard code provided")
				fmt.Println("❌ Invalid Steam Guard code!")
				fmt.Println("   The code may be expired or incorrect.")
				fmt.Println("   Get a fresh code and type: code <NEW_CODE>")
				
			default:
				s.log(fmt.Sprintf("❌ Unhandled auth error: %v", e.Result))
				fmt.Printf("❌ Authentication error: %v\n", e.Result)
			}
			
		case *steam.DisconnectedEvent:
			s.connected = false
			s.authenticated = false
			s.log("❌ Disconnected from Steam")
			fmt.Println("❌ Disconnected from Steam")
			
		case *steam.ChatMsgEvent:
			s.log(fmt.Sprintf("📨 Message from %d: %s", e.ChatterId, e.Message))
			fmt.Printf("📨 Message from %d: %s\n", e.ChatterId, e.Message)
			
		case *steam.MachineAuthUpdateEvent:
			s.log("🔐 Machine auth update received")
			
		default:
			s.log(fmt.Sprintf("Event: %T", e))
		}
		
		s.mutex.Unlock()
	}
}

func (s *InteractiveTestSession) waitForConnection() {
	// Start interactive command handler
	go s.handleCommands()
	
	// Keep session alive
	for {
		time.Sleep(1 * time.Second)
		
		s.mutex.Lock()
		if s.connected && s.authenticated {
			s.mutex.Unlock()
			break
		}
		s.mutex.Unlock()
	}
	
	fmt.Println("\n✅ Session established! Available commands:")
	fmt.Println("  code <CODE>     - Provide Steam Guard code")
	fmt.Println("  friends         - List friends") 
	fmt.Println("  msg <ID> <TEXT> - Send message to friend")
	fmt.Println("  status          - Show current status")
	fmt.Println("  quit            - Exit session")
	fmt.Println()
}

func (s *InteractiveTestSession) handleCommands() {
	scanner := bufio.NewScanner(os.Stdin)
	
	for scanner.Scan() {
		command := scanner.Text()
		s.log(fmt.Sprintf("Command: %s", command))
		
		if len(command) == 0 {
			continue
		}
		
		parts := strings.Fields(command)
		if len(parts) == 0 {
			continue
		}
		
		switch parts[0] {
		case "code":
			if len(parts) < 2 {
				fmt.Println("Usage: code <STEAM_GUARD_CODE>")
				continue
			}
			
			s.submitSteamGuardCode(parts[1])
			
		case "friends":
			s.listFriends()
			
		case "msg":
			if len(parts) < 3 {
				fmt.Println("Usage: msg <STEAM_ID> <MESSAGE>")
				continue
			}
			s.sendMessage(parts[1], strings.Join(parts[2:], " "))
			
		case "status":
			s.showStatus()
			
		case "quit":
			s.log("User requested quit")
			s.Stop()
			return
			
		default:
			fmt.Printf("Unknown command: %s\n", parts[0])
		}
	}
}

func (s *InteractiveTestSession) submitSteamGuardCode(code string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.needsCode {
		fmt.Println("❌ Steam Guard code not currently needed")
		return
	}
	
	s.log(fmt.Sprintf("Submitting Steam Guard code: %s", code))
	fmt.Printf("🔐 Submitting Steam Guard code: %s\n", code)
	
	s.client.Auth.LogOn(&steam.LogOnDetails{
		Username: s.username,
		Password: s.password,
		AuthCode: code,
	})
	
	s.needsCode = false
}

func (s *InteractiveTestSession) listFriends() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.authenticated {
		fmt.Println("❌ Not authenticated")
		return
	}
	
	friends := s.client.Social.Friends.GetCopy()
	s.log(fmt.Sprintf("Listed %d friends", len(friends)))
	
	fmt.Printf("👥 Friends (%d):\n", len(friends))
	for steamId, friend := range friends {
		fmt.Printf("   %d: %s (%s)\n", steamId, friend.Name, friend.PersonaState)
	}
}

func (s *InteractiveTestSession) sendMessage(steamIdStr, message string) {
	fmt.Printf("📤 Would send to %s: %s\n", steamIdStr, message)
	s.log(fmt.Sprintf("Message command: %s -> %s", steamIdStr, message))
	// TODO: Implement actual message sending
}

func (s *InteractiveTestSession) showStatus() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	fmt.Println("📊 Session Status:")
	fmt.Printf("   Connected: %v\n", s.connected)
	fmt.Printf("   Authenticated: %v\n", s.authenticated)
	fmt.Printf("   Needs Code: %v\n", s.needsCode)
	if s.authenticated {
		fmt.Printf("   Steam ID: %d\n", s.client.SteamId())
	}
}

func (s *InteractiveTestSession) runMessagingTests() {
	s.log("Starting messaging tests...")
	
	// Test friends list
	friends := s.client.Social.Friends.GetCopy()
	s.log(fmt.Sprintf("Friends test: %d friends found", len(friends)))
	
	// Test persona state
	s.client.Social.SetPersonaState(steamlang.EPersonaState_Online)
	s.log("Persona state test: Set to online")
	
	// More tests can be added here
	s.log("Basic messaging tests completed")
}

func (s *InteractiveTestSession) log(message string) {
	if s.logger != nil {
		s.logger.Println(message)
	}
}

func (s *InteractiveTestSession) Stop() {
	s.log("=== Stopping Interactive Steam Session ===")
	
	if s.client != nil {
		s.client.Disconnect()
	}
	
	if s.logFile != nil {
		s.logFile.Close()
	}
	
	fmt.Println("👋 Session ended. Check steam-test.log for details.")
}