package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func connectToDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	var db *sql.DB
	var err error

	// Intenta conectarse hasta 5 veces con espera exponencial
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("✅ Conexión a PostgreSQL establecida")
				return db, nil
			}
		}
		waitTime := time.Duration(i*i) * time.Second
		log.Printf("⚠️ Intento %d: Error conectando a PostgreSQL: %v. Reintentando en %v...", i+1, err, waitTime)
		time.Sleep(waitTime)
	}
	return nil, fmt.Errorf("no se pudo conectar a PostgreSQL después de 5 intentos")
}

func main() {
	// Conexión a PostgreSQL con reintentos
	db, err := connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Crear tabla si no existe
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL
		)`); err != nil {
		log.Fatal("Error creando tabla:", err)
	}

	// Configurar router
	router := mux.NewRouter()
	router.Use(jsonContentTypeMiddleware)
	router.Use(enableCORS)

	// Rutas
	router.HandleFunc("/users", getUsers(db)).Methods("GET")
	router.HandleFunc("/users/{id}", getUser(db)).Methods("GET")
	router.HandleFunc("/users", createUser(db)).Methods("POST")
	router.HandleFunc("/users/{id}", updateUser(db)).Methods("PUT")
	router.HandleFunc("/users/{id}", deleteUser(db)).Methods("DELETE")

	// Iniciar servidor
	log.Println("🚀 Servidor iniciado en http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}

// Middlewares
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Handlers
func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, email FROM users")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		users := []User{}
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
				respondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
			users = append(users, u)
		}
		respondWithJSON(w, http.StatusOK, users)
	}
}

func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var u User
		err := db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id).Scan(&u.ID, &u.Name, &u.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusNotFound, "Usuario no encontrado")
			} else {
				respondWithError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		respondWithJSON(w, http.StatusOK, u)
	}
}

func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			respondWithError(w, http.StatusBadRequest, "Datos inválidos")
			return
		}

		if u.Name == "" || u.Email == "" {
			respondWithError(w, http.StatusBadRequest, "Nombre y email son requeridos")
			return
		}

		err := db.QueryRow(
			"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
			u.Name, u.Email,
		).Scan(&u.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondWithJSON(w, http.StatusCreated, u)
	}
}

func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			respondWithError(w, http.StatusBadRequest, "Datos inválidos")
			return
		}

		_, err := db.Exec(
			"UPDATE users SET name = $1, email = $2 WHERE id = $3",
			u.Name, u.Email, id,
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, u)
	}
}

func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]string{"message": "Usuario eliminado"})
	}
}

// Helpers
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.WriteHeader(code)
	w.Write(response)
}

