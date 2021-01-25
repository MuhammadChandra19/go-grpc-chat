package user

import (
	"context"

	v1 "github.com/MuhammadChandra19/go-grpc-chat/api/v1"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/auth"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	Repository RepositoryInterface
	jwtManager *auth.Service
}

var (
	ErrAlreadyRegister = errors.N(errors.CodeValidationError, "data sudah terdaftar")
	ErrUserNotFound    = errors.N(errors.CodeNotFoundError, "data user tidak ditemukan")
)

func (s *Service) RegisterUser(ctx context.Context, req *v1.User) (*v1.TokenResponse, error) {

	filter := map[string]interface{}{
		"email": req.Email,
	}
	_, err := s.Repository.GetOne(ctx, filter)
	print(err)
	if err == nil {
		return nil, ErrAlreadyRegister
	}

	user := User{
		Name:     req.Name,
		Email:    req.Email,
		PhotoURL: req.Photourl,
		Username: req.Username,
	}

	err = s.Repository.InsertUser(ctx, user)
	if err != nil {
		return nil, err
	}

	token, err := s.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	return &v1.TokenResponse{
		Token: token,
	}, nil
}

// SearchUser ...
func (s *Service) SearchUser(ctx context.Context, req *v1.SearchParams) (*v1.SearchResponse, error) {
	users, err := s.Repository.getUserList(ctx, req.Query)
	if err != nil {
		return nil, err
	}
	var result []*v1.User
	for _, row := range users {
		temp := &v1.User{
			Username: row.Username,
			Name:     row.Name,
			Email:    row.Email,
			Photourl: row.PhotoURL,
		}
		result = append(result, temp)
	}
	return &v1.SearchResponse{Users: result}, nil
}

// SignIn method
func (s *Service) SignIn(ctx context.Context, req *v1.SignInRequest) (*v1.TokenResponse, error) {
	filter := map[string]interface{}{
		"email": req.Email,
	}
	user, err := s.Repository.GetOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	token, err := s.jwtManager.Generate(user)

	return &v1.TokenResponse{
		Token: token,
	}, nil
}
