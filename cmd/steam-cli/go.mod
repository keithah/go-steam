module steam-cli

go 1.23.0

toolchain go1.24.5

require (
	github.com/Philipp15b/go-steam/v3 v3.0.0
	golang.org/x/term v0.34.0
)

require (
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

replace github.com/Philipp15b/go-steam/v3 => ../../
