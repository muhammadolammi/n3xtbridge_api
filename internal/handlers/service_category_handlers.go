package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

func (cfg *Config) CreateServiceCategoryHandler(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Slug        string `json:"slug"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}
	if input.Name == "" || input.Description == "" || input.Slug == "" || input.Icon == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, name, description, slug and icon can't be empty ")
		return
	}

	_, err = cfg.DBQueries.CreateServiceCategory(r.Context(), database.CreateServiceCategoryParams{
		Slug:        input.Slug,
		Name:        input.Name,
		Icon:        input.Icon,
		Description: input.Description,
	})
	if err != nil {
		log.Println("DB ERROR: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error creating service category")
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, "")
}

func (cfg *Config) GetActiveServiceCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	serviceCategories, err := cfg.DBQueries.GetActiveServiceCategories(r.Context())
	if err != nil {
		log.Println("DB ERROR error getting service categories: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting service categories")
		return
	}
	// log.Println(serviceCategories)
	res := struct {
		ServiceCategories []ServiceCategory `json:"service_categories"`
	}{
		ServiceCategories: dbServiceCategoriesToServiceCategories(serviceCategories),
	}
	// log.Println(res)

	helpers.RespondWithJson(w, http.StatusOK, res)
}
