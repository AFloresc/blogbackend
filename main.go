package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

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

// Configuración de sesiones
var store = sessions.NewCookieStore([]byte("super-secret-key"))

func init() {
	store.Options = &sessions.Options{
		MaxAge:   1800, // 30 minutos de sesión activa
		HttpOnly: true,
	}
}

// Cargar artículos desde JSON
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

// Guardar artículos en JSON
func saveData(articles []Article) error {
	file, err := os.Create("data.json")
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(articles)
}

// Obtener lista de artículos con paginación
func getArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := loadData()
	if err != nil {
		http.Error(w, "Error loading articles", http.StatusInternalServerError)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 5 // Default: 5 artículos por página
	}

	start := (page - 1) * limit
	end := start + limit
	if start >= len(articles) {
		json.NewEncoder(w).Encode([]Article{})
		return
	}
	if end > len(articles) {
		end = len(articles)
	}

	json.NewEncoder(w).Encode(articles[start:end])
}

// Obtener un artículo por ID
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

// Autenticación de usuario con bcrypt
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
			return true
		}
	}
	return false
}

// Manejo de inicio de sesión con sesiones
func loginHandler(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()

	if authenticateUser(username, password) {
		session, _ := store.Get(r, "session-name")
		session.Values["authenticated"] = true
		session.Values["lastActive"] = time.Now().Unix()
		session.Save(r, w)

		json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

// Verificación de sesión activa
func isAuthenticated(r *http.Request) bool {
	session, _ := store.Get(r, "session-name")
	auth, ok := session.Values["authenticated"].(bool)
	lastActive, _ := session.Values["lastActive"].(int64)

	// Expiración automática por inactividad
	if ok && auth && time.Now().Unix()-lastActive < 1800 {
		session.Values["lastActive"] = time.Now().Unix() // Actualizar actividad
		session.Save(r, nil)
		return true
	}

	return false
}

// Manejo de cierre de sesión
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Save(r, w)

	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out"})
}

// Agregar un nuevo artículo (solo admin)
func addArticle(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
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
	newArticle.ID = len(articles) + 1
	articles = append(articles, newArticle)

	saveData(articles)
	json.NewEncoder(w).Encode(newArticle)
}

// Editar un artículo existente (solo admin)
func editArticle(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
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
			articles[i].ID = id
			saveData(articles)
			json.NewEncoder(w).Encode(articles[i])
			return
		}
	}

	http.Error(w, "Article not found", http.StatusNotFound)
}

// Eliminar un artículo (solo admin)
func deleteArticle(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
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

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Permitir solicitudes desde cualquier origen
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Configurar rutas
func main() {
	router := mux.NewRouter()

	router.HandleFunc("/articles", getArticles).Methods("GET")
	router.HandleFunc("/article/{id}", getArticle).Methods("GET")

	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/logout", logoutHandler).Methods("POST")

	router.HandleFunc("/admin/add", addArticle).Methods("POST")
	router.HandleFunc("/admin/edit/{id}", editArticle).Methods("PUT")
	router.HandleFunc("/admin/delete/{id}", deleteArticle).Methods("DELETE")

	// Aplicar el middleware CORS
	handler := enableCORS(router)

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", handler)
}
