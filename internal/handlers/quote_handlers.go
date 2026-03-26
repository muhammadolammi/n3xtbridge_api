package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

func (cfg *Config) CreateQuoteRequestHandler(w http.ResponseWriter, r *http.Request) {
	input := struct {
		UserID      uuid.UUID `json:"user_id"`
		ServiceID   uuid.UUID `json:"service_id"`
		Description string    `json:"description"`
		Attachments []string  `json:"attachments"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}
	if input.UserID == uuid.Nil || input.ServiceID == uuid.Nil || input.Description == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, user_id, service_id, and description can't be empty ")
		return
	}
	quoteRequest, err := cfg.DB.CreateQuoteRequest(r.Context(), database.CreateQuoteRequestParams{
		UserID:      input.UserID,
		ServiceID:   input.ServiceID,
		Description: input.Description,
		Attachments: input.Attachments,
	})

	if err != nil {
		log.Println("DB ERROR error creating quoterequest: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error creating quoterequest")
		return
	}

	helpers.RespondWithJson(w, http.StatusOK, DbQuoteRequestToQuoteRequest(quoteRequest))

}

func (cfg *Config) GetUserQuoteRequestsHandler(w http.ResponseWriter, r *http.Request) {
	user, httpstatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpstatus, err.Error())

	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10 // Default
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 // Default
	}

	qrs, err := cfg.DB.GetUserQuoteRequests(r.Context(), database.GetUserQuoteRequestsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
		UserID: user.ID,
	})
	if err != nil {
		log.Println("DB ERROR error getting quote requests: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting quote requests")
		return
	}

	count, err := cfg.DB.CountUserQuoteRequests(r.Context(), user.ID)
	if err != nil {
		log.Println("DB ERROR error getting user quote requests count: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting user quote requests  count")
		return
	}
	res := struct {
		QuoteRequests []QuoteRequest `json:"quote_requests"`
		Total         int64          `json:"total"`
	}{
		QuoteRequests: DbQuoteRequestsToQuoteRequests(qrs),
		Total:         count,
	}
	helpers.RespondWithJson(w, http.StatusOK, res)

}
func (cfg *Config) AdminGetQuoteRequestsHandler(w http.ResponseWriter, r *http.Request) {

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10 // Default
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 // Default
	}

	qrs, err := cfg.DB.GetQuoteRequests(r.Context(), database.GetQuoteRequestsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		log.Println("DB ERROR error getting quote requests: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting quote requests")
		return
	}

	count, err := cfg.DB.CountQuoteRequests(r.Context())
	if err != nil {
		log.Println("DB ERROR error getting user quoute requests count: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting user quote requests  count")
		return
	}
	res := struct {
		QuoteRequests []GetQuoteRequestsRow `json:"quote_requests"`
		Total         int64                 `json:"total"`
	}{
		QuoteRequests: DbQuoteRequestRowsToQuoteRequestsRow(qrs),
		Total:         count,
	}
	helpers.RespondWithJson(w, http.StatusOK, res)

}

func (cfg *Config) AdminCreateQuoteHandler(w http.ResponseWriter, r *http.Request) {
	type Item struct {
		Name        string `json:"name"`
		Cost        string `json:"cost"`
		Description string `json:"description"`
	}

	input := struct {
		QuoteRequestID uuid.UUID `json:"quote_request_id"`
		Amount         string    `json:"amount"`
		Breakdown      []Item    `json:"breakdown"`
		Notes          string    `json:"notes"`
		ExpiresAt      time.Time `json:"expires_at"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// 1. Precise Validation
	if input.QuoteRequestID == uuid.Nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Quote Request ID is required")
		return
	}
	if input.Amount == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Amount cannot be empty")
		return
	}
	if len(input.Breakdown) == 0 {
		helpers.RespondWithError(w, http.StatusBadRequest, "Breakdown must contain at least one item")
		return
	}
	if input.ExpiresAt.IsZero() || input.ExpiresAt.Before(time.Now()) {
		helpers.RespondWithError(w, http.StatusBadRequest, "A valid future expiration date is required")
		return
	}

	// 2. Marshal Breakdown to JSON for the DB
	breakdownJSON, err := json.Marshal(input.Breakdown)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "Error processing breakdown data")
		return
	}

	// 3. Execute DB Insert
	dbQuote, err := cfg.DB.CreateQuote(r.Context(), database.CreateQuoteParams{
		QuoteRequestID: input.QuoteRequestID,
		Amount:         input.Amount,
		Breakdown:      breakdownJSON, // sqlc expects json.RawMessage/[]byte
		Notes:          input.Notes,
		ExpiresAt:      input.ExpiresAt,
	})

	if err != nil {
		log.Println("DB ERROR creating quote: " + err.Error())
		// Handle unique constraint (one quote per request)
		if strings.Contains(err.Error(), "unique constraint") {
			helpers.RespondWithError(w, http.StatusConflict, "A quote already exists for this request")
			return
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// 4. Update QuoteRequest Status to 'quoted' automatically
	_ = cfg.DB.UpdateQuoteRequestStatus(r.Context(), database.UpdateQuoteRequestStatusParams{
		ID:     input.QuoteRequestID,
		Status: database.QuoteRequestStatusQuoted,
	})

	helpers.RespondWithJson(w, http.StatusCreated, DbQuoteToQuote(dbQuote))
}

func (cfg *Config) AdminUpdateQuoteStatusHandler(w http.ResponseWriter, r *http.Request) {
	input := struct {
		QuoteID uuid.UUID `json:"quote_id"`
		Status  string    `json:"status"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	if input.QuoteID == uuid.Nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Quote Request ID is required")
		return
	}
	if input.Status == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Amount cannot be empty")
		return
	}
	validStatuses := map[string]bool{
		"draft":    true,
		"sent":     true,
		"accepted": true,
		"declined": true,
		"expired":  true,
	}

	if !validStatuses[strings.ToLower(input.Status)] {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid quote status provided")
		return
	}
	_ = cfg.DB.UpdateQuoteStatus(r.Context(), database.UpdateQuoteStatusParams{
		ID:     input.QuoteID,
		Status: database.QuoteStatus(input.Status),
	})
}
