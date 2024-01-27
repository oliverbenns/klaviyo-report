package api

import (
	"html/template"

	"embed"

	"github.com/gin-gonic/gin"
	"github.com/oliverbenns/klaviyo-report/generated/klaviyo"
)

//go:embed report.html
var reportContent embed.FS

type KlaviyoReportTemplateData struct {
	AccountName string
}

func (s *Service) GetKlaviyoReport(c *gin.Context) {
	klaviyoAccountID := c.Param("klaviyo_account_id")
	if klaviyoAccountID == "" {
		s.Logger.Warn("klaviyo_account_id is required")
		c.Status(400)
		return
	}

	res, err := s.KlaviyoClient.GetAccountWithResponse(c.Request.Context(), klaviyoAccountID, &klaviyo.GetAccountParams{
		Revision: "2023-12-15",
	})
	if err != nil {
		s.Logger.Error("failed to get account", "error", err)
		c.Status(500)
		return
	}

	tmpl, err := template.ParseFS(reportContent, "report.html")
	if err != nil {
		s.Logger.Error("failed to parse template", "error", err)
		c.Status(500)
		return
	}

	err = tmpl.Execute(c.Writer, KlaviyoReportTemplateData{
		AccountName: res.JSON200.Data.Attributes.ContactInformation.OrganizationName,
	})
	if err != nil {
		s.Logger.Error("failed to execute template", "error", err)
		c.Status(500)
		return
	}

	c.Status(200)
	return
}
