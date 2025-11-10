package auth

import (
	"SPG_ATM_Machine/admin"
	"SPG_ATM_Machine/customer"
	"SPG_ATM_Machine/handler"
	"SPG_ATM_Machine/internal/db"
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func ParseIDCard(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text(), nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("file is empty")
}

func Login() (bool, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Println("Enter PIN: ")
	bytePin, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("\nError reading password:", err)
		return false, ""
	}
	fmt.Println()

	pin := strings.TrimSpace(string(bytePin))

	conn, err := db.Connect()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return false, ""
	}
	defer conn.Close()

	stmt, err := conn.Prepare("SELECT pin FROM users WHERE username = ?")
	if err != nil {
		fmt.Println("Database prepare error:", err)
		return false, ""
	}
	defer stmt.Close()

	var storedHash string
	err = stmt.QueryRow(username).Scan(&storedHash)

	if err == sql.ErrNoRows {
		fmt.Println("Invalid login")
		return false, ""
	} else if err != nil {
		fmt.Println("Database error:", err)
		return false, ""
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(pin))
	if err != nil {
		fmt.Println("Invalid login.")
		return false, ""
	}

	return true, username
}

func RouteUser(username string) {

	conn, err := db.Connect()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	defer conn.Close()

	stmt, err := conn.Prepare("SELECT role FROM users WHERE username = ?")
	if err != nil {
		fmt.Println("Error preparing statement:", err)
		return
	}
	defer stmt.Close()

	var role string
	err = stmt.QueryRow(username).Scan(&role)
	if err != nil {
		fmt.Println("Error fetching user role:", err)
		return
	}

	cardRole, err := ParseIDCard("auth/idcard.txt")
	if err != nil {
		fmt.Println("Error reading ID card.")
		return
	}

	dbRole := strings.ToLower(role)
	if cardRole != dbRole {
		fmt.Println("Invalid login.")
		return
	}

	fmt.Println("Login Successful")
	switch dbRole {
	case "admin":
		admin.Menu(username)
	case "customer":
		customer.Menu(username)
	case "cash handler":
		handler.Menu(username)
	default:
		fmt.Println("Error validating user type")
	}
}
