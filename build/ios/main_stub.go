//go:build !ios

package main

// Stub main function for non-iOS builds so that `go build ./...` can compile
// this directory without errors. The real iOS entry point is in main_ios.go.
func main() {}
