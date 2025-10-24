package models

type User struct {
	ID          int
	FullName    string
	DOB         string
	PIN         string
	StartingBal float64
	Username    string
	Role        string
}
