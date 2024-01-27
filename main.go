package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v4"
	"github.com/naeem4265/api-server/handlers"
)

func main() {

	// create a database object which can be used to connect with database.
	db, err := sql.Open("mysql", "root:@tcp(0.0.0.0:3306)/BookServer")
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Now its  time to connect with oru database, database object has a method Ping.
	// Ping returns error, if unable connect to database.
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Print("Database Connected\n")

	router := chi.NewRouter()

	// Use a closure to capture the 'db' object and pass it to the handler.
	router.Post("/signin", func(w http.ResponseWriter, r *http.Request) {
		handlers.SignIn(w, r, db)
	})
	router.Get("/signout", handlers.SignOut)

	router.Route("/albums", func(r chi.Router) {
		r.Use(middleware)
		r.Get("/", handlers.GetAlbums)
		r.Get("/{id}", handlers.GetAlbumById)
		r.Put("/{id}", handlers.PutAlbum)
		r.Post("/", handlers.PostAlbum)
		r.Delete("/{id}", handlers.DeleteAlbum)
	})

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for the "token" cookie
		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tknStr := c.Value

		claims := &handlers.Claims{}
		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return handlers.JWTKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				// Token signature is invalid, return unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// For any other error while parsing claims, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !tkn.Valid {
			// Token is not valid, return unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// If token is valid, continue to the next handler
		next.ServeHTTP(w, r)
	})
}
