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
		VNR2Key     string    `json:"vn_key"`
		VideoKey    string    `json:"video_key"`

		PromoIDS []string `json:"promo_ids"`
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
	user, err := cfg.DBQueries.GetUserByID(r.Context(), input.UserID)
	if err != nil {
		log.Println("DB ERROR error getting user: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting user")
		return

	}
	if user.Role == "admin" {
		helpers.RespondWithError(w, http.StatusBadRequest, "admin should not be creating request")
		return
	}

	quoteRequest, err := cfg.DBQueries.CreateQuoteRequest(r.Context(), database.CreateQuoteRequestParams{
		UserID:      input.UserID,
		ServiceID:   input.ServiceID,
		Description: input.Description,
		Attachments: input.Attachments,
		PromoIds:    input.PromoIDS,
		VnR2Key:     input.VNR2Key,
		VideoKey:    input.VideoKey,
	})

	if err != nil {
		log.Println("DB ERROR error creating quoterequest: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error creating quoterequest")
		return
	}
	res := struct {
		Qr QuoteRequest `json:"quote_request"`
	}{
		Qr: DbQuoteRequestToQuoteRequest(quoteRequest),
	}

	helpers.RespondWithJson(w, http.StatusOK, res)

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

	qrs, err := cfg.DBQueries.GetUserQuoteRequests(r.Context(), database.GetUserQuoteRequestsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
		UserID: user.ID,
	})
	if err != nil {
		log.Println("DB ERROR error getting quote requests: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting quote requests")
		return
	}

	count, err := cfg.DBQueries.CountUserQuoteRequests(r.Context(), user.ID)
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

	qrs, err := cfg.DBQueries.GetQuoteRequests(r.Context(), database.GetQuoteRequestsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		log.Println("DB ERROR error getting quote requests: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting quote requests")
		return
	}

	count, err := cfg.DBQueries.CountQuoteRequests(r.Context())
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
	if count > 0 {
		log.Println(qrs[0].ServiceName)
	}

	helpers.RespondWithJson(w, http.StatusOK, res)

}

