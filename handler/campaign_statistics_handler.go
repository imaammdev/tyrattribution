package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"tyrattribution/service"
)

type CampaignStatisticsHandler struct {
	campaignStatisticsService service.CampaignStatisticsService
}

func NewCampaignStatisticsHandler(campaignStatisticsService service.CampaignStatisticsService) *CampaignStatisticsHandler {
	return &CampaignStatisticsHandler{
		campaignStatisticsService: campaignStatisticsService,
	}
}

func (h *CampaignStatisticsHandler) GetCampaignStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	campaignIDStr := r.URL.Query().Get("campaign_id")
	if campaignIDStr == "" {
		http.Error(w, "campaign_id parameter is required", http.StatusBadRequest)
		return
	}

	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		http.Error(w, "Invalid campaign_id format", http.StatusBadRequest)
		return
	}

	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		groupBy = "daily"
	}

	if groupBy != "daily" && groupBy != "weekly" && groupBy != "monthly" {
		http.Error(w, "group_by must be daily, weekly, or monthly", http.StatusBadRequest)
		return
	}

	statistics, err := h.campaignStatisticsService.GetCampaignStatistics(r.Context(), campaignID, groupBy)
	if err != nil {
		http.Error(w, "Failed to get campaign statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(statistics)
}