package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Philipp15b/go-steam/v3/protocol/steamlang"
)

// RateLimitTracker tracks authentication attempts to prevent hitting Steam's rate limits
type RateLimitTracker struct {
	LastAttempt      time.Time `json:"last_attempt"`
	ConsecutiveFails int       `json:"consecutive_fails"`
	RateLimited      bool      `json:"rate_limited"`
	RateLimitUntil   time.Time `json:"rate_limit_until"`
}

var rateLimitFile string

func init() {
	homeDir, _ := os.UserHomeDir()
	rateLimitFile = homeDir + "/.steam-cli/rate_limit.json"
}

func checkRateLimit() bool {
	tracker := getRateLimitTracker()
	
	// Check if we're currently rate limited
	if tracker.RateLimited && time.Now().Before(tracker.RateLimitUntil) {
		remaining := time.Until(tracker.RateLimitUntil)
		fmt.Printf("🚫 RATE LIMITED - %v remaining\n", remaining.Round(time.Minute))
		fmt.Println("   Steam has temporarily blocked login attempts.")
		fmt.Println("   This is normal protection against brute force attacks.")
		return true
	}
	
	// Check for recent rapid attempts (SteamKit style protection)
	if time.Since(tracker.LastAttempt) < 5*time.Second {
		fmt.Println("⚠️  LOGIN ATTEMPT TOO SOON")
		fmt.Println("   Please wait a few seconds between login attempts.")
		fmt.Println("   This helps avoid triggering Steam's rate limiting.")
		return true
	}
	
	// Warn if we've had recent failures
	if tracker.ConsecutiveFails >= 3 {
		fmt.Printf("⚠️  WARNING: %d consecutive auth failures\n", tracker.ConsecutiveFails)
		fmt.Println("   Consider waiting 15+ minutes to avoid rate limiting.")
		fmt.Println("   Steam may block the account temporarily after too many failures.")
	}
	
	return false
}

func recordAuthAttempt(result steamlang.EResult) {
	tracker := getRateLimitTracker()
	tracker.LastAttempt = time.Now()
	
	switch result {
	case steamlang.EResult_OK:
		// Success - reset counters
		tracker.ConsecutiveFails = 0
		tracker.RateLimited = false
		
	case steamlang.EResult_RateLimitExceeded:
		// Rate limited - set cooldown period
		tracker.ConsecutiveFails++
		tracker.RateLimited = true
		tracker.RateLimitUntil = time.Now().Add(15 * time.Minute) // Steam typically uses 15+ minute cooldowns
		fmt.Println("🚫 RATE LIMITED BY STEAM")
		fmt.Println("   Account temporarily blocked for 15+ minutes.")
		fmt.Println("   This is Steam's protection against brute force attacks.")
		
	case steamlang.EResult_InvalidPassword, 
		 steamlang.EResult_AccountLogonDenied,
		 steamlang.EResult_InvalidLoginAuthCode:
		// Auth failures - increment counter
		tracker.ConsecutiveFails++
		
		// If we have many failures, assume we might be approaching rate limit
		if tracker.ConsecutiveFails >= 5 {
			tracker.RateLimited = true
			tracker.RateLimitUntil = time.Now().Add(15 * time.Minute)
			fmt.Printf("🚫 TOO MANY FAILURES (%d) - Enforcing cooldown\n", tracker.ConsecutiveFails)
			fmt.Println("   Preventing further attempts to avoid Steam rate limiting.")
		}
		
	default:
		// Other errors don't count toward rate limiting
	}
	
	saveRateLimitTracker(tracker)
}

func getRateLimitTracker() *RateLimitTracker {
	data, err := os.ReadFile(rateLimitFile)
	if err != nil {
		return &RateLimitTracker{}
	}
	
	var tracker RateLimitTracker
	if err := json.Unmarshal(data, &tracker); err != nil {
		return &RateLimitTracker{}
	}
	
	return &tracker
}

func saveRateLimitTracker(tracker *RateLimitTracker) {
	data, _ := json.MarshalIndent(tracker, "", "  ")
	os.WriteFile(rateLimitFile, data, 0600)
}

func clearRateLimit() {
	os.Remove(rateLimitFile)
	fmt.Println("✅ Rate limit history cleared")
}

// Enhanced error analysis based on SteamKit patterns
func analyzeAuthError(result steamlang.EResult) string {
	switch result {
	case steamlang.EResult_RateLimitExceeded:
		return "Rate limit exceeded - Steam has temporarily blocked login attempts. Wait 15+ minutes and try again. This is normal protection against brute force attacks."
		
	case steamlang.EResult_AccountLogonDenied:
		return "Account login denied - Usually means Steam Guard email verification is required. Check your email for a verification code, or the account may be restricted."
		
	case steamlang.EResult_InvalidPassword:
		return "Invalid username or password - Double-check your credentials. Too many invalid attempts may trigger rate limiting."
		
	case steamlang.EResult_InvalidLoginAuthCode:
		return "Invalid Steam Guard code - The code may be expired (they expire quickly) or mistyped. Request a new code if needed."
		
	case steamlang.EResult_AccountLoginDeniedNeedTwoFactor:
		return "Two-factor authentication required - Need to implement mobile authenticator support or disable mobile guard temporarily."
		
	case steamlang.EResult_TwoFactorCodeMismatch:
		return "Two-factor code mismatch - The mobile authenticator code is incorrect or expired."
		
	case steamlang.EResult_AccountDisabled:
		return "Account is disabled - Contact Steam Support. The account may be banned or suspended."
		
	case steamlang.EResult_AccountNotFound:
		return "Account not found - Check the username spelling. The account may not exist or may be hidden."
		
	case steamlang.EResult_ServiceUnavailable:
		return "Steam service unavailable - Steam servers may be down or under maintenance. Check steamstat.us for server status."
		
	case steamlang.EResult_Timeout:
		return "Connection timeout - Network issues or Steam servers are slow. Check your internet connection and try again."
		
	default:
		return fmt.Sprintf("Unknown authentication error (%v) - Check Steam status and verify your account is in good standing.", result)
	}
}