package handlers

import (
	"database/sql"
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

func (cfg *Config) VerifyPromoHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Promo code required")
		return
	}

	dbPromo, err := cfg.DBQueries.GetActivePromoByCode(r.Context(), code)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.RespondWithError(w, http.StatusNotFound, "Invalid or expired promotion")
			return
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, "Terminal registry error")
		return
	}

	helpers.RespondWithJson(w, http.StatusOK, map[string]any{
		"promotion": dbPromoToPromo(dbPromo),
	})
}

type CreatePromoInput struct {
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Breakdown   []DBDiscount `json:"breakdown"`
	ServiceID   uuid.UUID    `json:"service_id"`
	IsActive    bool         `json:"is_active"`
	StartsAt    time.Time    `json:"starts_at"`
	ExpiresAt   time.Time    `json:"expires_at"`
	Attachments []string     `json:"attachments"`
}

func (cfg *Config) AdminCreatePromotionHandler(w http.ResponseWriter, r *http.Request) {
	var input CreatePromoInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid payload")
		return
	}
	if input.ServiceID == uuid.Nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid service")
		return
	}
	log.Println(input.ServiceID)

	service, err := cfg.DBQueries.GetService(r.Context(), input.ServiceID)
	if err != nil {
		log.Println("DB ERR: error getting service: ", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting service")
		return
	}
	// Validation
	if input.Code == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Code is  required")
		return
	}
	jsonBreakdown, err := json.Marshal(input.Breakdown)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "error converting breakdown to jsonb")
		return
	}

	promo, err := cfg.DBQueries.CreatePromotion(r.Context(), database.CreatePromotionParams{
		Code:        strings.ToUpper(input.Code),
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},

		Breakdown:   jsonBreakdown,
		ServiceID:   uuid.NullUUID{Valid: true, UUID: service.ID},
		IsActive:    sql.NullBool{Bool: input.IsActive, Valid: true},
		StartsAt:    sql.NullTime{Time: input.StartsAt, Valid: !input.StartsAt.IsZero()},
		ExpiresAt:   sql.NullTime{Time: input.ExpiresAt, Valid: !input.ExpiresAt.IsZero()},
		Attachments: input.Attachments,
	},
	)

	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to register promotion: "+err.Error())
		return
	}

	helpers.RespondWithJson(w, http.StatusCreated, dbPromoToPromo(promo))
}

func (cfg *Config) AdminListPromotionsHandler(w http.ResponseWriter, r *http.Request) {
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
	promos, err := cfg.DBQueries.ListPromos(r.Context(), database.ListPromosParams{
		Offset: int32(offset),
		Limit:  int32(limit),
	})
	if err != nil {
		log.Println("DB ERROR error getting promos: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting promos")
		return
	}

	count, err := cfg.DBQueries.CountPromos(r.Context())
	if err != nil {
		log.Println("DB ERROR error getting  promos count: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting  promos  count")
		return
	}
	res := struct {
		Promos []Promotion `json:"promotions"`
		Total  int64       `json:"total"`
	}{
		Promos: dbPromosToPromos(promos),
		Total:  count,
	}
	// log.Println(res.Promos)
	// log.Println(res.Total)

	helpers.RespondWithJson(w, http.StatusOK, res)
}

func (cfg *Config) GetActivePromosHandler(w http.ResponseWriter, r *http.Request) {

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

	promos, err := cfg.DBQueries.GetActivePromos(r.Context(), database.GetActivePromosParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		log.Println("DB ERROR error getting promos: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting promos")
		return
	}
	count, err := cfg.DBQueries.CountActivePromos(r.Context())
	if err != nil {
		log.Println("DB ERROR error getting active promos count: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting active promos count")
		return
	}
	res := struct {
		Promotion []Promotion `json:"promotions"`
		Total     int64       `json:"total"`
	}{
		Promotion: dbPromosToPromos(promos),
		Total:     count,
	}
	helpers.RespondWithJson(w, http.StatusOK, res)
}

func (cfg *Config) GetPromoHandler(w http.ResponseWriter, r *http.Request) {
	serviceId := chi.URLParam(r, "id")
	if serviceId == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "")
		return
	}

	parsedID, err := uuid.Parse(serviceId)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "error parsing id")
		return
	}
	promo, err := cfg.DBQueries.GetPromotionByID(r.Context(), parsedID)
	if err != nil {
		log.Println("DB ERROR error getting promo: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting promo")
		return
	}
	res := struct {
		Promo Promotion `json:"promotion"`
	}{
		Promo: dbPromoToPromo(promo),
	}
	helpers.RespondWithJson(w, http.StatusOK, res)
}
