package config

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var (
	configOnce, redisOnce,
	engineOnce sync.Once
	Instance  *Config
	RedisConn *redis.Client
	GinEngine *gin.Engine
)

type Config struct {
	ServerName  string `mapstructure:"SERVER_NAME"`
	ServerPort  string `mapstructure:"SERVER_PORT"`
	ServerDebug bool   `mapstructure:"SERVER_DEBUG"`
	RedisDsnURL string `mapstructure:"REDIS_DSN_URL"`
}

func LoadEnv() *Config {
	// notify that app try to load config file
	log.Println("Load configuration file . . . .")
	configOnce.Do(func() {
		// find environment file
		viper.AutomaticEnv()
		// error handling for specific case
		if err := viper.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) {
				// Config file not found; ignore error if desired
				log.Fatal(".env file not found!, please copy .env.example and paste as .env")
			}
			log.Fatalf("ENV_ERROR: %s", err.Error())
		}
		// notify that config file is ready
		log.Println("configuration file: ready")
		// extract config to struct
		if err := viper.Unmarshal(&Instance); err != nil {
			log.Fatalf("ENV_ERROR: %s", err.Error())
		}
	})
	return Instance
}

func (cfg *Config) InitRedisConnection() *Config {
	redisOnce.Do(func() {
		log.Println("Trying to open redis connection pool . . . .")
		opts, err := redis.ParseURL(cfg.RedisDsnURL)
		if err != nil {
			log.Fatalf("REDIS_ERROR: %s", err.Error())
		}
		RedisConn = redis.NewClient(opts)
		if err := RedisConn.Ping(context.Background()).Err(); err != nil {
			log.Fatalf("REDIS_ERROR: %s", err.Error())
		}
		log.Println("Redis connection pool created . . . .")
	})
	return cfg
}

func (cfg *Config) InitGinEngine() *Config {
	var (
		allowOrigins = []string{
			"http://localhost:3000",
			"http://localhost:8000",
		}

		allowHeaders = []string{
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
			"Cache-Control",
			"Origin",
			"Cookie",
		}
	)
	engineOnce.Do(func() {
		log.Printf("Trying to init engine (GIN %s) . . . .", gin.Version)
		// set gin mode
		gin.SetMode(func() string {
			if cfg.ServerDebug {
				return gin.DebugMode
			}
			return gin.ReleaseMode
		}())
		// set global variables
		GinEngine = gin.Default()
		// set cors middleware
		GinEngine.Use(func(ctx *gin.Context) {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", strings.Join(allowOrigins, ","))
			ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			ctx.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ","))
			ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
			if ctx.Request.Method == "OPTIONS" {
				ctx.AbortWithStatus(http.StatusNoContent)
				return
			}
			ctx.Next()
		})
	})
	return cfg
}