func (cfg *Config) AdminCreateQuoteHandler(w http.ResponseWriter, r *http.Request) {

	input := struct {
		UserID         uuid.UUID    `json:"user_id"`
		QuoteRequestID uuid.UUID    `json:"quote_request_id"`
		Amount         string       `json:"amount"`
		Breakdown      []DBItem     `json:"breakdown"`
		Discounts      []DBDiscount `json:"discounts"`
		PromoIDS       []string     `json:"promo_ids"`
		Notes          string       `json:"notes"`
		ExpiresAt      time.Time    `json:"expires_at"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Println(err)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// 1. Precise Validation
	if input.QuoteRequestID == uuid.Nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Quote Request ID is required")
		return
	}
	if input.UserID == uuid.Nil {

		helpers.RespondWithError(w, http.StatusBadRequest, "User  ID is required")
		return
	}
	if input.Amount == "" {
		log.Println("here")

		helpers.RespondWithError(w, http.StatusBadRequest, "Amount cannot be empty")
		return
	}
	if len(input.Breakdown) == 0 {
		log.Println("here")

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
	// 2. Marshal Discounts to JSON for the DB
	discountsJSON, err := json.Marshal(input.Discounts)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "Error processing discounts data")
		return
	}

	// 3. Execute DB Insert
	dbQuote, err := cfg.DBQueries.CreateQuote(r.Context(), database.CreateQuoteParams{
		QuoteRequestID: input.QuoteRequestID,
		UserID:         input.UserID,
		Amount:         input.Amount,
		Breakdown:      breakdownJSON,
		PromoIds:       input.PromoIDS,
		Discounts:      discountsJSON,
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
	_ = cfg.DBQueries.UpdateQuoteRequestStatus(r.Context(), database.UpdateQuoteRequestStatusParams{
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
		"draft": true,
		"sent":  true,
		// admin and staff should not be accepting or declining
		// "accepted":  true,
		// "declined":  true,
		"expired":   true,
		"in-review": true,
	}

	if !validStatuses[strings.ToLower(input.Status)] {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid quote status provided")
		return
	}
	quote, err := cfg.DBQueries.GetQuote(r.Context(), parsedId)
	if err != nil {
		log.Println("DB ERROR error getting quote  " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting quote")
		return

	}
	if quote.Status == "declined" {
		helpers.RespondWithError(w, http.StatusBadGateway, "user rejected quote already")
		return

	}

	err = cfg.DBQueries.UpdateQuoteStatus(r.Context(), database.UpdateQuoteStatusParams{
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

	qs, err := cfg.DBQueries.GetQuotes(r.Context(), database.GetQuotesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		log.Println("DB ERROR error getting quotes: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting quotes")
		return
	}

	count, err := cfg.DBQueries.CountQuotes(r.Context())
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
				Discounts:      q.Discounts,
				Notes:          q.Notes,
				Status:         q.Status,
				ExpiresAt:      q.ExpiresAt,
				CreatedAt:      q.CreatedAt,
				UpdatedAt:      q.UpdatedAt,
			}),
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
	qr, err := cfg.DBQueries.GetQuoteRequest(context.Background(), parsedId)
	if err != nil {
		log.Println("DB ERROR error getting qr: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting qr ")
		return
	}
	if qr.UserID != user.ID {
		helpers.RespondWithError(w, http.StatusUnauthorized, "")
		return

	}
	err = cfg.DBQueries.UpdateQuoteRequestDescription(r.Context(), database.UpdateQuoteRequestDescriptionParams{
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

	qs, err := cfg.DBQueries.GetUserQuotesWithService(r.Context(), database.GetUserQuotesWithServiceParams{
		Limit:  int32(limit),
		Offset: int32(offset),
		UserID: user.ID,
	})
	if err != nil {
		log.Println("DB ERROR error getting quotes: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting quotes")
		return
	}

	count, err := cfg.DBQueries.CountUserQuotes(r.Context(), user.ID)
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

	q, err := cfg.DBQueries.GetUserQuoteWithService(r.Context(), database.GetUserQuoteWithServiceParams{
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
			ServiceName:    q.ServiceName,
			ServiceID:      q.ServiceID,
		}),
	}

	helpers.RespondWithJson(w, http.StatusOK, res)
}

func (cfg *Config) CustomerUpdateQuoteStatusHandler(w http.ResponseWriter, r *http.Request) {
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
		// "draft": true,
		// "sent":  true,
		// user can only accept or decline
		"accepted": true,
		"declined": true,
		// "expired":   true,
		// "in-review": true,
	}

	if !validStatuses[strings.ToLower(input.Status)] {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid quote status provided")
		return
	}
	// validation

	user, httpStatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpStatus, err.Error())
		return
	}
	quote, err := cfg.DBQueries.GetQuote(r.Context(), parsedId)
	if err != nil {
		log.Println("DB ERROR error getting quote  " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting quote")
		return

	}
	//  make sure they own the quote request of the quote and status is accept or reject
	if input.Status != "accepted" && input.Status != "declined" {
		helpers.RespondWithError(w, http.StatusUnauthorized, "unauthorized action")
		return
	}
	qr, err := cfg.DBQueries.GetQuoteRequest(r.Context(), quote.QuoteRequestID)
	if err != nil {
		log.Println("DB ERROR error getting quote request in validating user request to update quote status : " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error validating user")
		return

	}
	if user.ID != qr.UserID {
		helpers.RespondWithError(w, http.StatusUnauthorized, "unauthorized action")
		return

	}

	if input.Status == "accepted" {

		err = helpers.AcceptQuoteAndCreateInvoice(r.Context(), helpers.AcceptQuoteAndCreateInvoiceParams{
			Quote:         &quote,
			Customer:      &user,
			Db:            cfg.DBConn,
			Queries:       cfg.DBQueries,
			InvoiceNumber: GenerateInvoiceNumber(),
			AdminID:       quote.UserID,
		})
		if err != nil {
			log.Println("DB ERROR error accepting quote and generating quote: " + err.Error())
			helpers.RespondWithError(w, http.StatusInternalServerError, "error updating quote status")
			return
		}

		helpers.RespondWithJson(w, http.StatusOK, "quote status updated")
		return
	}

	err = cfg.DBQueries.UpdateQuoteStatus(r.Context(), database.UpdateQuoteStatusParams{
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
