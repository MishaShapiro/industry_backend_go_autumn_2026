package main

import (
	"fmt"
	"strings"
)

func greet(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = "World"
	}
	return fmt.Sprintf("Hello, %s!", trimmed)
}
