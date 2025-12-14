package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"spc-evoting/internal/schemes"
	"strings"

	"github.com/manifoldco/promptui"
)

// cool green arrow from promptui
// var greenArrow = "\033[1;32mv\033[0m"

// templates
var registerTemplate = `Register a new user
---------------------------------
| type exit to quit the program |
---------------------------------
`

// error codes
var errReturn = errors.New("user requested to return to menu")

// apparently you can add methods to a class / struct this way
// user can be a voter or an election official
func (a *Admin) RegisterUser(token string) {
	user := map[string]string{
		"Username":      "",
		"Password":      "",
		"First Name":    "",
		"Last Name":     "",
		"Date of Birth": "",
		"Account Type":  "",
	}

	url := "http://localhost:8080/users"

	for {
		prompt := promptui.Select{
			Label: registerTemplate,
			Items: []string{
				fmt.Sprintf("Username: %s", user["Username"]),
				fmt.Sprintf("Password: %s", user["Password"]),
				fmt.Sprintf("First Name: %s", user["First Name"]),
				fmt.Sprintf("Last Name: %s", user["Last Name"]),
				fmt.Sprintf("Date of Birth: %s", user["Date of Birth"]),
				fmt.Sprintf("Account Type: %s", user["Account Type"]),
				"Submit",
				"Return to Menu",
			},
		}

		i, _, _ := prompt.Run()

		switch i {
		case 0:
			userN, _ := getUserInput("Enter user's username")
			user["Username"] = userN
		case 1:
			pass, _ := getUserInput("Enter user's password")
			user["Password"] = pass
		case 2:
			fname, _ := getUserInput("Enter user's first name")
			user["First Name"] = fname
		case 3:
			lname, _ := getUserInput("Enter user's last name")
			user["Last Name"] = lname
		case 4:
			dob, _ := getUserInput("Enter user's date of birth (YYYY-MM-DD)")
			user["Date of Birth"] = dob
		case 5:
			getAccountType(user)
		case 6:
			// check if they left anything out
			if isIncomplete(user) {
				continue
			} else {

				// handle payload
				payload := schemes.CreateUserReq{
					UserID:    user["Username"],
					Password:  user["Password"],
					FirstName: user["First Name"],
					LastName:  user["Last Name"],
					DOB:       user["Date of Birth"],
					Role:      user["Account Type"],
				}

				data, _ := json.Marshal(payload)
				req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))

				if err != nil {
					fmt.Printf("Error creating request: %v\n", err)
					return
				}

				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				res, err := http.DefaultClient.Do(req)
				if err != nil {
					fmt.Printf("Error registering user: %v\n", err)
					return
				}

				defer res.Body.Close()

				body, err := io.ReadAll(res.Body)
				if err != nil {
					fmt.Println("Error reading response:", err)
					os.Exit(1)
				}

				// handle HTTP errors
				if res.StatusCode != http.StatusNoContent {
					fmt.Printf("Registration failed (%d): %s\n", res.StatusCode, string(body))
					os.Exit(1)
				}

				clear()
				fmt.Printf("\n🎉  You registered a new %s\n!", user["Account Type"])
				return
			}
		case 7:
			// exit
			clear()
			return
		}
	}
}

// HELPERS:

// manual prompt for input
func getUserInput(prompt string) (string, error) {

	// create a reader from standard in
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", prompt)
	result, _ := reader.ReadString('\n') // this def has a security vulnerability
	result = strings.TrimSpace(result)

	if strings.EqualFold(result, "exit") {
		clear()
		os.Exit(0) // exit gracefully
	} else if strings.EqualFold(result, "return") {
		return "return", errReturn
	}

	return result, nil
}

// this is probably not the way to go about this but also the only
// way I could find online that actually clears the terminal
func clear() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls") // Windows
	default:
		cmd = exec.Command("clear") // macOS / Linux
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

// select user account type
func getAccountType(user map[string]string) {
	prompt := promptui.Select{
		Label: "Choose the user account type:",
		Items: []string{
			"Voter",
			"Election Official",
			"Admin",
		},
	}

	i, _, _ := prompt.Run()
	switch i {
	case 0:
		// indicate that they're a voter in the http request
		user["Account Type"] = "voter"
	case 1:
		// indicate it's an election official
		user["Account Type"] = "official"
	case 2:
		user["Account Type"] = "admin"
	default:
		return
	}
}

func isIncomplete(user map[string]string) bool {
	var incomplete = false
	for key, value := range user {
		if value == "" {
			fmt.Printf("⚠️  %s is missing\n", key)
			incomplete = true
		}
	}
	return incomplete
}
