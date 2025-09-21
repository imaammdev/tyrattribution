package handler

import (
	"encoding/json"
	"net/http"

	"tyrattribution/service"
)

type CampaignJournalHandler struct {
	campaignJournalService service.CampaignJournalService
}

func NewCampaignJournalHandler(campaignJournalService service.CampaignJournalService) *CampaignJournalHandler {
	return &CampaignJournalHandler{
		campaignJournalService: campaignJournalService,
	}
}

type CampaignJournalResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (h *CampaignJournalHandler) CalculateYesterdayMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.campaignJournalService.CalculateYesterdayMetrics(r.Context()); err != nil {
		http.Error(w, "Failed to calculate yesterday metrics", http.StatusInternalServerError)
		return
	}

	response := CampaignJournalResponse{
		Message: "Yesterday metrics calculated and saved successfully",
		Status:  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}