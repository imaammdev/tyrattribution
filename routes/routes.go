package routes

import (
	"net/http"
	"tyrattribution/handler"
	"tyrattribution/publisher"
	"tyrattribution/service"
)

func SetupRoutes(clickEventPublisher *publisher.ClickEventPublisher, conversionEventPublisher *publisher.ConversionEventPublisher, campaignJournalService service.CampaignJournalService, campaignStatisticsService service.CampaignStatisticsService) *http.ServeMux {
	mux := http.NewServeMux()

	clickEventHandler := handler.NewClickEventHandler(clickEventPublisher)
	conversionEventHandler := handler.NewConversionEventHandler(conversionEventPublisher)
	campaignJournalHandler := handler.NewCampaignJournalHandler(campaignJournalService)
	campaignStatisticsHandler := handler.NewCampaignStatisticsHandler(campaignStatisticsService)

	mux.HandleFunc("POST /api/clicks", clickEventHandler.CreateClickEvent)
	mux.HandleFunc("POST /api/conversions", conversionEventHandler.CreateConversionEvent)
	mux.HandleFunc("POST /api/calculate-yesterday-metrics", campaignJournalHandler.CalculateYesterdayMetrics)
	mux.HandleFunc("GET /api/campaign-statistics", campaignStatisticsHandler.GetCampaignStatistics)

	return mux
}
