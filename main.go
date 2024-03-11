package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

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
	// retrieve port number from .env file
	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}
	// retrieve database url from .env file for accessing database
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
	db := database.New(conn)
	apiCfg := apiConfig{
		DB: db,
	}
	// starts rss scraping service, will run forever as long as server is up
	go startScraping(db, 10, time.Minute)
	// create routers
	router := chi.NewRouter()
	// Using a standard cors handler for basic http requests
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	// create v1 namespace router
	v1Router := chi.NewRouter()
	// common
	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerError)
	// users
	v1Router.Post("/users", apiCfg.handlerCreateUser)
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
	// feeds
	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)
	// posts
	v1Router.Get("/posts", apiCfg.middlewareAuth(apiCfg.handlerGetPostsForUser))
	// feed_follows
	v1Router.Post("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	v1Router.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	v1Router.Delete("/feed_follows/{feedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))

	// mount subRouters, create & start server
	router.Mount("/v1", v1Router)
	server := &http.Server{
		Handler: router,
		Addr:    ":" + port,
	}
	// ListenAndServe() function will run forever unless some error occurs
	log.Printf("Server starting on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
