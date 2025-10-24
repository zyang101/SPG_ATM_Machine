package main

import (
    "SPG_ATM_Machine/auth"
    "SPG_ATM_Machine/internal/db"
    "SPG_ATM_Machine/utils"
    "fmt"
    "strings"
)

func main() {
	db, err := db.Connect()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	fmt.Println("SQLITE DB connected")
	fmt.Println("Welcome to JP Goldman Stanley ATM!")
	for {
		answer := strings.ToUpper(utils.TypeInput("Would you like to Login? Y/N"))
		if answer == "Y" {
			isSucess, username := auth.Login()
			if isSucess {
				fmt.Println("Login Successful")
				auth.RouteUser(username)
			}	else	{
				fmt.Println("Username or Password is incorrect, Try Again")
			}
		} else if answer == "N" {
			fmt.Println("Bye Bye!")
			break
		} else {
			fmt.Println("Please answer Y or N.")
		}
	}

}
