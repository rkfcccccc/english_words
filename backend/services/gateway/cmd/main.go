package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rkfcccccc/english_words/services/gateway/internal/handler"
	"github.com/rkfcccccc/english_words/services/gateway/internal/service"
	"github.com/rkfcccccc/english_words/services/gateway/pkg/auth"
	"github.com/rkfcccccc/english_words/services/gateway/pkg/server"
	"github.com/rkfcccccc/english_words/shared_pkg/cache/redcache"
	"github.com/rkfcccccc/english_words/shared_pkg/redis"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	redis := redis.NewClient(os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	cache := redcache.NewCacheRepository(redis)

	authHelper := auth.NewHelper(os.Getenv("JWT_KEY"), cache)

	services := service.NewService()
	handlers := handler.NewHandlers(services, authHelper)

	router := gin.Default()
	router.Use(gin.Recovery())

	apiGroup := router.Group("/api")

	userGroup := apiGroup.Group("/user")
	userGroup.POST("/signup", handlers.UserSignup)
	userGroup.POST("/login", handlers.UserLogin)
	userGroup.POST("/refresh", handlers.UserRefresh)
	// TODO: userGroup.POST("/recovery", handlers.UserRecovery)

	// authorized := router.Group("/", authorizedHandler)

	// movieGroup := authorized.Group("/movies")
	// movieGroup.GET("/:id") - get info about :id
	// movieGroup.UPDATE("/:id/favorite") - make movie :id unfavorite
	// movieGroup.DELETE("/:id/favorite") - add :id favorite
	// movieGroup.GET("/") - search for movie

	// vocabularyGroup := authorized.Group("/vocabulary")
	// vocabularyGroup.GET("/challenge") - get a challenge to show
	// vocabularyGroup.PATCH("/challenge") - submit the challenge result

	server := server.NewServer(router)
	server.Run()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")

}
