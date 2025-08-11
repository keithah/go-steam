package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	steam "github.com/Philipp15b/go-steam/v3"
	"github.com/Philipp15b/go-steam/v3/protocol/steamlang"
)

const (
	daemonPidFile = ".steam-cli/daemon.pid"
	daemonStateFile = ".steam-cli/daemon_state.json"
)

type DaemonState struct {
	PID         int       `json:"pid"`
	StartTime   time.Time `json:"start_time"`
	Connected   bool      `json:"connected"`
	SteamID     uint64    `json:"steam_id,string"`
	Username    string    `json:"username"`
}

// Check if daemon is running
func isDaemonRunning() bool {
	pidPath := filepath.Join(os.Getenv("HOME"), daemonPidFile)
	data, err := ioutil.ReadFile(pidPath)
	if err != nil {
		return false
	}
	
	var state DaemonState
	if err := json.Unmarshal(data, &state); err != nil {
		return false
	}
	
	// Check if process exists
	process, err := os.FindProcess(state.PID)
	if err != nil {
		return false
	}
	
	// Try to send signal 0 to check if process is alive
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// Start the daemon
func startDaemon() error {
	if isDaemonRunning() {
		return fmt.Errorf("daemon already running")
	}
	
	// Fork the process
	cmd := os.Args[0]
	args := []string{"daemon", "run"}
	
	attr := &os.ProcAttr{
		Files: []*os.File{nil, nil, nil}, // Detach from terminal
	}
	
	process, err := os.StartProcess(cmd, append([]string{cmd}, args...), attr)
	if err != nil {
		return fmt.Errorf("failed to start daemon: %v", err)
	}
	
	// Save daemon state
	state := DaemonState{
		PID:       process.Pid,
		StartTime: time.Now(),
	}
	
	if err := saveDaemonState(state); err != nil {
		return fmt.Errorf("failed to save daemon state: %v", err)
	}
	
	fmt.Printf("✅ Steam daemon started (PID: %d)\n", process.Pid)
	return nil
}

// Run the daemon (this is the actual daemon process)
func runDaemon() {
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Load session
	session := getCurrentSession()
	if !session.Authenticated {
		fmt.Println("No authenticated session found")
		os.Exit(1)
	}
	
	// Create client
	client := steam.NewClient()
	
	// Start event handler
	go handleDaemonEvents(client)
	
	// Connect
	client.Connect()
	time.Sleep(2 * time.Second)
	
	// Login
	client.Auth.LogOn(&steam.LogOnDetails{
		Username: session.Username,
		Password: session.Password,
	})
	
	// Update daemon state periodically
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-sigChan:
			fmt.Println("Daemon shutting down...")
			client.Disconnect()
			cleanupDaemon()
			os.Exit(0)
			
		case <-ticker.C:
			// Update state
			state := DaemonState{
				PID:       os.Getpid(),
				StartTime: time.Now(),
				Connected: client.Connected(),
				SteamID:   uint64(client.SteamId()),
				Username:  session.Username,
			}
			saveDaemonState(state)
		}
	}
}

func handleDaemonEvents(client *steam.Client) {
	for event := range client.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			fmt.Println("Daemon connected to Steam")
			
		case *steam.LoggedOnEvent:
			fmt.Println("Daemon authenticated")
			
			// CRITICAL: Set online status explicitly and persistently
			fmt.Println("🐛 DEBUG: Setting persona state to ONLINE (not snooze/sleeping)")
			client.Social.SetPersonaState(steamlang.EPersonaState_Online)
			fmt.Println("Daemon set to online status")
			
			// Wait a moment then set online again to ensure it sticks
			go func() {
				time.Sleep(2 * time.Second)
				fmt.Println("🐛 DEBUG: Reinforcing ONLINE persona state")
				client.Social.SetPersonaState(steamlang.EPersonaState_Online)
				
				// Keep reinforcing online status every 30 seconds to prevent sleeping
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						if client.Connected() {
							fmt.Println("🐛 DEBUG: Periodic ONLINE status reinforcement")
							client.Social.SetPersonaState(steamlang.EPersonaState_Online)
						}
					}
				}
			}()
			
			// CRITICAL: Request offline messages immediately after login - this enables message delivery
			fmt.Println("🐛 DEBUG: Requesting offline messages to initialize message delivery (SteamKit pattern)")
			client.Social.RequestOfflineMessages()
			fmt.Println("🐛 DEBUG: Offline messages requested - should enable incoming message packets")
			
		case *steam.DisconnectedEvent:
			fmt.Println("Daemon disconnected, reconnecting...")
			time.Sleep(5 * time.Second)
			client.Connect()
			
		case *steam.LogOnFailedEvent:
			fmt.Printf("Daemon login failed: %v\n", e.Result)
			
		case *steam.ChatMsgEvent:
			// Log messages to a file for the CLI to read
			logMessage(e)
		}
	}
}

func saveDaemonState(state DaemonState) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".steam-cli")
	os.MkdirAll(configDir, 0700)
	
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	
	pidPath := filepath.Join(configDir, "daemon.pid")
	return ioutil.WriteFile(pidPath, data, 0600)
}

func cleanupDaemon() {
	configDir := filepath.Join(os.Getenv("HOME"), ".steam-cli")
	pidPath := filepath.Join(configDir, "daemon.pid")
	os.Remove(pidPath)
}

func stopDaemon() error {
	if !isDaemonRunning() {
		return fmt.Errorf("daemon not running")
	}
	
	pidPath := filepath.Join(os.Getenv("HOME"), daemonPidFile)
	data, err := ioutil.ReadFile(pidPath)
	if err != nil {
		return err
	}
	
	var state DaemonState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	
	process, err := os.FindProcess(state.PID)
	if err != nil {
		return err
	}
	
	// Send SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon: %v", err)
	}
	
	fmt.Println("✅ Steam daemon stopped")
	return nil
}

func getDaemonStatus() error {
	if !isDaemonRunning() {
		fmt.Println("❌ Steam daemon not running")
		return nil
	}
	
	pidPath := filepath.Join(os.Getenv("HOME"), daemonPidFile)
	data, err := ioutil.ReadFile(pidPath)
	if err != nil {
		return err
	}
	
	var state DaemonState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	
	fmt.Println("✅ Steam daemon status:")
	fmt.Printf("   PID: %d\n", state.PID)
	fmt.Printf("   Started: %s\n", state.StartTime.Format(time.RFC3339))
	fmt.Printf("   Connected: %v\n", state.Connected)
	if state.SteamID != 0 {
		fmt.Printf("   Steam ID: %d\n", state.SteamID)
		fmt.Printf("   Username: %s\n", state.Username)
	}
	
	return nil
}

func logMessage(msg *steam.ChatMsgEvent) {
	// Log received messages to console for now
	// In a full implementation, this would write to a messages.json file
	fmt.Printf("[DAEMON] 📨 Message from %d: %s\n", msg.ChatterId, msg.Message)
}