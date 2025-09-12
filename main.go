package main

import "mixlink/internal/server"

// goreleaser release --snapshot --rm-dist
func main() {
	server.Start()
}
