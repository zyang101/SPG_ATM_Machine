package main

import (
	"fmt"
	"spc-evoting/internal/client"
)

func main() {

	// during actual run time the login will return what kind of user has
	// the right privelages
	username, token, acct_type := client.Login()

	fmt.Println("Successfully logged in!")

	switch acct_type {
	case "admin":
		admin := client.Admin{}
		admin.Run(token, username)
	case "official":
		client.OfficialLoop(token, username)
	case "voter":
		client.VoterLoop(token, username)
	default:
		fmt.Println("404 Unauthorized")
	}

}
