package api

import (
	"embed"
	"fmt"
	"html/template"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/oliverbenns/klaviyo-report/generated/klaviyo"
)

//go:embed home.html
var homeContent embed.FS

type HomeTemplateAccount struct {
	URL  string
	Name string
}

type HomeTemplateData struct {
	Accounts []HomeTemplateAccount
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

	// Even though we are listing and linking to accounts we are only supporting
	// 1 API key at the minute which is at the account level therefore this
	// kind of doesn't make too much sense as only 1 of the reports is going to work.
	homeAccounts := []HomeTemplateAccount{}
	for _, account := range res.JSON200.Data {
		reportURLStr := fmt.Sprintf("/reports/%s", account.Id)
		reportURL, _ := url.Parse(reportURLStr)
		q := url.Values{}
		q.Set("api_key", s.ApiKey)
		reportURL.RawQuery = q.Encode()

		homeAccounts = append(homeAccounts, HomeTemplateAccount{
			Name: account.Attributes.ContactInformation.OrganizationName,
			URL:  reportURL.String(),
		})
	}

	tmpl, err := template.ParseFS(homeContent, "home.html")
	if err != nil {
		s.Logger.Error("failed to parse template", "error", err)
		c.Status(500)
		return
	}

	err = tmpl.Execute(c.Writer, HomeTemplateData{
		Accounts: homeAccounts,
	})
	if err != nil {
		s.Logger.Error("failed to execute template", "error", err)
		c.Status(500)
		return
	}

	c.Status(200)
	return
}
