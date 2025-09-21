package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"tyrattribution/publisher"
)

type ConversionEventHandler struct {
	conversionEventPub *publisher.ConversionEventPublisher
}

func NewConversionEventHandler(conversionEventPub *publisher.ConversionEventPublisher) *ConversionEventHandler {
	return &ConversionEventHandler{
		conversionEventPub: conversionEventPub,
	}
}

type ConversionEventRequest struct {
	ConversionID   string  `json:"conversion_id,omitempty"`
	UserID         string  `json:"user_id"`
	CampaignID     string  `json:"campaign_id"`
	ConversionDate string  `json:"conversion_date"`
	Value          float64 `json:"value"`
	Type           string  `json:"type"`
	Source         string  `json:"source"`
}

type ConversionEventResponse struct {
	ConversionID string `json:"conversion_id"`
	Message      string `json:"message"`
	Status       string `json:"status"`
}

func (h *ConversionEventHandler) CreateConversionEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConversionEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "Invalid user_id format", http.StatusBadRequest)
		return
	}

	campaignID, err := uuid.Parse(req.CampaignID)
	if err != nil {
		http.Error(w, "Invalid campaign_id format", http.StatusBadRequest)
		return
	}

	conversionDate, err := time.Parse(time.RFC3339, req.ConversionDate)
	if err != nil {
		http.Error(w, "Invalid conversion_date format, use RFC3339", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		http.Error(w, "Type is required", http.StatusBadRequest)
		return
	}

	if req.Source == "" {
		http.Error(w, "Source is required", http.StatusBadRequest)
		return
	}

	var conversionID uuid.UUID
	if req.ConversionID != "" {
		var err error
		conversionID, err = uuid.Parse(req.ConversionID)
		if err != nil {
			http.Error(w, "Invalid conversion_id format", http.StatusBadRequest)
			return
		}
	} else {
		conversionID = uuid.New()
	}

	var value *decimal.Decimal
	if req.Value != 0 {
		val := decimal.NewFromFloat(req.Value)
		value = &val
	}

	conversionEvent := publisher.ConversionEvent{
		ConversionID:   conversionID,
		UserID:         userID,
		CampaignID:     campaignID,
		ConversionDate: conversionDate,
		Value:          value,
		Type:           req.Type,
		Source:         req.Source,
		CreatedAt:      time.Now(),
	}

	if err := h.conversionEventPub.PublishConversionEvent(conversionEvent); err != nil {
		http.Error(w, "Failed to create conversion event", http.StatusInternalServerError)
		return
	}

	response := ConversionEventResponse{
		ConversionID: conversionEvent.ConversionID.String(),
		Message:      "Conversion event created successfully",
		Status:       "success",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}