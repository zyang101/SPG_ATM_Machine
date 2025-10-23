package main

import (

	"fmt"
    // "SPG_ATM_Machine/db"
    "SPG_ATM_Machine/auth"
	// "SPG_ATM_Machine/customer"
	// "SPG_ATM_Machine/admin"
)

func main() {
	fmt.Println("Welcome to JP Goldman Stanley ATM!")

	username, password := auth.Login()

    fmt.Println("Logged in as:", username, "with role:", password)
}
