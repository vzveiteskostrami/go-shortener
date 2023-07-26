package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	"github.com/vzveiteskostrami/go-shortener/internal/shorturl"
)

func Test_setLink(t *testing.T) {
	//config.ReadData()
	//surl.SetURLNum(dbf.DBFInit())
	//defer dbf.DBFClose()
	dbf.MakeStorage()

	tests := []struct {
		method       string
		url          string
		expectedCode int
		body         string
		expectedBody string
	}{
		//method: http.MethodPut, url: "/", expectedCode: http.StatusBadRequest, body: "", expectedBody: "Ожидался метод " + http.MethodGet},
		//{method: http.MethodDelete, url: "/", expectedCode: http.StatusBadRequest, body: "", expectedBody: "Ожидался метод " + http.MethodGet},
		//{method: http.MethodPost, url: "/", expectedCode: http.StatusBadRequest, body: "www.yandex.ru", expectedBody: "Ожидался метод " + http.MethodGet},
		{method: http.MethodGet, url: "/", expectedCode: http.StatusCreated, body: "www.yandex.ru", expectedBody: ""},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			s := strings.NewReader(tc.body)
			r := httptest.NewRequest(tc.method, tc.url, s)
			w := httptest.NewRecorder()
			var owner int64 = 0

			h := shorturl.SetLink()
			c := context.WithValue(context.WithValue(r.Context(), auth.CPownerID, owner), auth.CPownerValid, true)

			h.ServeHTTP(w, r.WithContext(c))

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, strings.Trim(w.Body.String(), " \n\r"), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}

func Test_getLink(t *testing.T) {
	//config.ReadData()
	///surl.SetURLNum(dbf.DBFInit())
	//defer dbf.DBFClose()
	dbf.MakeStorage()
	tests := []struct {
		method       string
		url          string
		expectedCode int
		body         string
		expectedBody string
	}{
		//{method: http.MethodGet, url: "/1234", expectedCode: http.StatusBadRequest, body: "", expectedBody: `Ожидался метод ` + http.MethodPost},
		//{method: http.MethodPut, url: "/", expectedCode: http.StatusBadRequest, body: "", expectedBody: `Ожидался метод ` + http.MethodPost},
		//{method: http.MethodDelete, url: "/", expectedCode: http.StatusBadRequest, body: "", expectedBody: `Ожидался метод ` + http.MethodPost},
		{method: http.MethodPost, url: "/", expectedCode: http.StatusBadRequest, body: "www.yandex.ru", expectedBody: "Не найден shortURL"},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			s := strings.NewReader(tc.body)
			r := httptest.NewRequest(tc.method, tc.url, s)
			w := httptest.NewRecorder()

			h := shorturl.GetLink()
			h.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, strings.Trim(w.Body.String(), " \n\r"), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}

func Test_setJSONLink(t *testing.T) {
	//config.ReadData()
	//surl.SetURLNum(dbf.DBFInit())
	//defer dbf.DBFClose()
	dbf.MakeStorage()
	tests := []struct {
		method       string
		url          string
		expectedCode int
		body         string
		expectedBody string
	}{
		//{method: http.MethodGet, url: "/1234", expectedCode: http.StatusBadRequest, body: "", expectedBody: `Ожидался метод ` + http.MethodPost},
		//{method: http.MethodPut, url: "/", expectedCode: http.StatusBadRequest, body: "", expectedBody: `Ожидался метод ` + http.MethodPost},
		//{method: http.MethodDelete, url: "/", expectedCode: http.StatusBadRequest, body: "", expectedBody: `Ожидался метод ` + http.MethodPost},
		//{method: http.MethodPost, url: "/api/shorten", expectedCode: http.StatusCreated, body: `{"url": "www.yandex.ru"}`, expectedBody: `{"result":"http://127.0.0.1:8080/0"}`},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			s := strings.NewReader(tc.body)
			r := httptest.NewRequest(tc.method, tc.url, s)
			w := httptest.NewRecorder()

			h := shorturl.SetJSONLink()
			h.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, strings.Trim(w.Body.String(), " \n\r"), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}
