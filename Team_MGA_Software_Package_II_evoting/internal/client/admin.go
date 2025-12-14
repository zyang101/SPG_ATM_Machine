package client

import (
	"os"

	"github.com/manifoldco/promptui"
)

type Admin struct{}

// allow you to return void and run an infinite loop
func (a *Admin) Run(token string, username string) {
	for {
		prompt := promptui.Select{
			Label: "Select your task",
			Items: []string{
				"Create a new Election",
				"Register a new Voter",
				"Exit",
			},
		}
		_, result, _ := prompt.Run()

		switch result {
		case "Create a new Election":
			a.CreateElection(token, username)
		case "Register a new Voter":
			a.RegisterUser(token)
		case "Exit":
			os.Exit(0) // exit the application
		}
	}
}
