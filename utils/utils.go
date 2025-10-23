package utils

import (
	"fmt"
	"strings"
)

func TypeInput(prompt string) string {
	var input string
	fmt.Print(prompt)
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}
