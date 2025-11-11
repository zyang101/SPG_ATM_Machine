package auth

import (
	"SPG_ATM_Machine/admin"
	"SPG_ATM_Machine/customer"
	"SPG_ATM_Machine/handler"
	"SPG_ATM_Machine/internal/db"
	"SPG_ATM_Machine/internal/api"
	"bufio"
	"fmt"
	"os"
	"strings"
	"log"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

// Check idcard.txt for the role. Simulates putting in atm card. 
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

// Prompts the user for their username.
func PromptUsername() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')
	return strings.TrimSpace(username)
}

// Prompts the user for their PIN.
func PromptPIN() string {
	fmt.Print("Enter PIN: ")
	bytePin, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		log.Println("Error reading PIN:", err)
		return ""
	}
	return strings.TrimSpace(string(bytePin))
}

func Login() (bool, string) {
	username := PromptUsername()
	if username == "" {
		return false, ""
	}

	pin := PromptPIN()
	if pin == "" {
		return false, ""
	}

	conn, err := db.Connect()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return false, ""
	}
	defer conn.Close()

	userInfo, err := api.GetUserAuth(conn, username)
	if err != nil {
		fmt.Println("Invalid login.")
		return false, ""
	}

	if userInfo.Locked {
		fmt.Println("Account is locked. Contact admin.")
		return false, ""
	}

	// Compare hashed Pin
	err = bcrypt.CompareHashAndPassword([]byte(userInfo.PINHash), []byte(pin))
	if err != nil {
		// Increment failed attempts via API
		newAttempts, locked, apiErr := api.IncrementFailedAttempts(conn, username)
		if apiErr != nil {
			log.Println("DB error:", apiErr)
			fmt.Println("An error occurred. Contact admin.")
			return false, ""
		}

		if locked {
			fmt.Println("Too many failed attempts. Your account has been locked. Contact an Admin")
			return false, ""
		}

		fmt.Printf("Invalid login. (%d/3 attempts)\n", newAttempts)
		return false, ""
	}

	// Reset failed attempts on successful login
	if err := api.ResetFailedAttempts(conn, username); err != nil {
		log.Println("DB error:", err)
		fmt.Println("An error occurred. Contact admin.")
	}

	return true, username
}

func RouteUser(username string) {

	conn, err := db.Connect()
	if err != nil {
		log.Println("DB error:", err)
		fmt.Println("An error occurred. Contact admin.", err)
		return
	}
	defer conn.Close()


	dbRole, err := api.FetchUserRole(conn, username)
	if err != nil {
		log.Println("Error fetching role:", err)
		fmt.Println("An error occurred. Contact admin.")
		return
	}

	cardRole, err := ParseIDCard("auth/idcard.txt")
	if err != nil || cardRole != dbRole {
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
