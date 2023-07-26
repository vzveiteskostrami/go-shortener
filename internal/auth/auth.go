package auth

import (
	"context"
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

const SECRET_KEY = "pomidoryichesnok"

var (
	CP_ownerID    ContextParamName = "OwnerID"
	CP_ownerValid ContextParamName = "OwnerValid"
	NextOWNERID   int64            = 0
	lockMakeToken sync.Mutex
)

func AuthHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ownerID int64 = 0
		var token string
		var ok bool
		var ownerValid bool = true

		cu, err := r.Cookie("token")

		if err != nil {
			ownerValid = false
			token, ownerID, err = makeToken()
		} else if ownerID, ok = getOwnerID(cu.Value); !ok {
			ownerValid = false
			token, ownerID, err = makeToken()
		}

		if err != nil {
			logging.S().Panic(err)
		}

		if token != "" {
			http.SetCookie(w, &http.Cookie{Name: "token", Value: token, HttpOnly: true})
		}

		c := context.WithValue(context.WithValue(r.Context(), CP_ownerID, ownerID), CP_ownerValid, ownerValid)

		next.ServeHTTP(w, r.WithContext(c))
	})
}

func getOwnerID(tokenString string) (int64, bool) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(SECRET_KEY), nil
	})

	if err != nil {
		logging.S().Errorw(err.Error())
		return -1, false
	}

	if !token.Valid {
		logging.S().Errorw("Token is not valid: " + tokenString)
		return -1, false
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.OwnerID, true
}

func makeToken() (string, int64, error) {
	lockMakeToken.Lock()
	defer lockMakeToken.Unlock()
	n := NextOWNERID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		OwnerID: NextOWNERID,
	})
	NextOWNERID++

	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", -1, err
	}

	return tokenString, n, nil
}