package internal

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/0xForked/sse-r2g/config"
	"github.com/0xForked/sse-r2g/web"
	"github.com/gin-gonic/gin"
)

func StartServer() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// router engine & redis connection
	routerEngine := config.GinEngine
	redisConnection := config.RedisConn
	// register router & notification module
	routerEngine.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusTemporaryRedirect, "/fe")
	})
	routerEngine.StaticFS("/fe", http.FS(web.Resource))
	apiRoute := routerEngine.Group("/api/v1")
	NewNotificationProvider(apiRoute, redisConnection)
	// server defines parameters for running an HTTP server.
	tenSec := 10
	server := &http.Server{
		Addr:              config.Instance.ServerPort,
		Handler:           routerEngine,
		ReadHeaderTimeout: time.Second * time.Duration(tenSec),
	}
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(
			err, http.ErrServerClosed,
		) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	// Listen for the interrupt signal.
	<-ctx.Done()
	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")
	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	timeToHandle := 5
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(timeToHandle)*time.Second)
	defer cancel()
	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %s", err)
	}
	// Close redis connections
	if err := redisConnection.Close(); err != nil {
		log.Printf("Error close redis connection: %v", err)
	}
	// notify user of shutdown
	log.Println("Server exiting")
}
