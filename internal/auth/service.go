package auth

import (
	"fmt"
	"time"

	"github.com/MuhammadChandra19/go-grpc-chat/internal/user"
	"github.com/dgrijalva/jwt-go"
)

// Service struct
type Service struct {
	tokenSecret string
}

// JwtPayload payload for jwt token
type JwtPayload struct {
	jwt.StandardClaims
	user.User
}

// NewJWTManager returns a new JWT manager
func NewJWTManager(secretKey string) *Service {
	return &Service{secretKey}
}

// GenerateJwtToken generates and signs a new token for a user
func (s *Service) GenerateJwtToken(payload *user.User, exp time.Duration) (*string, error) {
	claims := JwtPayload{
		StandardClaims: jwt.StandardClaims{
			Issuer: "Kopdar",
		},
		User: *payload,
	}
	signedToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.tokenSecret))
	if err != nil {
		return nil, err
	}

	return &signedToken, nil
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
