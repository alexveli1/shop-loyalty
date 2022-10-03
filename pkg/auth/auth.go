package auth

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"github.com/alexveli/diploma/internal/config"
	"github.com/alexveli/diploma/internal/domain"
	mylog "github.com/alexveli/diploma/pkg/log"
)

type Manager struct {
	signingKey      string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewManager(cfg config.JWTConfig) (*Manager, error) {
	if cfg == (config.JWTConfig{}) {
		return nil, errors.New("empty jwt configuration provided")
	}
	return &Manager{
		signingKey:      cfg.SigningKey,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}, nil
}

type Claims struct {
	jwt.StandardClaims
	Username string `json:"username"`
}

func (m *Manager) GenerateToken(userID int64) (string, error) {

	claims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(m.accessTokenTTL).Unix(),
		Subject:   fmt.Sprint(userID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(m.signingKey))

}

func (m *Manager) TokenValid(c *gin.Context) error {
	tokenString := m.ExtractToken(c)
	claims := jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {

			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.signingKey), nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) ExtractToken(c *gin.Context) string {
	token := c.Query("token")
	mylog.SugarLogger.Infof("token is %s", token)
	if token != "" {

		return token
	}
	fmt.Println(c.Request.Header)
	bearerToken := c.Request.Header.Get("Authorization")
	if bearerToken != "" {
		if strings.Contains(bearerToken, " ") {
			if len(strings.Split(bearerToken, " ")) == 2 {

				return strings.Split(bearerToken, " ")[1]
			}
		}
	}
	return bearerToken
}

func (m *Manager) ExtractUserIDFromToken(c *gin.Context) (int64, error) {
	tokenString := m.ExtractToken(c)
	claims := jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {

			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.signingKey), nil
	})
	if err != nil {
		return 0, err
	}
	if token.Valid {
		userid, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			mylog.SugarLogger.Errorf("cannot convert userid to int64, %v", err)

			return 0, err
		}

		return userid, nil
	}
	mylog.SugarLogger.Infof("token not valid")
	return 0, domain.ErrAuthorizationInvalidToken
}
