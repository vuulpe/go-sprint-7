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
    handler := http.HandlerFunc(mainHandle) // ИСПОЛЬЗУЕМ httptest вместо реального сервера

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
        // Создаем запрос через httptest, а не реальный HTTP
        response := httptest.NewRecorder()
        url := fmt.Sprintf("/cafe?city=moscow&count=%d", tt.count)
        req := httptest.NewRequest("GET", url, nil)
        
        // Выполняем обработчик напрямую
        handler.ServeHTTP(response, req)

        // Проверяем, что запрос успешно обработан
        require.Equal(t, http.StatusOK, response.Code, "Код ответа не 200")

        // Читаем тело ответа
        body, err := io.ReadAll(response.Body)
        require.NoError(t, err, "Ошибка при чтении ответа")

        // Если ожидаем 0 кафе, проверяем пустую строку
        if tt.want == 0 {
            assert.Equal(t, "", string(body), "Ожидалась пустая строка")
            continue
        }

        // Разбиваем ответ на слайс кафе
        cafes := strings.Split(string(body), ",")
        
        // Проверяем количество возвращенных кафе
        assert.Equal(t, tt.want, len(cafes), "Неверное количество кафе в ответе")
    }
}

func TestCafeSearch(t *testing.T) {
    handler := http.HandlerFunc(mainHandle) // ИСПОЛЬЗУЕМ httptest вместо реального сервера

    requests := []struct {
        search    string
        wantCount int
    }{
        {search: "фасоль", wantCount: 0},
        {search: "кофе", wantCount: 2},      // "Мир кофе" и "Кофе и завтраки"
        {search: "вилка", wantCount: 1},     // "Ложка и вилка"
    }

    for _, tt := range requests {
        // Создаем запрос через httptest
        response := httptest.NewRecorder()
        url := fmt.Sprintf("/cafe?city=moscow&search=%s", tt.search)
        req := httptest.NewRequest("GET", url, nil)
        
        // Выполняем обработчик напрямую
        handler.ServeHTTP(response, req)

        // Проверяем, что запрос успешно обработан
        require.Equal(t, http.StatusOK, response.Code, "Код ответа не 200")

        // Читаем тело ответа
        body, err := io.ReadAll(response.Body)
        require.NoError(t, err, "Ошибка при чтении ответа")

        // Если ожидаем 0 кафе, проверяем пустую строку
        if tt.wantCount == 0 {
            assert.Equal(t, "", string(body), "Ожидалась пустая строка")
            continue
        }

        // Разбиваем ответ на слайс кафе
        cafes := strings.Split(string(body), ",")
        
        // Проверяем количество найденных кафе
        assert.Equal(t, tt.wantCount, len(cafes), 
            "Неверное количество найденных кафе для поиска '%s'", tt.search)

        // Проверяем, что каждое кафе содержит искомую строку (без учета регистра)
        searchLower := strings.ToLower(tt.search)
        for _, cafe := range cafes {
            cafeLower := strings.ToLower(cafe)
            assert.True(t, strings.Contains(cafeLower, searchLower),
                "Кафе '%s' не содержит строку '%s'", cafe, tt.search)
        }
    }
}