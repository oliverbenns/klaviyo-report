package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

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
	router.GET("/reports/:klaviyo_account_id", s.GetKlaviyoReport)

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
	res, err := s.KlaviyoClient.GetAccountsWithResponse(c.Request.Context(), &klaviyo.GetAccountsParams{
		Revision: "2023-12-15",
	})
	if err != nil {
		s.Logger.Error("failed to get accounts", "error", err)
		c.AbortWithStatus(500)
		return
	}

	accountList := "<ul>"
	for _, account := range res.JSON200.Data {
		reportURLStr := fmt.Sprintf("/reports/%s", account.Id)
		reportURL, _ := url.Parse(reportURLStr)
		q := url.Values{}
		q.Set("api_key", s.ApiKey)
		reportURL.RawQuery = q.Encode()
		accountList += fmt.Sprintf("<li><a href=\"%s\" target=\"_blank\">%s</a></li>", reportURL, account.Attributes.ContactInformation.OrganizationName)
	}
	accountList += "</ul>"

	html := fmt.Sprintf(`
		<html>
			<head>
				<title>Klaviyo Report Prototype</title>
			</head>
			<body>
				<h1>Klaviyo Report Prototype</h1>
				<hr />
				<h3>Accounts</h3>
				%s
				<hr />
				<a href="https://github.com/oliverbenns/klaviyo-report" target="_blank">Source code</a>
			</body>
		</html>
	`, accountList)

	c.Writer.WriteString(html)

	c.Status(200)
	return
}

func (s *Service) GetPing(c *gin.Context) {
	c.PureJSON(200, gin.H{
		"message": "pong",
	})
}
