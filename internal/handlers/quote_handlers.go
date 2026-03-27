package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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
		return

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
		QuoteRequests []GetUserQuoteRequestsRow `json:"quote_requests"`
		Total         int64                     `json:"total"`
	}{
		QuoteRequests: DbUserQuoteRequestRowsToUserQuoteRequestsRow(qrs),
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

	input := struct {
		QuoteRequestID uuid.UUID   `json:"quote_request_id"`
		Amount         string      `json:"amount"`
		Breakdown      []QuoteItem `json:"breakdown"`
		Notes          string      `json:"notes"`
		ExpiresAt      time.Time   `json:"expires_at"`
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
	quoteId := chi.URLParam(r, "id")
	if quoteId == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "")
		return
	}

	parsedId, err := uuid.Parse(quoteId)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "error parsing id")
		return
	}
	input := struct {
		Status string `json:"status"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	if input.Status == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Amount cannot be empty")
		return
	}
	validStatuses := map[string]bool{
		"draft":     true,
		"sent":      true,
		"accepted":  true,
		"declined":  true,
		"expired":   true,
		"in-review": true,
	}

	if !validStatuses[strings.ToLower(input.Status)] {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid quote status provided")
		return
	}
	err = cfg.DB.UpdateQuoteStatus(r.Context(), database.UpdateQuoteStatusParams{
		ID:     parsedId,
		Status: database.QuoteStatus(input.Status),
	})
	if err != nil {
		log.Println("DB ERROR error updating quote status: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error updating quote status")
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, "quote status updated")
}
func (cfg *Config) AdminGetQuotesHandler(w http.ResponseWriter, r *http.Request) {

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

	qs, err := cfg.DB.GetQuotes(r.Context(), database.GetQuotesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		log.Println("DB ERROR error getting quotes: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting quotes")
		return
	}

	count, err := cfg.DB.CountQuotes(r.Context())
	if err != nil {
		log.Println("DB ERROR error getting  quotes requests count: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting  quotes  count")
		return
	}
	type ReturnedQuoteType struct {
		Quote
		ServiceName string `json:"service_name"`
		ServiceIcon string `json:"service_icon"`
	}
	newqs := []ReturnedQuoteType{}
	for _, q := range qs {
		newqs = append(newqs, ReturnedQuoteType{

			Quote: DbQuoteToQuote(database.Quote{
				ID:             q.ID,
				QuoteRequestID: q.QuoteRequestID,
				Amount:         q.Amount,
				Breakdown:      q.Breakdown,
				Notes:          q.Notes,
				Status:         q.Status,
				ExpiresAt:      q.ExpiresAt,
				CreatedAt:      q.CreatedAt,
				UpdatedAt:      q.UpdatedAt,
			}),
			ServiceIcon: q.ServiceIcon,
			ServiceName: q.ServiceName,
		})
	}
	res := struct {
		Quotes []ReturnedQuoteType `json:"quotes"`
		Total  int64               `json:"total"`
	}{
		Quotes: newqs,
		Total:  count,
	}
	helpers.RespondWithJson(w, http.StatusOK, res)

}

func (cfg *Config) UpdateUserQuoteRequestDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	user, httpstatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpstatus, err.Error())
		return
	}
	qrId := chi.URLParam(r, "id")
	if qrId == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "")
		return
	}

	parsedId, err := uuid.Parse(qrId)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "error parsing id")
		return
	}
	input := struct {
		Description string `json:"description"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}
	if input.Description == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, description, icon and image can't be empty ")
		return
	}
	qr, err := cfg.DB.GetQuoteRequest(context.Background(), parsedId)
	if err != nil {
		log.Println("DB ERROR error getting qr: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting qr ")
		return
	}
	if qr.UserID != user.ID {
		helpers.RespondWithError(w, http.StatusUnauthorized, "")
		return

	}
	err = cfg.DB.UpdateQuoteRequestDescription(r.Context(), database.UpdateQuoteRequestDescriptionParams{
		ID:          qr.ID,
		Description: input.Description,
	})
	if err != nil {
		log.Println("DB ERROR error updating  qr desc: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error updating qr desc")
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, "qr desc updated")
}

func (cfg *Config) GetUserQuotesWithServiceHandler(w http.ResponseWriter, r *http.Request) {
	user, httpstatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpstatus, err.Error())
		return

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

	qs, err := cfg.DB.GetUserQuotesWithService(r.Context(), database.GetUserQuotesWithServiceParams{
		Limit:  int32(limit),
		Offset: int32(offset),
		UserID: user.ID,
	})
	if err != nil {
		log.Println("DB ERROR error getting quotes: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting quotes")
		return
	}

	count, err := cfg.DB.CountUserQuotes(r.Context(), user.ID)
	if err != nil {
		log.Println("DB ERROR error getting user quotes count: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting user quotes  count")
		return
	}
	res := struct {
		Quotes []GetUserQuotesWithServiceRow `json:"quotes"`
		Total  int64                         `json:"total"`
	}{
		Quotes: DbUserQuotesWithServiceRowsToUserQuotesWithServiceRows(qs),
		Total:  count,
	}
	helpers.RespondWithJson(w, http.StatusOK, res)

}

func (cfg *Config) GetUserQuoteWithServiceHandler(w http.ResponseWriter, r *http.Request) {
	quoteId := chi.URLParam(r, "id")
	if quoteId == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Quote ID is required")
		return // Added return
	}

	parsedId, err := uuid.Parse(quoteId)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "error parsing id")
		return // Added return
	}

	user, httpstatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpstatus, err.Error())
		return // Added return
	}

	// You can remove limit/offset logic here, it's not used for a single GET

	q, err := cfg.DB.GetUserQuoteWithService(r.Context(), database.GetUserQuoteWithServiceParams{
		ID:     parsedId,
		UserID: user.ID,
	})
	if err != nil {
		log.Println("DB ERROR error getting quote: " + err.Error())
		// If not found, send 404 so frontend shows "Record Not Found"
		helpers.RespondWithError(w, http.StatusNotFound, "quote not found")
		return
	}

	// Wrap in a "quote" key to match your frontend setQuote(res.data.quote)
	res := struct {
		Quote GetUserQuotesWithServiceRow `json:"quote"`
	}{
		Quote: DbUserQuotesWithServiceRowToUserQuotesWithServiceRow(database.GetUserQuotesWithServiceRow{
			ID:             q.ID,
			QuoteRequestID: q.QuoteRequestID,
			Amount:         q.Amount,
			Breakdown:      q.Breakdown,
			Notes:          q.Notes,
			Status:         q.Status,
			CreatedAt:      q.CreatedAt,
			UpdatedAt:      q.UpdatedAt,
			ExpiresAt:      q.ExpiresAt,
			ServiceIcon:    q.ServiceIcon,
			ServiceName:    q.ServiceName,
			ServiceID:      q.ServiceID,
		}),
	}

	helpers.RespondWithJson(w, http.StatusOK, res)
}
