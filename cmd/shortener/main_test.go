package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_setLink(t *testing.T) {
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
		{method: http.MethodGet, url: "/", expectedCode: http.StatusCreated, body: "www.yandex.ru", expectedBody: "http://127.0.0.1:8080/0"},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			s := strings.NewReader(tc.body)
			r := httptest.NewRequest(tc.method, tc.url, s)
			w := httptest.NewRecorder()

			h := setLink()
			h.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, strings.Trim(w.Body.String(), " \n\r"), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}

func Test_getLink(t *testing.T) {
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

			h := getLink()
			h.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, strings.Trim(w.Body.String(), " \n\r"), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}

/*
func Test_main(t *testing.T) {
	tests := []struct {
		method       string
		url          string
		expectedCode int
		body         string
		expectedBody string
	}{
		//{method: http.MethodGet, url: "/1234", expectedCode: http.StatusBadRequest, body: "", expectedBody: `Не найден shortURL 1234`},
		//{method: http.MethodPut, url: "/", expectedCode: http.StatusBadRequest, body: "", expectedBody: ""Ожидался POST или GET"},
		//{method: http.MethodDelete, url: "/", expectedCode: http.StatusBadRequest, body: "", expectedBody: ""Ожидался POST или GET"},
		{method: http.MethodPost, url: "/", expectedCode: http.StatusOK, body: "www.yandex.ru", expectedBody: "http:/127.0.0.1:8080/0"},
		{method: http.MethodPost, url: "/", expectedCode: http.StatusOK, body: "close", expectedBody: "Сервер выключен"},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			//s := strings.NewReader(tc.body)
			client := &http.Client{
				Timeout: time.Second * 1,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					fmt.Println(req.URL)
					return nil
				},
			}
			req, _ := http.NewRequest(
				tc.method, tc.url, strings.NewReader(tc.body),
			)
			// добавляем заголовки
			//req.Header.Add("Accept", "text/html")     // добавляем заголовок Accept
			//req.Header.Add("User-Agent", "MSIE/15.0") // добавляем заголовок User-Agent

			resp, err := client.Do(req)

			assert.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if err == nil {
				defer resp.Body.Close()
			}
		})
	}
}
*/
