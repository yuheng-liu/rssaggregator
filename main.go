package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/yuheng-liu/rssaggregator/internal/database"

	_ "github.com/lib/pq" // need to import this for it's side effects
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	// retrieve environment variables and store into appConfig
	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}
	// retrieve database url for accessing database
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("dbURL environment variable is not set")
	}
	// connect to the database if no errors
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Can't connecting to postgres database")
	}
	// create apiCfg struct
	apiCfg := apiConfig{
		DB: database.New(conn),
	}
	// create routers
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	// create v1 namespace router
	v1Router := chi.NewRouter()
	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerError)
	v1Router.Post("/users", apiCfg.handlerCreateUser)

	// mount subRouters, create & start server
	router.Mount("/v1", v1Router)
	server := &http.Server{
		Handler: router,
		Addr:    ":" + port,
	}

	log.Printf("Server starting on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
