package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v4"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
)

// 0 = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJPd25lcklEIjowfQ.u6d3Bcz7A-MulX5WbdBJypc56uRF2DOILD_WxqOsvOk
// 1 = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJPd25lcklEIjoxfQ.cOg2cbX9qBBQUH1yqvNIgMWX-w-PnXdPxr5tbmXg4fw

type ContextParamName string

type Claims struct {
	jwt.RegisteredClaims
	OwnerID int64
}

const SecretKey = "pomidoryichesnok"

type tokenServiceStruct struct {
	NewOWNERID int64
	locker     sync.Mutex
}

var (
	CPownerID    ContextParamName = "OwnerID"
	CPownerValid ContextParamName = "OwnerValid"
	TokenService tokenServiceStruct
)

func AuthHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ownerID int64 = 0
		var token string
		var err error
		ownerValid := true

		cu, err := r.Cookie("token")

		if err != nil {
			ownerValid = false
			token, ownerID, err = TokenService.MakeToken()
		} else if ownerID, err = GetOwnerID(cu.Value); err != nil {
			ownerValid = false
			token, ownerID, err = TokenService.MakeToken()
		}

		if err != nil {
			logging.S().Error(err)
			return
		}

		if token != "" {
			http.SetCookie(w, &http.Cookie{Name: "token", Value: token, HttpOnly: true})
		}

		c := context.WithValue(context.WithValue(r.Context(), CPownerID, ownerID), CPownerValid, ownerValid)

		next.ServeHTTP(w, r.WithContext(c))
	})
}

func GetOwnerID(tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

	if err != nil {
		logging.S().Errorw(err.Error())
		return 0, err
	}

	if !token.Valid {
		err := errors.New("Token is not valid: " + tokenString)
		logging.S().Errorw(err.Error())
		return 0, err
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.OwnerID, nil
}

func (t *tokenServiceStruct) MakeToken() (string, int64, error) {
	t.locker.Lock()
	defer t.locker.Unlock()
	n := t.NewOWNERID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		OwnerID: t.NewOWNERID,
	})
	t.NewOWNERID++

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, n, nil
}

func SetNewOwnerID(n int64) {
	TokenService.NewOWNERID = n
}

func init() {
	TokenService.NewOWNERID = 0
}
