package auth

import (
    "crypto/rand"
    "encoding/hex"
    "time"
    "crypto/sha1"
    "strings"

)

func HashPassword(plain string) (string, error) {
    h := sha1.Sum([]byte(plain))
    return hex.EncodeToString(h[:]), nil
}

func CheckPassword(hash, plain string) bool {
    h := sha1.Sum([]byte(plain))
    return strings.EqualFold(hash, hex.EncodeToString(h[:]))
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



