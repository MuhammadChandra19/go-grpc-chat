package auth

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

// Service struct
type Service struct {
	tokenSecret string
}

// JwtPayload payload for jwt token
type JwtPayload struct {
	jwt.StandardClaims
	// *v1.User
	data interface{}
}

// NewJWTManager returns a new JWT manager
func NewJWTManager(secretKey string) *Service {
	return &Service{secretKey}
}

func (s *Service) Generate(payload interface{}) (string, error) {
	claims := JwtPayload{
		StandardClaims: jwt.StandardClaims{
			Issuer: "Kopdar",
		},
		data: payload,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("secret"))
}

// Verify verifies the access token string and return a user claim if the token is valid
func (s *Service) Verify(accessToken string) (interface{}, error) {
	key, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte("secret"), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := key.Claims.(jwt.MapClaims)

	if !ok || !key.Valid {
		return nil, err
	}
	return claims, nil
}
