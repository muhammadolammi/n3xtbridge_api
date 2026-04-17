package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

func (cfg *Config) CreateServiceHandler(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		IsFeatured  bool     `json:"is_featured"`
		Tags        []string `json:"tags"`
		Image       string   `json:"image"`
		MinPrice    string   `json:"min_price"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}
	if input.Name == "" || input.Category == "" || input.Description == "" || input.Image == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, name, category, description and image can't be empty ")
		return
	}

	service, err := cfg.DBQueries.CreateService(r.Context(), database.CreateServiceParams{
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		IsFeatured:  input.IsFeatured,
		Tags:        pq.StringArray(input.Tags),
		Image:       input.Image,
		MinPrice:    input.MinPrice,
	})
	if err != nil {
		log.Println("DB ERROR: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error creating service request ")
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, dbServiceToService(service))
}

func (cfg *Config) GetActiveServicesHandler(w http.ResponseWriter, r *http.Request) {

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

	services, err := cfg.DBQueries.GetActiveServices(r.Context(), database.GetActiveServicesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		log.Println("DB ERROR error getting services: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting services")
		return
	}
	count, err := cfg.DBQueries.CountActiveServices(r.Context())
	if err != nil {
		log.Println("DB ERROR error getting active services count: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting active services count")
		return
	}
	res := struct {
		Services []Service `json:"services"`
		Total    int64     `json:"total"`
	}{
		Services: dbServicesToServices(services),
		Total:    count,
	}
	helpers.RespondWithJson(w, http.StatusOK, res)
}

func (cfg *Config) AdminListAllServicesHandler(w http.ResponseWriter, r *http.Request) {

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

	services, err := cfg.DBQueries.GetServices(r.Context(), database.GetServicesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		log.Println("DB ERROR error getting services: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting services")
		return
	}
	count, err := cfg.DBQueries.CountServices(r.Context())
	if err != nil {
		log.Println("DB ERROR error getting active services count: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting active services count")
		return
	}
	res := struct {
		Services []Service `json:"services"`
		Total    int64     `json:"total"`
	}{
		Services: dbServicesToServices(services),
		Total:    count,
	}
	helpers.RespondWithJson(w, http.StatusOK, res)

}

func (cfg *Config) AdminUpdateServiceStatusHandler(w http.ResponseWriter, r *http.Request) {
	serviceId := chi.URLParam(r, "id")
	if serviceId == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "")
		return
	}

	parsedId, err := uuid.Parse(serviceId)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "error parsing id")
		return
	}
	input := struct {
		IsActive *bool `json:"is_Active"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}
	if input.IsActive == nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, name, category, description, icon and image can't be empty ")
		return
	}
	service, err := cfg.DBQueries.GetService(context.Background(), parsedId)
	if err != nil {
		log.Println("DB ERROR error getting service: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting service ")
		return
	}

	err = cfg.DBQueries.UpdateServiceStatus(r.Context(), database.UpdateServiceStatusParams{
		ID:       service.ID,
		IsActive: *input.IsActive,
	})
	if err != nil {
		log.Println("DB ERROR error updating service status: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error updating service status")
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, "service status updated")

}

func (cfg *Config) GetServiceHandler(w http.ResponseWriter, r *http.Request) {
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
	service, err := cfg.DBQueries.GetService(r.Context(), parsedID)
	if err != nil {
		log.Println("DB ERROR error getting service: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting service")
		return
	}
	res := struct {
		Services Service `json:"service"`
	}{
		Services: dbServiceToService(service),
	}
	helpers.RespondWithJson(w, http.StatusOK, res)
}
