package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Dev-AustinPeter/url-shortner-go/db"
	"github.com/Dev-AustinPeter/url-shortner-go/db/repository"
	"github.com/Dev-AustinPeter/url-shortner-go/handler/urlshortner"
	"github.com/Dev-AustinPeter/url-shortner-go/middleware"
	"github.com/Dev-AustinPeter/url-shortner-go/services/cachemanager"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
)

type APIServer struct {
	addr string
}

func NewAPIServer(addr string) *APIServer {
	return &APIServer{
		addr: addr,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	// Set up CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Change this to specific origins in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Apply CORS Middleware
	handler := corsMiddleware.Handler(router)

	subrouter := router.PathPrefix("/api/v1").Subrouter()

	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	rateLimiter := middleware.NewRateLimiter(1*time.Second, 5*time.Minute, &logger)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go func() {
		<-stop
		rateLimiter.StopCleanup() // Stop background cleanup
		log.Println("Server shutting down...")
		os.Exit(0)
	}()

	db := db.NewConnection("localhost", "5432", "postgres", "", "url_shortner_go", "postgres")
	if db == nil {
		logger.Error().Msg("failed to connect to database")
		return fmt.Errorf("%s", "failed to connect to database")
	}

	defer db.DB.Close()

	if err := db.DB.Ping(); err != nil {
		fmt.Println("failed to connect to database")
		logger.Error().Msg("failed to connect to database")
		return err
	}

	// repository : SQL query are written here
	repository := repository.NewRepository(db)

	// cacheManager : Redis cache
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Error().Err(err).Msg("failed to connect to redis")
		return err
	}

	cacheManager := cachemanager.NewCacheManager(redisClient, logger)

	// handler : API routes are written here
	// 1. shorten : POST /api/v1/shorten
	// 2. getShorten : GET /api/v1/shorten/{shortUrl}
	// 3. createTaskId : GET /api/v1/shorten
	// 4. getTaskBaseOnTaskId : GET /api/v1/task/{taskId}
	shortUrlHandler := urlshortner.NewHandler(repository, &logger, cacheManager)
	shortUrlHandler.RegisterRoutes(subrouter, rateLimiter)

	log.Println("[INFO]: Listening on port", s.addr)
	return http.ListenAndServe(s.addr, handler)
}

func main() {
	// Initialize the application and run it
	app := NewAPIServer(":8080")
	app.Run()
}
