package handler

import (
	"net/http"
	"tyrattribution/publisher"
	"tyrattribution/service"
)

func SetupRoutes(clickEventPublisher *publisher.ClickEventPublisher, conversionEventPublisher *publisher.ConversionEventPublisher, campaignJournalService service.CampaignJournalService, campaignStatisticsService service.CampaignStatisticsService) *http.ServeMux {
	mux := http.NewServeMux()

	clickEventHandler := NewClickEventHandler(clickEventPublisher)
	conversionEventHandler := NewConversionEventHandler(conversionEventPublisher)
	campaignJournalHandler := NewCampaignJournalHandler(campaignJournalService)
	campaignStatisticsHandler := NewCampaignStatisticsHandler(campaignStatisticsService)

	mux.HandleFunc("POST /api/clicks", clickEventHandler.CreateClickEvent)
	mux.HandleFunc("POST /api/conversions", conversionEventHandler.CreateConversionEvent)
	mux.HandleFunc("POST /api/calculate-yesterday-metrics", campaignJournalHandler.CalculateYesterdayMetrics)
	mux.HandleFunc("GET /api/campaign-statistics", campaignStatisticsHandler.GetCampaignStatistics)

	return mux
}
