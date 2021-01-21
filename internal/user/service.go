package user

import (
	"context"
	"github.com/MuhammadChandra19/go-grpc-chat/api/v1"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/errors"
)

type Service struct {
	Repository RepositoryInterface
}

var (
	ErrAlreadyRegister = errors.N(errors.CodeValidationError, "data sudah terdaftar")
	ErrUserNotFound    = errors.N(errors.CodeNotFoundError, "data user tidak ditemukan")
)

func (s *Service) RegisterUser(ctx context.Context, req *v1.User) (*v1.RegisterResponse, error) {

	_, err := s.Repository.getByEmail(ctx, req.Email)

	if err == nil {
		return nil, ErrAlreadyRegister
	}

	user := User{
		Name:     req.GetName(),
		Email:    req.GetEmail(),
		PhotoURL: req.GetPhotourl(),
	}

	err = s.Repository.InsertUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &v1.RegisterResponse{
		Result: req,
	}, nil
}

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

func (s *Service) SignIn(ctx context.Context, req *v1.SignInRequest) (*v1.User, error) {
	user, err := s.Repository.getByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	return &v1.User{
		Email:    user.Email,
		Username: user.Username,
		Name:     user.Name,
		Photourl: user.PhotoURL,
	}, nil
}
