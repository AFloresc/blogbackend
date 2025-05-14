package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 🧪 **Test: Cargar datos desde JSON**
func TestLoadData(t *testing.T) {
	articles, err := loadData()
	if err != nil {
		t.Errorf("Error loading data: %v", err)
	}
	if len(articles) == 0 {
		t.Errorf("Expected articles, got empty list")
	}
}

// 🧪 **Test: Obtener artículos con paginación**
func TestGetArticles(t *testing.T) {
	req, _ := http.NewRequest("GET", "/articles?page=1&limit=2", nil)
	rr := httptest.NewRecorder()

	getArticles(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}

	var articles []Article
	json.Unmarshal(rr.Body.Bytes(), &articles)
	if len(articles) > 2 {
		t.Errorf("Pagination error: Expected max 2 articles, got %d", len(articles))
	}
}

// 🧪 **Test: Obtener artículo específico**
func TestGetArticle(t *testing.T) {
	req, _ := http.NewRequest("GET", "/article/1", nil)
	rr := httptest.NewRecorder()

	getArticle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}
}

// 🧪 **Test: Agregar un artículo**
func TestAddArticle(t *testing.T) {
	newArticle := Article{Title: "Test", Content: "Content", Date: "2025-05-14", Image: "test.jpg"}
	body, _ := json.Marshal(newArticle)

	req, _ := http.NewRequest("POST", "/admin/add", bytes.NewBuffer(body))
	req.SetBasicAuth("admin", "password")
	rr := httptest.NewRecorder()

	addArticle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}
}

// 🧪 **Test: Editar un artículo**
func TestEditArticle(t *testing.T) {
	updatedArticle := Article{Title: "Updated", Content: "New Content", Date: "2025-06-01", Image: "updated.jpg"}
	body, _ := json.Marshal(updatedArticle)

	req, _ := http.NewRequest("PUT", "/admin/edit/1", bytes.NewBuffer(body))
	req.SetBasicAuth("admin", "password")
	rr := httptest.NewRecorder()

	editArticle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}
}

// 🧪 **Test: Eliminar un artículo**
func TestDeleteArticle(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/admin/delete/1", nil)
	req.SetBasicAuth("admin", "password")
	rr := httptest.NewRecorder()

	deleteArticle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}
}

// 🧪 **Test: Autenticación de usuario**
func TestAuthenticateUser(t *testing.T) {
	valid := authenticateUser("admin", "password")
	if !valid {
		t.Errorf("Expected valid credentials")
	}

	invalid := authenticateUser("fakeuser", "fakepassword")
	if invalid {
		t.Errorf("Expected authentication to fail")
	}
}
