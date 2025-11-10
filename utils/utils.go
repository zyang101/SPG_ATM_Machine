package utils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func TypeInput(prompt string) string {
	fmt.Print(prompt + " ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func ParseAmount(amountStr string) (float64, bool) {
	amountStr = strings.TrimSpace(amountStr)

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		fmt.Println("Invalid input. Please enter a valid number (e.g., 100.50).")
		return 0, false
	}

	if amount <= 0 {
		fmt.Println("Amount must be greater than zero.")
		return 0, false
	}

	return amount, true
}

func ValidatePIN(pin string) bool {
	match, _ := regexp.MatchString(`^\d{6}$`, pin)
	if !match {
		fmt.Println("PIN must be exactly 6 digits.")
	}
	return match
}

func ValidateName(name string) bool {
	match, _ := regexp.MatchString(`^[A-Za-z\s]+$`, name)
	if !match {
		fmt.Println("Name can only contain letters and spaces.")
	}
	return match
}

func ValidateDate(date string) bool {
	match, _ := regexp.MatchString(`^(0[1-9]|1[0-2])/(0[1-9]|[12]\d|3[01])/\d{2,4}$`, date)
	if !match {
		fmt.Println("Date must be in MM/DD/YY or MM/DD/YYYY format.")
	}
	return match
}
