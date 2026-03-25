package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
	invoicepackage "github.com/muhammadolammi/n3xtbridge_api/internal/invoice"
)

func (cfg *Config) CreateInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	var input invoicepackage.InvoiceInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}
	user, httpstatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpstatus, err.Error())

	}

	invoice := invoicepackage.CreateInvoice(input)

	totalStr := fmt.Sprintf("%.2f", invoice.Total)
	jsonBItems, err := json.Marshal(input.Items)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "error converting items to jsonb")
		return
	}
	jsonBDiscounts, err := json.Marshal(input.Discounts)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "error converting discounts to jsonb")
		return
	}
	// fmt.Printf("Received Discounts: %+v\n", input.Discounts)
	// fmt.Printf("Marshalled Discounts: %+v\n", string(jsonBDiscounts))

	// helpers.RespondWithError(w, http.StatusInternalServerError, "")
	// return

	dbParams := database.CreateInvoiceParams{
		InvoiceNumber: invoice.InvoiceNumber,
		UserID:        user.ID,
		CustomerName:  invoice.CustomerName,
		CustomerEmail: invoice.CustomerEmail,
		CustomerPhone: sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		Total:         totalStr,
		Notes:         input.Notes,
		Discounts:     jsonBDiscounts,
		Items:         jsonBItems,
	}

	// Save invoice to database
	ctx := context.Background()
	dbInvoice, err := cfg.DB.CreateInvoice(ctx, dbParams)
	if err != nil {
		log.Println("failed to save invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to save invoice: ")
		return
	}

	invoice.ID = dbInvoice.ID
	helpers.RespondWithJson(w, http.StatusCreated, invoice)
}

func (cfg *Config) GetInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	invoiceId := chi.URLParam(r, "id")
	if invoiceId == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "")
		return
	}

	parsedId, err := uuid.Parse(invoiceId)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "error parsing id")
		return
	}
	invoice, err := cfg.DB.GetInvoice(r.Context(), parsedId)
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting invoice")
		return
	}
	// authorization
	user, httpstatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpstatus, err.Error())

	}
	if user.Role != "admin" {
		//  staff must own the invoice
		if user.ID != invoice.UserID {
			helpers.RespondWithError(w, http.StatusUnauthorized, "user not authorize")
			return
		}
	}
	log.Println(invoice)
	log.Println(dbInvoicetoInvoice(invoice))

	helpers.RespondWithJson(w, http.StatusOK, dbInvoicetoInvoice(invoice))
}

func (cfg *Config) GetInvoicesHandler(w http.ResponseWriter, r *http.Request) {
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

	invoices, err := cfg.DB.GetUserInvoices(r.Context(), database.GetUserInvoicesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
		UserID: user.ID,
	})
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting invoice")
		return
	}

	helpers.RespondWithJson(w, http.StatusOK, dbInvoicestoInvoices(invoices))
}

func (cfg *Config) AdminListAllInvoicesHandler(w http.ResponseWriter, r *http.Request) {

	invoices, err := cfg.DB.ListInvoices(r.Context(), database.ListInvoicesParams{
		Offset: 10,
		Limit:  10,
	})
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting invoice")
		return
	}

	helpers.RespondWithJson(w, http.StatusOK, dbInvoicestoInvoices(invoices))
}
