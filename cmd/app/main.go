package main

import (
	"github.com/0xForked/sse-r2g/config"
	"github.com/0xForked/sse-r2g/internal"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile(".env")
	config.LoadEnv().
		InitRedisConnection().
		InitGinEngine()
	internal.StartServer()
}
