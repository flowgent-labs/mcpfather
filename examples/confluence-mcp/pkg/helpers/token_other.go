//go:build !darwin && !windows

package mcputils

func getFromKeychain() string { return "" }

func getFromWinCred() string { return "" }
