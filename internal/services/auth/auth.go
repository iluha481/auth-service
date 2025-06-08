package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

type Auth struct {
	log             *slog.Logger
	usrSaver        UserSaver
	usrProvider     UserProvider
	appProvider     AppProvider
	tokenTTL        time.Duration
	refreshTokenTTL time.Duration
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:        userSaver,
		usrProvider:     userProvider,
		log:             log,
		appProvider:     appProvider,
		tokenTTL:        tokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Login checks if user exists in the system and password correct, returns acess token
//
// if user exists, but password is incorrect, returns error
// if user doesnt exists, returns error
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
		// password
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	refresh_token, err := jwt.NewRefreshToken(user, app, a.refreshTokenTTL)
	if err != nil {
		a.log.Error("failed to generate refresh token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return token, refresh_token, nil
}

func (a *Auth) RefreshToken(
	ctx context.Context,
	refresh_token string,
	appID int,
) (string, string, error) {
	const op = "Auth.RefreshToken"
	app, err := a.appProvider.App(ctx, appID)

	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)

	}

	claims, err := jwt.ParseJwtToken(refresh_token, app.Refresh_secret)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	email := claims["email"].(string)
	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	access_token, err := jwt.NewToken(user, app, a.tokenTTL)

	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	new_refresh_token, err := jwt.NewRefreshToken(user, app, a.refreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	return access_token, new_refresh_token, nil
}
