package graph

import jwt "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	IsVerified bool `json:"is_verified"`
	jwt.RegisteredClaims
}
