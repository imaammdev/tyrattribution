package handler

import (
	"encoding/json"
	"net/http"
	"time"
	"tyrattribution/publisher"

	"github.com/google/uuid"
)

type ClickEventHandler struct {
	clickEventPub *publisher.ClickEventPublisher
}

func NewClickEventHandler(clickEventPub *publisher.ClickEventPublisher) *ClickEventHandler {
	return &ClickEventHandler{
		clickEventPub: clickEventPub,
	}
}

type ClickEventRequest struct {
	ClickID    string `json:"click_id,omitempty"`
	CampaignID string `json:"campaign_id"`
	UserID     string `json:"user_id"`
	ClickDate  string `json:"click_date"`
	Source     string `json:"source"`
}

type ClickEventResponse struct {
	ClickID string `json:"click_id"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (h *ClickEventHandler) CreateClickEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClickEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	campaignID, err := uuid.Parse(req.CampaignID)
	if err != nil {
		http.Error(w, "Invalid campaign_id format", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "Invalid user_id format", http.StatusBadRequest)
		return
	}

	clickDate, err := time.Parse(time.RFC3339, req.ClickDate)
	if err != nil {
		http.Error(w, "Invalid click_date format, use RFC3339", http.StatusBadRequest)
		return
	}

	if req.Source == "" {
		http.Error(w, "Source is required", http.StatusBadRequest)
		return
	}

	var clickID uuid.UUID
	if req.ClickID != "" {
		var err error
		clickID, err = uuid.Parse(req.ClickID)
		if err != nil {
			http.Error(w, "Invalid click_id format", http.StatusBadRequest)
			return
		}
	} else {
		clickID = uuid.New()
	}

	clickEvent := publisher.ClickEvent{
		ClickID:    clickID,
		CampaignID: campaignID,
		UserID:     userID,
		ClickDate:  clickDate,
		Source:     req.Source,
		CreatedAt:  time.Now(),
	}

	if err := h.clickEventPub.PublishClickEvent(clickEvent); err != nil {
		http.Error(w, "Failed to create click event", http.StatusInternalServerError)
		return
	}

	response := ClickEventResponse{
		ClickID: clickEvent.ClickID.String(),
		Message: "Click event created successfully",
		Status:  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
