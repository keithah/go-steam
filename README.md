# go-steam (Improved Fork)

A modernized and bug-fixed fork of [go-steam](https://github.com/Philipp15b/go-steam) with improvements ported from [SteamKit](https://github.com/SteamRE/SteamKit).

## 🚀 Key Improvements

- ✅ **Fixed Steam Guard authentication** - Works with modern Steam email verification
- ✅ **Enhanced error handling** - SteamKit-style error analysis and rate limiting  
- ✅ **Nil pointer fixes** - Resolved critical crashes in authentication flow
- ✅ **GitHub-style CLI tool** - Easy testing and debugging interface
- ✅ **Comprehensive testing framework** - Automated validation of Steam protocol features

## 📊 Feature Status

| Feature | Status | Tested | Notes |
|---------|--------|--------|-------|
| **Authentication** |
| Basic Login | ✅ Working | ✅ Yes | Username/password authentication |
| Steam Guard (Email) | ✅ Working | ✅ Yes | Email verification code support |
| Steam Guard (Mobile) | ❌ Not Implemented | ❌ No | Requires mobile authenticator support |
| Rate Limiting Protection | ✅ Working | ✅ Yes | Prevents "too many retries" errors |
| Session Persistence | ✅ Working | ✅ Yes | Maintains auth between CLI commands |
| **Social Features** |
| Friends List | ✅ Working | ✅ Yes | Retrieve and display friends |
| Friend Requests (by username) | ⚠️ Partial | ❌ Limited | Sends requests, acceptance issues |
| Friend Requests (by friend code) | ⚠️ Partial | ❌ Limited | Protocol may need updates |
| Friend Removal | ✅ Working | ❌ No | Implementation exists, untested |
| **Messaging** |
| Send Messages | ✅ Working | ❌ No | Requires friends to test |
| Receive Messages | ✅ Working | ❌ No | Event handlers implemented |
| Group Chat | ✅ Working | ❌ No | Basic support exists |
| **Trading** |
| Trade Offers | ❓ Unknown | ❌ No | Legacy implementation, needs testing |
| Inventory | ❓ Unknown | ❌ No | Legacy implementation, needs testing |
| **Protocol** |
| Connection | ✅ Working | ✅ Yes | Stable Steam server connection |
| Heartbeat | ✅ Working | ✅ Yes | Maintains connection |
| Event Handling | ✅ Working | ✅ Yes | Callbacks and event system |
| Protobuf Messages | ✅ Working | ✅ Yes | Modern Steam protocol support |

## 🔧 SteamKit Improvements Ported

| Improvement | Original Issue | Status | Implementation |
|-------------|----------------|--------|----------------|
| **Authentication Error Handling** | Crashes on auth failures | ✅ Fixed | Enhanced error analysis with detailed explanations |
| **Rate Limiting Protection** | "Too many retries" blocks | ✅ Fixed | Automatic cooldown periods and attempt tracking |
| **Nil Pointer Safety** | Segfaults in auth flow | ✅ Fixed | Null checks for optional protobuf fields |
| **Steam Guard Flow** | Email verification broken | ✅ Fixed | Proper code submission and retry logic |
| **Session Management** | Auth state not persistent | ✅ Fixed | JSON-based session storage |
| **Protocol Version Updates** | Outdated message types | ⚠️ Partial | Some updates applied, more needed |

## 🛠 Testing Framework

### CLI Tool (`cmd/steam-cli/`)
```bash
# Authentication
./steam auth login <username> <password>
./steam auth code <CODE>
./steam auth status

# Social features  
./steam friends list
./steam friends add <username_or_code>
./steam msg <steam_id> <message>

# Utilities
./steam status
./steam auth clear-rate-limit  # For testing
```

### Test Suite (`testing/`)
```bash
# Run comprehensive tests
go run cmd/test-runner/enhanced_main.go --username <user> --password <pass> --enhanced

# Interactive testing
go run cmd/test-runner/interactive_main.go --username <user> --password <pass>
```

## 📋 Known Issues

| Issue | Impact | Workaround | Priority |
|-------|--------|------------|----------|
| Friend requests don't appear | Social features limited | Add friends manually in Steam client | Medium |
| New account restrictions | Testing limitations | Use established Steam accounts | Low |
| Mobile Guard not supported | Modern 2FA broken | Use email-based Steam Guard | High |
| Protocol version gaps | Some features may break | Continue porting SteamKit updates | Medium |

## 🧪 Testing Status

### ✅ **Thoroughly Tested**
- Basic authentication flow
- Steam Guard email verification
- Rate limiting and error handling  
- Session persistence
- Connection stability
- Friends list retrieval

### ⚠️ **Partially Tested**
- Friend request sending (protocol works, acceptance issues)
- Event handling system (basic events confirmed)

### ❌ **Not Yet Tested**
- Message sending/receiving (requires mutual friends)
- Group chat functionality
- Trading and inventory features
- Advanced social features

## 🔄 Ongoing Work

1. **Protocol Updates** - Continue porting SteamKit improvements
2. **Friend Request Debugging** - Investigate acceptance issues
3. **Mobile Guard Support** - Implement TOTP/mobile authenticator
4. **Message Testing** - Establish mutual friends for testing
5. **Matrix Bridge Development** - Use this as foundation

## 🤝 Original Credits

- **Original go-steam**: [Philipp15b/go-steam](https://github.com/Philipp15b/go-steam)
- **Protocol Reference**: [SteamRE/SteamKit](https://github.com/SteamRE/SteamKit)
- **Steam Protocol Documentation**: [SteamDB](https://steamdb.info/) and [SteamRE](https://github.com/SteamRE)

## 📝 License

Inherits the license from the original go-steam project (New BSD License).

## 🚀 Usage for Matrix Bridge

This fork is specifically maintained for Matrix bridge development. The enhanced authentication and error handling make it suitable for production use in bridge applications.

### Example Integration
```go
import "github.com/yourusername/go-steam-fork"

client := steam.NewClient()
// Enhanced error handling and rate limiting built-in
client.Connect()
```

## 📞 Support

For issues specific to this fork's improvements, please file issues in this repository. For general go-steam questions, refer to the original project.