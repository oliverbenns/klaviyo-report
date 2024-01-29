package api

import (
	"context"
	"fmt"
	"time"

	"github.com/oliverbenns/klaviyo-report/generated/klaviyo"
	"github.com/oliverbenns/klaviyo-report/internal/conv"
)

// Openapi generator does not generate this - uses inline struct.
type MetricAggAttributes struct {
	By           *[]klaviyo.MetricAggregateQueryResourceObjectAttributesBy          `json:"by,omitempty"`
	Filter       []string                                                           `json:"filter"`
	Interval     *klaviyo.MetricAggregateQueryResourceObjectAttributesInterval      `json:"interval,omitempty"`
	Measurements []klaviyo.MetricAggregateQueryResourceObjectAttributesMeasurements `json:"measurements"`
	MetricId     string                                                             `json:"metric_id"`
	PageCursor   *string                                                            `json:"page_cursor,omitempty"`
	PageSize     *int                                                               `json:"page_size,omitempty"`
	ReturnFields *[]string                                                          `json:"return_fields,omitempty"`
	Sort         *klaviyo.MetricAggregateQueryResourceObjectAttributesSort          `json:"sort,omitempty"`
	Timezone     *string                                                            `json:"timezone,omitempty"`
}

type Metric struct {
	Count   int
	Revenue float64
}

type MetricsByCampaignID map[string]Metric

func (s *Service) getMetrics(ctx context.Context, by string) (MetricsByCampaignID, error) {
	now := time.Now().UTC()
	oneMonthAgo := now.AddDate(0, 0, -30)

	metricsRes, err := s.KlaviyoClient.GetMetricsWithResponse(ctx, &klaviyo.GetMetricsParams{
		Revision: "2023-12-15",
		Filter:   conv.Ptr("equals(integration.name,'Shopify')"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	placedOrderMetricID := ""
	for _, metric := range metricsRes.JSON200.Data {
		if *metric.Attributes.Name == "Placed Order" {
			placedOrderMetricID = metric.Id
			break
		}
	}

	if placedOrderMetricID == "" {
		return nil, fmt.Errorf("failed to find placed order metric")
	}

	params := &klaviyo.QueryMetricAggregatesParams{
		Revision: "2023-12-15",
	}

	body := klaviyo.QueryMetricAggregatesJSONRequestBody{
		Data: klaviyo.MetricAggregateQueryResourceObject{
			Type: "metric-aggregate",
			Attributes: MetricAggAttributes{
				MetricId: placedOrderMetricID,
				Measurements: []klaviyo.MetricAggregateQueryResourceObjectAttributesMeasurements{
					"count",
					"sum_value",
				},
				By: &[]klaviyo.MetricAggregateQueryResourceObjectAttributesBy{
					klaviyo.MetricAggregateQueryResourceObjectAttributesBy(by),
				},
				Interval: conv.Ptr(klaviyo.MetricAggregateQueryResourceObjectAttributesInterval("month")),
				Filter: []string{
					fmt.Sprintf("greater-or-equal(datetime,%s),less-than(datetime,%s)", oneMonthAgo.Format(time.RFC3339), now.Format(time.RFC3339)),
				},
			},
		},
	}

	aggRes, err := s.KlaviyoClient.QueryMetricAggregatesWithResponse(ctx, params, body)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric aggregates: %w", err)
	}

	metrics := MetricsByCampaignID{}
	for _, aggResult := range aggRes.JSON200.Data.Attributes.Data {
		campaignID := aggResult.Dimensions[0]

		count := aggResult.Measurements["count"]
		totalCount, err := sumMeasurement(count)
		if err != nil {
			return nil, fmt.Errorf("failed to sum metric count aggregate measurements: %w", err)
		}

		revenue := aggResult.Measurements["sum_value"]
		totalRevenue, err := sumMeasurement(revenue)
		if err != nil {
			return nil, fmt.Errorf("failed to sum metric sum_value aggregate measurements: %w", err)
		}

		metrics[campaignID] = Metric{
			Count:   int(totalCount),
			Revenue: totalRevenue,
		}
	}

	return metrics, nil
}

// sumMeasurement sums the measurements in a metric aggregate response.
// We query over 30 days but sometimes there are 2 results (31 day months?)
func sumMeasurement(measurements interface{}) (float64, error) {
	vals, ok := measurements.([]interface{})
	if !ok {
		return 0, fmt.Errorf("failed to convert metric aggregate measurement")
	}

	sum := 0.
	for _, val := range vals {
		realVal, ok := val.(float64)
		if !ok {
			return 0, fmt.Errorf("failed to convert metric aggregate measurement to float64")
		}
		sum += realVal
	}

	return sum, nil
}
