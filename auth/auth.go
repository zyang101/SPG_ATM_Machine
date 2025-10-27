package auth

import (
	"SPG_ATM_Machine/admin"
	"SPG_ATM_Machine/internal/db"
	"SPG_ATM_Machine/customer"
	"SPG_ATM_Machine/handler"
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

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

	// Look up stored hash for username
	var storedHash string
	query := `SELECT pin FROM users WHERE username = ?`
	err = conn.QueryRow(query, username).Scan(&storedHash)

	if err == sql.ErrNoRows {
		fmt.Println("Invalid username or PIN.")
		return false, ""
	} else if err != nil {
		fmt.Println("Database error:", err)
		return false, ""
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(pin))
	if err != nil {
		fmt.Println("Invalid username or PIN.")
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

	var role string
	err = conn.QueryRow(`SELECT role FROM users WHERE username = ?`, username).Scan(&role)
	if err != nil {
		fmt.Println("Error fetching user role:", err)
		return
	}

	switch strings.ToLower(role) {
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
