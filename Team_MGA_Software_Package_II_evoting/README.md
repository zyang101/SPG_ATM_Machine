# SPC Pentagon: E-Voting Machine
CLI-based electronic voting system written in Go.  
Supports multiple user roles: Admin, Voter, and Election Official.

## Prerequisites
- [Go 1.25.4](https://go.dev/doc/install)

## Running the Project

In the root of the project run the following
```bash
# ensure you're at the root of the project
cd spc-evoting 
pwd 

# should print: ./spc-evoting

# install all dependencies
go mod tidy
```

To run the server, ensure you're located at the root of the project (in one terminal):
```bash
go run cmd/server/main.go
```

To run the client CLI (in a separate terminal):
```bash
go run cmd/client/main.go
```

After successfully running the server then client, on the client terminal you should see:
```bash
Welcome to SPC-Evoting!
---------------------------
Enter your username: 
```
## Notes
* See [User Manual PDF](./SPC_EVoting_UserManual.pdf)
* To create users (including the District Official to open/close elections), you will need an Election Admin account. You can use the account with username `akwok1` and password `12345` to start out. 
* If the prompt asks for User ID, that also means username
* Before creating an election, know the User ID of the District Official that will be tied to the election. Only that specific Official will be able to open/close/tally the election.
* After creating a new election, you will need to restart the server to see the election
  