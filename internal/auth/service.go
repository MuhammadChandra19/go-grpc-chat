package auth

import (
	"fmt"

	v1 "github.com/MuhammadChandra19/go-grpc-chat/api/v1"
	"github.com/dgrijalva/jwt-go"
)

// Service struct
type Service struct {
	tokenSecret string
}

// JwtPayload payload for jwt token
type JwtPayload struct {
	jwt.StandardClaims
	*v1.User
}

// NewJWTManager returns a new JWT manager
func NewJWTManager(secretKey string) *Service {
	return &Service{secretKey}
}

// GenerateJwtToken generates and signs a new token for a user
func (s *Service) GenerateJwtToken(payload *v1.User) (string, error) {
	claims := JwtPayload{
		StandardClaims: jwt.StandardClaims{
			Issuer: "Kopdar",
		},
		User: payload,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.tokenSecret))
}

// Verify verifies the access token string and return a user claim if the token is valid
func (s *Service) Verify(accessToken string) (*JwtPayload, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&JwtPayload{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return []byte(s.tokenSecret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JwtPayload)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
