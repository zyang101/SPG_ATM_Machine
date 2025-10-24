package utils

import (
	"fmt"
	"strings"
	"strconv"
)

func TypeInput(prompt string) string {
	var input string
	fmt.Println(prompt)
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

func parseAmount(amountStr string) (float64, bool) {
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

	floatAmount, ok := parseAmount(depositAmount)
	if !ok	{
		return
	}
	// add db change here

	fmt.Printf("Deposited $%.2f into %s's account.\n", floatAmount, username)
}

func Withdraw(username string)	{
	withdrawAmount := TypeInput("Enter Amount to Withdraw: ")
	floatAmount, ok:= parseAmount(withdrawAmount)
	if !ok	{
		return
	}
	// add db change here

	fmt.Printf("Withdrew $%.2f from %f's account\n", floatAmount, username)
}