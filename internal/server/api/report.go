package api

import (
	"context"
	"fmt"
	"html/template"
	"time"

	"embed"

	"github.com/gin-gonic/gin"
	"github.com/oliverbenns/klaviyo-report/generated/klaviyo"
)

//go:embed report.html
var reportContent embed.FS

type KlaviyoReportTemplateCampaign struct {
	Name                string
	TotalRecipients     int
	OrdersPlaced        int
	ConversionRate      float64
	ConversionValue     float64
	RevenuePerRecipient float64
}

type KlaviyoReportTemplateData struct {
	AccountName string
	Campaigns   []KlaviyoReportTemplateCampaign
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

	campaigns, err := s.getKlaviyoReportCampaigns(c.Request.Context())
	if err != nil {
		s.Logger.Error("failed to get campaigns", "error", err)
		c.Status(500)
		return
	}

	// Create a template with the custom function
	funcMap := template.FuncMap{
		"formatPercent": formatPercent,
		"formatCcy":     formatCcy,
	}

	tmpl, err := template.New("report.html").Funcs(funcMap).ParseFS(reportContent, "report.html")
	if err != nil {
		s.Logger.Error("failed to parse template", "error", err)
		c.Status(500)
		return
	}

	err = tmpl.Execute(c.Writer, KlaviyoReportTemplateData{
		AccountName: res.JSON200.Data.Attributes.ContactInformation.OrganizationName,
		Campaigns:   campaigns,
	})
	if err != nil {
		s.Logger.Error("failed to execute template", "error", err)
		c.Status(500)
		return
	}

	c.Status(200)
	return
}

func formatCcy(value float64) string {
	return fmt.Sprintf("€%.2f", value)
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.4f%%", value*100)
}

func (s *Service) getKlaviyoReportCampaigns(ctx context.Context) ([]KlaviyoReportTemplateCampaign, error) {
	now := time.Now().UTC()

	res, err := s.KlaviyoClient.GetCampaignsWithResponse(ctx, &klaviyo.GetCampaignsParams{
		Revision: "2023-12-15",
		Filter:   fmt.Sprintf("equals(messages.channel,'email'),equals(archived,false),less-than(scheduled_at,%s)", now.Format(time.RFC3339)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get campaigns: %w", err)
	}

	metrics, err := s.getMetrics(ctx, "$attributed_message")
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics conversions: %w", err)
	}

	templateCampaigns := []KlaviyoReportTemplateCampaign{}
	for _, campaign := range res.JSON200.Data {
		recipientRes, err := s.KlaviyoClient.GetCampaignRecipientEstimationWithResponse(ctx, campaign.Id, &klaviyo.GetCampaignRecipientEstimationParams{
			Revision: "2023-12-15",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get campaign recipient estimation: %w", err)
		}

		recipientCount := recipientRes.JSON200.Data.Attributes.EstimatedRecipientCount

		metric, ok := metrics[campaign.Id]
		if !ok {
			s.Logger.Warn("failed to get metrics for campaign", "campaign_id", campaign.Id)
			continue
		}

		templateCampaign := calculateCampaign(campaign.Attributes.Name, metric, recipientCount)
		templateCampaigns = append(templateCampaigns, templateCampaign)
	}

	return templateCampaigns, nil
}

func calculateCampaign(name string, metric Metric, recipientCount int) KlaviyoReportTemplateCampaign {
	campaign := KlaviyoReportTemplateCampaign{}
	campaign.Name = name
	campaign.TotalRecipients = recipientCount
	campaign.OrdersPlaced = metric.Count

	if metric.Count == 0 || metric.Revenue == 0 || recipientCount == 0 {
		return campaign
	}

	campaign.ConversionRate = float64(metric.Count) / float64(recipientCount)
	campaign.ConversionValue = metric.Revenue / float64(metric.Count)
	campaign.RevenuePerRecipient = metric.Revenue / float64(recipientCount)

	return campaign
}
