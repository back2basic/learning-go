package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/back2basic/learning-go/01-rssagg/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	// feed, err := urlToFeed("https://www.wagslane.dev/index.xml")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(feed)

	godotenv.Load()

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not set")
	}

	dbURL := os.Getenv("DB_URL")
	if portString == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}

	apiCfg := apiConfig{
		DB: database.New(conn),
	}

	// Scraper
	db := database.New(conn)
	go startScraping(db, 10, time.Minute)

	// Router
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		// AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{"Link"},
		MaxAge:         300, // Maximum value not ignored by any of major browsers
	}))

	v1Router := chi.NewRouter()
	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerERror)
	v1Router.Post("/users", apiCfg.handlerCreateUser)
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))

	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds", (apiCfg.handlerGetFeed))

	v1Router.Post("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	v1Router.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollow))
	v1Router.Delete("/feed_follows/{feedFollowId}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))

	v1Router.Get("/posts", apiCfg.middlewareAuth(apiCfg.handlerGetPostsForUsers))
	
	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%s", portString),
	}

	log.Printf("Server starting on port %s\n", portString)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
