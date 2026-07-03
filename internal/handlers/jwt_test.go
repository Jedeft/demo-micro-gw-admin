package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jedeft/demo-micro-gw-admin/internal/configs"
	"github.com/Jedeft/demo-micro-gw-admin/internal/testutil"
)

func TestCreateToken_Success(t *testing.T) {
	testutil.SetTestConfig()
	claims := &AuthPlyLoad{
		User: &UserInfo{ID: 1, Username: "admin", Name: "Admin"},
	}
	token, err := CreateToken(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotZero(t, claims.ExpiresAt)
}

func TestParseToken_Success(t *testing.T) {
	testutil.SetTestConfig()
	mr := testutil.StartRedis(t)

	claims := &AuthPlyLoad{
		User: &UserInfo{ID: 1, Username: "admin", Name: "Admin"},
	}
	token, err := CreateToken(claims)
	require.NoError(t, err)

	mr.Set("admin", "true")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	result, err := ParseToken(token, c)
	require.NoError(t, err)
	assert.NotNil(t, result)

	user, ok := c.Get(ParsedUserKey).(*UserInfo)
	require.True(t, ok)
	assert.Equal(t, uint32(1), user.ID)
	assert.Equal(t, "admin", user.Username)
}

func TestParseToken_InvalidToken(t *testing.T) {
	testutil.SetTestConfig()
	testutil.StartRedis(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	result, err := ParseToken("invalid-token-string", c)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestParseToken_NotInRedis(t *testing.T) {
	testutil.SetTestConfig()
	testutil.StartRedis(t)

	claims := &AuthPlyLoad{
		User: &UserInfo{ID: 1, Username: "ghost", Name: "Ghost"},
	}
	token, err := CreateToken(claims)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	result, err := ParseToken(token, c)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestInvalidJWT(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := InvalidJWT(nil, c)
	assert.Error(t, err)
	var he *echo.HTTPError
	require.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusUnauthorized, he.Code)
}

func TestAuthSkipper_FilterToken(t *testing.T) {
	testutil.SetTestConfig()
	configsSetFilterToken(true)
	defer configsSetFilterToken(false)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/user", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/v1/user")

	skip := AuthSkipper(c)
	assert.True(t, skip)

	user, ok := c.Get(ParsedUserKey).(*UserInfo)
	require.True(t, ok)
	assert.Equal(t, uint32(1), user.ID)
}

func TestAuthSkipper_HealthCheck(t *testing.T) {
	testutil.SetTestConfig()
	configsSetFilterToken(false)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health/check", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/health/check")

	assert.True(t, AuthSkipper(c))
}

func TestAuthSkipper_Login(t *testing.T) {
	testutil.SetTestConfig()
	configsSetFilterToken(false)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/v1/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/v1/login")

	assert.True(t, AuthSkipper(c))
}

func TestAuthSkipper_OtherPath(t *testing.T) {
	testutil.SetTestConfig()
	configsSetFilterToken(false)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/user", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/v1/user")

	assert.False(t, AuthSkipper(c))
}

func TestGetUserInfo(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expected := &UserInfo{ID: 42, Username: "test", Name: "Test"}
	c.Set(ParsedUserKey, expected)

	got := GetUserInfo(c)
	assert.Equal(t, expected, got)
}

// configsSetFilterToken 设置 IsFilterToken 配置项。
func configsSetFilterToken(v bool) {
	configs.Config.Debug.IsFilterToken = v
}
