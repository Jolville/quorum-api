package graph

import (
	"context"
	"net/http"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	IsVerified bool `json:"is_verified"`
	jwt.RegisteredClaims
}

type authCtxKey struct{}

// Middleware decodes the share session cookie and packs the session into context
func Middleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("authorization")
			token, err := jwt.ParseWithClaims(
				tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecret), nil
				})

			// Allow unauthenticated users in
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if claims, ok := token.Claims.(*JWTClaims); ok && claims.IsVerified {
				userID, err := uuid.Parse(claims.Subject)
				if err != nil {
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
