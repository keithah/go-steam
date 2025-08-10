# Changelog

All notable changes to this go-steam fork are documented here.

## [Unreleased] - 2025-08-09

### 🚀 Added
- **Steam Guard Email Support** - Full email verification code flow
- **Rate Limiting Protection** - SteamKit-style error handling and cooldown periods
- **GitHub-style CLI Tool** - Easy testing and debugging interface
- **Comprehensive Testing Framework** - Automated validation suite
- **Enhanced Error Analysis** - Detailed explanations for all authentication errors
- **Session Persistence** - Maintains authentication state between commands
- **Friend Management CLI** - Add friends by username or friend code
- **Event Listening System** - Enhanced event handling and debugging

### 🔧 Fixed
- **Critical Nil Pointer Crash** - Fixed segfault in `auth.go:113` when `WebapiAuthenticateUserNonce` is nil
- **Steam Guard Flow** - Completely rewritten authentication flow for modern Steam
- **Connection Stability** - Improved reconnection logic and session management
- **Error Handling** - Proper error categorization and user guidance

### 🏗 Changed
- **Authentication Architecture** - Moved from simple auth to stateful session management
- **CLI Structure** - GitHub-style command interface (`steam auth login`, `steam friends add`, etc.)
- **Error Messages** - Enhanced with specific guidance and next steps
- **Testing Approach** - Added multiple testing frameworks for different use cases

### 📊 SteamKit Improvements Ported
- **Authentication Exception Patterns** - Based on SteamKit's error handling model
- **Rate Limiting Strategy** - Similar cooldown and retry logic
- **Enhanced Protocol Safety** - Null checking for optional protobuf fields
- **Session Management** - Inspired by SteamKit's authentication session approach

### 🧪 Testing Additions
- **Interactive Test Framework** - Real-time Steam protocol testing
- **Enhanced Test Runner** - Comprehensive validation with detailed reporting
- **CLI Testing Tools** - Easy debugging and protocol exploration
- **Rate Limit Testing** - Utilities to test and reset rate limiting

### 📋 Known Limitations
- Friend requests work but acceptance may be limited by account restrictions
- Mobile Guard (TOTP) not yet implemented
- Some protocol messages may need updates for latest Steam features
- Trading features untested (inherited from original go-steam)

### 🔄 Migration from Original go-steam
- **Drop-in Replacement** - Same import path and API
- **Enhanced Error Handling** - More detailed error information
- **Additional CLI Tools** - New testing utilities
- **Session Management** - New persistent session files in `~/.steam-cli/`

## Previous Versions

This fork is based on the original go-steam repository (last updated December 2021). All changes prior to this fork are credited to the original maintainers.