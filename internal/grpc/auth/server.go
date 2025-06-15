package auth

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/services/auth"
	"sso/internal/storage"

	ssov1 "github.com/iluha481/protos/gen/go/sso"

	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}
type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, refresh_token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	RefreshToken(
		ctx context.Context,
		refresh_token string,
		appID int,
	) (token string, new_refresh_token string, err error)
}

func Register(gRPCServer *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	in *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if in.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	token, refresh_token, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword(), int(in.GetAppId()))
	if err != nil {
		// Ошибку auth.ErrInvalidCredentials мы создадим ниже
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		fmt.Print(err)
		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssov1.LoginResponse{Token: token, RefreshToken: refresh_token}, nil
}
func (s *serverAPI) Register(
	ctx context.Context,
	in *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		// Ошибку storage.ErrUserExists мы создадим ниже
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssov1.RegisterResponse{UserId: uid}, nil
}

func (s *serverAPI) RefreshToken(
	ctx context.Context,
	in *ssov1.RefreshRequest,
) (*ssov1.RefreshResponse, error) {
	if in.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}
	if in.AppId == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}
	token, refresh_token, err := s.auth.RefreshToken(ctx, in.RefreshToken, int(in.AppId))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to refresh token")
	}
	return &ssov1.RefreshResponse{Token: token, RefreshToken: refresh_token}, nil
}
