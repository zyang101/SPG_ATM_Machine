package utils

import (
	"fmt"
	"strings"
	"strconv"
	"regexp"

)

func TypeInput(prompt string) string {
	var input string
	fmt.Println(prompt)
	fmt.Scanln(&input)
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

func Deposit(username string)	{
	depositAmount := TypeInput("Enter Amount to Deposit: ")
	floatAmount, ok := ParseAmount(depositAmount)
	if !ok	{
		return
	}
	// add db change here

	fmt.Printf("Deposited $%.2f into %s's account.\n", floatAmount, username)
}

func Withdraw(username string)	{
	withdrawAmount := TypeInput("Enter Amount to Withdraw: ")
	floatAmount, ok:= ParseAmount(withdrawAmount)
	if !ok	{
		return
	}
	// add db change here
	fmt.Printf("Withdrew $%.2f from %s's account\n", floatAmount, username)
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