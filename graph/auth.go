package graph

import (
	"context"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	IsVerified bool `json:"is_verified"`
	jwt.RegisteredClaims
}

type authCtxKey struct{}

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, _ := strings.CutPrefix(r.Header.Get("authorization"), "Bearer ")
			token, err := jwt.ParseWithClaims(
				tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecret), nil
				})

			// Allow unauthenticated users in
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if claims, ok := token.Claims.(*JWTClaims); ok && claims.IsVerified {
				userID, err := uuid.Parse(claims.Subject)
				if err == nil {
					ctx := context.WithValue(r.Context(), authCtxKey{}, userID)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetVerifiedCustomer(ctx context.Context) uuid.NullUUID {
	raw, ok := ctx.Value(authCtxKey{}).(uuid.UUID)
	if !ok {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{
		Valid: true,
		UUID:  raw,
	}
}
