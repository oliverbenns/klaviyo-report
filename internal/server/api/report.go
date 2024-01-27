package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/oliverbenns/klaviyo-report/generated/klaviyo"
)

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

	html := fmt.Sprintf(`
		<html>
			<head>
				<title>Klaviyo Report Prototype</title>
			</head>
			<body>
				<h1>Klaviyo Report Prototype</h1>
				<hr />
				<h3>Report for %s</h3>

				<p>The report goes here</p>
			</body>
		</html>
	`, res.JSON200.Data.Attributes.ContactInformation.OrganizationName)

	c.Writer.WriteString(html)

	c.Status(200)
	return
}
