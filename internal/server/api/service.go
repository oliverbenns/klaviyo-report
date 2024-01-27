package api

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/oliverbenns/klaviyo-report/generated/klaviyo"
	redis "github.com/redis/go-redis/v9"
	sloggin "github.com/samber/slog-gin"
)

type Service struct {
	RedisClient   *redis.Client
	Port          int
	Logger        *slog.Logger
	AppURL        string
	ApiKey        string
	KlaviyoClient *klaviyo.ClientWithResponses
}

func (s *Service) Run(ctx context.Context) error {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(sloggin.New(s.Logger))

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true

	router.Use(cors.New(config))
	router.Use(gin.Logger())
	router.Use(s.middleware)

	router.GET("/", s.GetHome)

	router.GET("/ping", s.GetPing)

	addr := fmt.Sprintf(":%d", s.Port)
	router.Run(addr)

	return nil
}

func (s *Service) middleware(c *gin.Context) {
	apiKey := c.Query("api_key")
	if apiKey != s.ApiKey {
		c.AbortWithStatus(401)
		return
	}

	c.Next()
}

func (s *Service) GetHome(c *gin.Context) {
	html := fmt.Sprintf(`
		<html>
			<head>
				<title>Klaviyo Report Prototype</title>
			</head>
			<body>
				<h1>Klaviyo Report Prototype</h1>
				<p></p>
				<hr />
				<a href="https://github.com/oliverbenns/klaviyo-report" target="_blank">Source code</a>
			</body>
		</html>
	`)

	c.Writer.WriteString(html)

	c.Status(200)
	return
}

func (s *Service) GetPing(c *gin.Context) {
	c.PureJSON(200, gin.H{
		"message": "pong",
	})
}
