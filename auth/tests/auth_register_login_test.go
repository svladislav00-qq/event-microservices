package tests

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/stretchr/testify/require"
	"github.com/svladislav00-qq/event-microservices/auth/pb"
	"github.com/svladislav00-qq/event-microservices/auth/tests/suite"
)

const (
	passDefaultLen = 8
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()
	username := gofakeit.Username()

	respReq, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: password,
		Username: username,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReq.GetAccount().Id)

	respLogin, err := st.AuthClient.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: password,
	})

	require.NoError(t, err)

	// loginTime := time.Now().UTC()

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(st.Cfg.JWTSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, respReq.Account.Id, claims["user_id"])
	assert.Equal(t, respReq.Account.Username, claims["username"])
	assert.Equal(t, "user", claims["role"])

	loginTime := time.Now().UTC()

	exp := int64(claims["exp"].(float64))

	assert.InDelta(
		t,
		loginTime.Add(st.Cfg.TokenTTL).Unix(),
		exp,
		2,
	)
}

func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	username := gofakeit.Username()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: pass,
		Username: username,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.Account.Id)

	respReg, err = st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: pass,
		Username: username,
	})
	require.Error(t, err)
	assert.Nil(t, respReg)
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_failCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		username    string
		expectedErr string
	}{
		{
			name:        "Register with empty Password",
			email:       gofakeit.Email(),
			password:    "",
			username:    gofakeit.Username(),
			expectedErr: "password is required",
		},
		{
			name:        "Register with empty Email",
			email:       "",
			password:    randomFakePassword(),
			username:    gofakeit.Username(),
			expectedErr: "email is required",
		},
		{
			name:        "Register with empty Username",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			username:    "",
			expectedErr: "username is required",
		},
		{
			name:        "Register with all empty",
			email:       "",
			password:    "",
			username:    "",
			expectedErr: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
				Username: tt.username,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestLogin_failCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Login with empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "Login with empty Email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "email is required",
		},
		{
			name:        "Login with both empty",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
		{
			name:        "Login with non-matching Password",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			expectedErr: "invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
				Email:    gofakeit.Email(),
				Password: randomFakePassword(),
				Username: gofakeit.Username(),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &pb.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, false, false, passDefaultLen)
}
