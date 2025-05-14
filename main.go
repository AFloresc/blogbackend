package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// Data structures
type Article struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Date    string `json:"date"`
	Image   string `json:"image"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Read articles from JSON
func loadData() ([]Article, error) {
	file, err := os.Open("data.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var articles []Article
	err = json.NewDecoder(file).Decode(&articles)
	return articles, err
}

// Write articles to JSON
func saveData(articles []Article) error {
	file, err := os.Create("data.json")
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(articles)
}

// API Endpoints
func getArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := loadData()
	if err != nil {
		http.Error(w, "Error loading articles", http.StatusInternalServerError)
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 5 // Default limit
	}

	start := (page - 1) * limit
	end := start + limit
	if start >= len(articles) {
		json.NewEncoder(w).Encode([]Article{}) // Return empty array if page out of bounds
		return
	}

	if end > len(articles) {
		end = len(articles)
	}

	json.NewEncoder(w).Encode(articles[start:end])
}

func getArticle(w http.ResponseWriter, r *http.Request) {
	articles, _ := loadData()
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	for _, article := range articles {
		if article.ID == id {
			json.NewEncoder(w).Encode(article)
			return
		}
	}
	http.Error(w, "Article not found", http.StatusNotFound)
}

// Authentication (Basic Auth)

func authenticateUser(username, password string) bool {
	file, err := os.Open("users.json")
	if err != nil {
		fmt.Println("Error opening users file:", err)
		return false
	}
	defer file.Close()

	var users []User
	json.NewDecoder(file).Decode(&users)

	for _, user := range users {
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err == nil {
			return true // Password matches
		}
	}
	return false
}

// Protected Admin Routes
func adminHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Fprint(w, "Welcome Admin!")
}

func editArticle(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()
	if !authenticateUser(username, password) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	articles, _ := loadData()
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	var updatedArticle Article
	err := json.NewDecoder(r.Body).Decode(&updatedArticle)
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	for i, article := range articles {
		if article.ID == id {
			articles[i] = updatedArticle
			articles[i].ID = id // Keep the original ID
			saveData(articles)
			json.NewEncoder(w).Encode(articles[i])
			return
		}
	}

	http.Error(w, "Article not found", http.StatusNotFound)
}

func addArticle(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()
	if !authenticateUser(username, password) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var newArticle Article
	err := json.NewDecoder(r.Body).Decode(&newArticle)
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	articles, _ := loadData()
	newArticle.ID = len(articles) + 1 // Assign a new ID
	articles = append(articles, newArticle)

	saveData(articles)
	json.NewEncoder(w).Encode(newArticle)
}

func deleteArticle(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()
	if !authenticateUser(username, password) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	articles, _ := loadData()
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	for i, article := range articles {
		if article.ID == id {
			articles = append(articles[:i], articles[i+1:]...)
			saveData(articles)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	http.Error(w, "Article not found", http.StatusNotFound)
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

var store = sessions.NewCookieStore([]byte("super-secret-key"))

func loginHandler(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()

	if authenticateUser(username, password) {
		session, _ := store.Get(r, "session-name")
		session.Values["authenticated"] = true
		session.Save(r, w)

		json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func isAuthenticated(r *http.Request) bool {
	session, _ := store.Get(r, "session-name")
	auth, ok := session.Values["authenticated"].(bool)
	return ok && auth
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Save(r, w)

	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out"})
}

// Main function
func main() {
	router := mux.NewRouter()

	router.HandleFunc("/articles", getArticles).Methods("GET")
	router.HandleFunc("/article/{id}", getArticle).Methods("GET")
	router.HandleFunc("/admin", adminHandler).Methods("GET")
	router.HandleFunc("/admin/add", addArticle).Methods("POST")
	router.HandleFunc("/admin/edit/{id}", editArticle).Methods("PUT")
	router.HandleFunc("/admin/delete/{id}", deleteArticle).Methods("DELETE")
	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/logout", logoutHandler).Methods("POST")

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", router)
}
