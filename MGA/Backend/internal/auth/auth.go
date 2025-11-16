package auth

import (
    "crypto/rand"
    "encoding/hex"
    "time"

    "golang.org/x/crypto/bcrypt"
)

func HashPassword(plain string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

func CheckPassword(hash, plain string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func NewToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}

type Session struct {
    Token        string
    UserID       int64
    Username     string
    Role         string
    HomeownerID  int64 // for guests/technicians linked to a homeowner
    ExpiresAt    time.Time
}



