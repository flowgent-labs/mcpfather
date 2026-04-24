//go:build !darwin && !windows

package mcputils

import "sync"

var keychainToken string
var keychainOnce sync.Once

var wincredToken string
var wincredOnce sync.Once

func getFromKeychain() string { return "" }

func getFromWinCred() string { return "" }
