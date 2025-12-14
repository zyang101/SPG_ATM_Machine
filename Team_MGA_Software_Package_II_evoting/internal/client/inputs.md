- `POST /auth/login` (Create user session)
  - Client sends:
    ```json
    {
        "username": "str",
        "password": "str"
    }
    ```
  - Server returns:
    ```json
    {
        "token": "str",
        "expires":  "str",
        "role": "str: admin,official,voter"
    }
    ```

- `POST /auth/logout` (Close the session)
  - Client sends no payload
  - Server returns: 204 code if successful 

- `POST /users` (Create a user)
  - Client sends:
    ```json
    {
        "username": "str",
        "password": "str",
        "first_name": "str",
        "last_name": "str",
        "date_of_birth": "str: yyyy-mm-dd",
        "role": "str: admin,official,voter"
    }
    ```
  - Server returns: 201 code if successful

- `POST /ballots` (Send casted ballot for an election)
  - Client sends (`CastedBallotPayload`):
    ```json
    {
        "username": "str",
        "election_id": 1234,
        "ballot": [
            {
                "position_id": 1234,
                "candidate_id": 1234
            },
            {
                "position_id": 1234,
                "candidate_id": 1234
            }
        ]
    }
    ```
  - Server returns: 201 code if successful 

- `GET /elections` (List of elections (only open elections for voters))
  - Client sends no payload
  - Server sends:
  ```json
  {
    "elections": [
        {
            "election_id": 1234, 
            "official_id": 1234,
            "election_name": "str",
            "district_name": "str",
            "is_active": 1
        }
    ]
  }
  ```

- `POST /elections` (Create an election)
  - Client sends (`ElectionPayload`):
    ```json
    {
        "election_name": "str",
        "official_username": "str",
        "district_name": "str",
        "positions": [
            {
                "name": "president",
                "candidates": [
                    {
                        "name": "John Doe",
                        "party": "Democrat"
                    },
                    {
                        "name": "Sarah Park",
                        "party": "Republican"
                    }
                ]
            }
        ]
    }
    ```
  - Server returns: 201 code if successful 

- `GET /elections/{id}` (Get election for voter to create ballot)
  - Client sends no payload
  - Server sends:
    ```json
    {
        "election_id": 1234,
        "election_name": "2026 NY State Elections",
        "positions": [
            {
                "position_id": 1234,
                "position_name": "Governor",
                "candidates": [
                    {
                        "id": 1234,
                        "name": "Jonathan Baker",
                        "party": "Democratic"
                    },
                    {
                        "id": 1234,
                        "name": "Jessica Adams",
                        "party": "Republican"
                    }
                ]
            }
        ]
    }
    ```

- `POST /election/{id}/open` (Open an election)
  - Client sends no payload
  - Server returns: 204 code if successful 
- `POST /election/{id}/close` (Close an election)
  - Client sends no payload
  - Server returns: 204 code if successful 
  
- `GET /elections/{id}/results` (View election results)
  - Client sends no payload
  - Server sends:
    ```json
    {
        "election_name": "str",
        "is_active": 0,
        "positions": [
            {
                "position_id": 1234,
                "position_name": "str",
                "winner_id": 1234,
                "candidates": [
                    {
                        "candidate_id": 1234,
                        "name": "str",
                        "party": "str",
                        "vote_count": 1234
                    }
                ]
            }
        ]
    }
    ```


- Example integration (credits: https://stackoverflow.com/questions/24455147/how-do-i-send-a-json-string-in-a-post-request-in-go)

```golang

import {
    "net/http"
    "encoding/json"
    "bytes"
}

type LoginPayload struct {
    User string `json:"user"`
    Pass string `json:"pass"`
}

func registerVoterRequest(user: string, pass: string) string {

    url := "http://localhost:3000/spc-evoting/registerVoter/"
    payload := LoginPayload{
        User: user,
        Pass: pass,
    }

    jsonData, _ := json.Marshal(payload)
    res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))

    if err != nil {
        fmt.Println("Error registering this voter...")
        os.Exit(0)
    }

    body, err := io.ReadAll(res.Body)
    if err != nil {
        fmt.Println("Error reading response:", err)
        os.Exit(1)
    }

    return string(body)
}

```