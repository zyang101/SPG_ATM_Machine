package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"spc-evoting/internal/schemes"
)

var welcome = `Welcome to SPC-Evoting!
---------------------------`

func Login() (string, string, string) {

	// set up request
	url := "http://localhost:8080/auth/login"
	var username = ""
	var password = ""

	fmt.Printf("%s\n", welcome)
	for {

		username, _ = getUserInput("Enter your username")
		password, _ = getUserInput("Enter your password")

		if username == "" || password == "" {
			continue
		} else {
			break
		}

	}

	payload := schemes.LoginReq{
		UserID:   username,
		Password: password,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error encoding request:", err)
		os.Exit(1)
	}

	res, err := http.Post(url, "application/json", bytes.NewBuffer(data))

	if err != nil {
		fmt.Printf("Error sending request %s", err)
		os.Exit(0)
	}

	// close the response body, don't leave descriptors open
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	// handle HTTP errors
	if res.StatusCode != http.StatusOK {
		fmt.Printf("Login failed with status(%d): %s\n", res.StatusCode, string(body))
		os.Exit(1)
	}

	// read the response
	var result schemes.LoginRes
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Error parsing JSON:", err)
		fmt.Println("Raw response:", string(body))
		os.Exit(1)
	}

	return username, result.Token, result.Role

}
