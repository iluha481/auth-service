package tests

import (
	"sso/tests/suite"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	ssov1 "github.com/iluha481/protos/gen/go/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	appID          = 1
	appSecret      = "secret"
	refreshSecret  = "refresh_secret"
	passDefaultLen = 10
)

func TestRegisterLogin_Login_refresh_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	// Регистрация пользователя
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	require.NotEmpty(t, respReg.GetUserId())

	// Логин пользователя
	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appID,
	})
	require.NoError(t, err)

	// Проверка access token
	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	loginTime := time.Now()

	// Парсинг access token
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Проверка содержимого access token
	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)

	// Проверка refresh token
	refreshToken := respLogin.GetRefreshToken()
	require.NotEmpty(t, refreshToken)

	// Парсинг refresh token
	refreshTokenParsed, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(refreshSecret), nil
	})
	require.NoError(t, err)

	refreshClaims, ok := refreshTokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Проверка содержимого refresh token
	assert.Equal(t, respReg.GetUserId(), int64(refreshClaims["uid"].(float64)))
	assert.Equal(t, email, refreshClaims["email"].(string))
	assert.Equal(t, appID, int(refreshClaims["app_id"].(float64)))
	assert.InDelta(t, loginTime.Add(st.Cfg.RefreshTokenTTL).Unix(), refreshClaims["exp"].(float64), deltaSeconds)

	// Обновление токенов
	respRefresh, err := st.AuthClient.RefreshToken(ctx, &ssov1.RefreshRequest{
		RefreshToken: refreshToken,
		AppId:        appID,
	})
	require.NoError(t, err)
	require.NotNil(t, respRefresh)

	// Проверка нового access token
	refreshedToken := respRefresh.GetToken()
	require.NotEmpty(t, refreshedToken)

	refreshedTokenParsed, err := jwt.Parse(refreshedToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	refreshedClaims, ok := refreshedTokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Проверка содержимого нового access token
	assert.Equal(t, respReg.GetUserId(), int64(refreshedClaims["uid"].(float64)))
	assert.Equal(t, email, refreshedClaims["email"].(string))
	assert.Equal(t, appID, int(refreshedClaims["app_id"].(float64)))
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), refreshedClaims["exp"].(float64), deltaSeconds)

	// Проверка нового refresh token
	newRefreshToken := respRefresh.GetRefreshToken()
	require.NotEmpty(t, newRefreshToken)

	newRefreshTokenParsed, err := jwt.Parse(newRefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(refreshSecret), nil
	})
	require.NoError(t, err)

	newRefreshClaims, ok := newRefreshTokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Проверка содержимого нового refresh token
	assert.Equal(t, respReg.GetUserId(), int64(newRefreshClaims["uid"].(float64)))
	assert.Equal(t, email, newRefreshClaims["email"].(string))
	assert.Equal(t, appID, int(newRefreshClaims["app_id"].(float64)))
	assert.InDelta(t, loginTime.Add(st.Cfg.RefreshTokenTTL).Unix(), newRefreshClaims["exp"].(float64), deltaSeconds)
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}

func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	require.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")

}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "Regster with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "email is required",
		},
		{
			name:        "Register with Both Empty",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)

		})

	}

}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appID       int32
		expectedErr string
	}{
		{
			name:        "Login with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			appID:       appID,
			expectedErr: "password is required",
		},
		{
			name:        "Login with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "email is required",
		},

		{
			name:        "Login with Both Empty Email and Password",
			email:       "",
			password:    "",
			appID:       appID,
			expectedErr: "email is required",
		},
		{
			name:        "Login with Non-Matching Password",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "invalid email or password",
		},
		{
			name:  "login withoutAppID",
			email: gofakeit.Email(), password: randomFakePassword(), appID: 0, expectedErr: "app_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    gofakeit.Email(),
				Password: randomFakePassword(),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})

	}

}
