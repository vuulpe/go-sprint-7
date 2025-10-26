package main

import (
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
    handler := http.HandlerFunc(mainHandle)

    requests := []struct {
        request string
        status  int
        message string
    }{
        {"/cafe", http.StatusBadRequest, "unknown city"},
        {"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
        {"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
    }
    for _, v := range requests {
        response := httptest.NewRecorder()
        req := httptest.NewRequest("GET", v.request, nil)
        handler.ServeHTTP(response, req)

        assert.Equal(t, v.status, response.Code)
        assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
    }
}

func TestCafeCount(t *testing.T) {
    requests := []struct {
        count int 
        want  int 
    }{
        {count: 0, want: 0},
        {count: 1, want: 1},
        {count: 2, want: 2},
        {count: 100, want: len(cafeList["moscow"])}, 
    }

    for _, tt := range requests {
        url := fmt.Sprintf("http://localhost:8080/cafe?city=moscow&count=%d", tt.count)
        
        resp, err := http.Get(url)
        require.NoError(t, err, "Ошибка при выполнении запроса")
        defer resp.Body.Close()

        require.Equal(t, http.StatusOK, resp.StatusCode, "Код ответа не 200")

        body, err := io.ReadAll(resp.Body)
        require.NoError(t, err, "Ошибка при чтении ответа")

        if tt.want == 0 {
            assert.Equal(t, "", string(body), "Ожидалась пустая строка")
            continue
        }

        cafes := strings.Split(string(body), ",")
        
        assert.Equal(t, tt.want, len(cafes), "Неверное количество кафе в ответе")
    }
}

func TestCafeSearch(t *testing.T) {
    requests := []struct {
        search    string 
        wantCount int
    }{
        {search: "фасоль", wantCount: 0},
        {search: "кофе", wantCount: 2},
        {search: "вилка", wantCount: 1},
    }

    for _, tt := range requests {
        url := fmt.Sprintf("http://localhost:8080/cafe?city=moscow&search=%s", tt.search)

        resp, err := http.Get(url)
        require.NoError(t, err, "Ошибка при выполнении запроса")
        defer resp.Body.Close()

        require.Equal(t, http.StatusOK, resp.StatusCode, "Код ответа не 200")

        body, err := io.ReadAll(resp.Body)
        require.NoError(t, err, "Ошибка при чтении ответа")

        if tt.wantCount == 0 {
            assert.Equal(t, "", string(body), "Ожидалась пустая строка")
            continue
        }

        cafes := strings.Split(string(body), ",")
        
        assert.Equal(t, tt.wantCount, len(cafes), 
            "Неверное количество найденных кафе для поиска '%s'", tt.search)

        searchLower := strings.ToLower(tt.search)
        for _, cafe := range cafes {
            cafeLower := strings.ToLower(cafe)
            assert.True(t, strings.Contains(cafeLower, searchLower),
                "Кафе '%s' не содержит строку '%s'", cafe, tt.search)
        }
    }
}