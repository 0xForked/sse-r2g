package internal

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type notificationHandler struct {
	redisConn *redis.Client
}

func (handler notificationHandler) fireNotify(ctx *gin.Context) {
	sessionID := ctx.Param("session_id")
	channelKey := fmt.Sprintf("session:%s:notify", sessionID)
	newNotifyID := uuid.New().String()
	publisher := handler.redisConn.Publish(ctx, channelKey, fmt.Sprintf(
		"[%d] New notify: %s", time.Now().Unix(), newNotifyID))
	if err := publisher.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusCreated, "Notify sent . . .")
}

func (handler notificationHandler) streamNotify(ctx *gin.Context) {
	// Get session ID from URL
	sessionID := ctx.Param("session_id")
	fmt.Println("New client has connected", sessionID)
	// Set header
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	// subscribe to redis pub-sub channel
	channelKey := fmt.Sprintf("session:%s:notify", sessionID)
	subscriber := handler.redisConn.Subscribe(ctx, channelKey)
	defer func() { _ = subscriber.Close() }()
	// Send message to client
	for {
		select {
		case msg := <-subscriber.Channel():
			// Write to the client
			ctx.SSEvent("message", msg.Payload)
			// Flush the buffer to client
			ctx.Writer.Flush()
		case <-ctx.Request.Context().Done():
			fmt.Println(
				"The client has disconnected,",
				"so stop sending messages",
				sessionID,
			)
			return
		}
	}
}

func NewNotificationProvider(
	router *gin.RouterGroup,
	redisConn *redis.Client,
) {
	handler := &notificationHandler{redisConn}
	router.GET("/notifications/:session_id/fire", handler.fireNotify)
	router.GET("/notifications/:session_id/stream", handler.streamNotify)
}
