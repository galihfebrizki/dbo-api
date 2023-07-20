package middleware

import "github.com/dgrijalva/jwt-go"

type UserInfo struct {
	Username string
	UserId   string
	AuthSign string
}

// Claims represents the JWT claims
type Claims struct {
	Username      string `json:"username"`
	UserId        string `json:"user_id"`
	AuthSignature string `json:"auth_sign"`
	jwt.StandardClaims
}
