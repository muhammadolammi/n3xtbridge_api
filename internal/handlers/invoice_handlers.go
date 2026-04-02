package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

func (cfg *Config) CreateInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	type InvoiceInput struct {
		CustomerName  string       `json:"customer_name"`
		CustomerEmail string       `json:"customer_email"`
		CustomerPhone string       `json:"customer_phone"`
		Items         []DBItem     `json:"items"`
		Discounts     []DBDiscount `json:"discounts"`
		Notes         string       `json:"notes"`
	}
	var input InvoiceInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}
	user, httpstatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpstatus, err.Error())
		return

	}

	totalStr := fmt.Sprintf("%.2f", CalculateInvoiceTotal(input.Items, input.Discounts))
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
	paymentToken := base64.URLEncoding.EncodeToString([]byte(uuid.New().String()))
	dbParams := database.CreateInvoiceParams{
		InvoiceNumber: GenerateInvoiceNumber(),
		UserID:        user.ID,
		CustomerName:  input.CustomerName,
		CustomerEmail: input.CustomerEmail,
		CustomerPhone: sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		Total:         totalStr,
		Notes:         input.Notes,
		Discounts:     jsonBDiscounts,
		Items:         jsonBItems,
		PaymentToken:  paymentToken,
	}

	dbInvoice, err := cfg.DBQueries.CreateInvoice(r.Context(), dbParams)
	if err != nil {
		log.Println("failed to save invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to save invoice: ")
		return
	}

	log.Printf("Admin %v created invoice: %s", user.Email, dbInvoice.InvoiceNumber)
	inv := dbInvoicetoInvoice(dbInvoice)

	// // send a mail to the customer
	// go func() {
	// 	err := cfg.EmailSender.SendInvoice(mailer.InvoiceData{
	// 		InvoiceNumber: inv.InvoiceNumber,
	// 		CustomerName:  inv.CustomerName,
	// 		CustomerEmail: inv.CustomerEmail,
	// 		Items:         inv.Items,
	// 		Discounts:     inv.Discounts,
	// 		Total:         inv.Total,
	// 		Notes:         input.Notes,
	// 		PaymentLink:   invoiceURL,
	// 	})
	// 	if err != nil {
	// 		log.Println("error sending invoice email: " + err.Error())
	// 	} else {
	// 		log.Printf("Invoice email sent to %s for invoice %s", inv.CustomerEmail, inv.InvoiceNumber)
	// 	}

	// }()
	go sendInvoiceEmail(cfg, inv)

	helpers.RespondWithJson(w, http.StatusCreated, dbInvoicetoInvoice(dbInvoice))
}

func (cfg *Config) PublicGetInvoiceHandler(w http.ResponseWriter, r *http.Request) {
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
	invoice, err := cfg.DBQueries.GetInvoice(r.Context(), parsedId)
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting invoice")
		return
	}

	helpers.RespondWithJson(w, http.StatusOK, dbInvoicetoInvoice(invoice))
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
	invoice, err := cfg.DBQueries.GetInvoice(r.Context(), parsedId)
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting invoice")
		return
	}
	// authorization
	user, httpstatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpstatus, err.Error())
		return

	}
	if user.Role != "admin" {
		//  staff must own the invoice
		if user.ID != invoice.UserID && invoice.CustomerEmail != user.Email {

			helpers.RespondWithError(w, http.StatusUnauthorized, "user not authorize")
			return
		}
	}

	helpers.RespondWithJson(w, http.StatusOK, dbInvoicetoInvoice(invoice))
}

func (cfg *Config) GetQuoteInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	quoteID := chi.URLParam(r, "id")
	if quoteID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "")
		return
	}

	parsedId, err := uuid.Parse(quoteID)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "error parsing id")
		return
	}
	invoice, err := cfg.DBQueries.GetInvoiceByQuoteID(r.Context(), uuid.NullUUID{Valid: true, UUID: parsedId})
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting invoice")
		return
	}
	// authorization
	user, httpstatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpstatus, err.Error())
		return

	}
	if user.Role != "admin" {
		//  staff must own the invoice
		if user.ID != invoice.UserID && invoice.CustomerEmail != user.Email {

			helpers.RespondWithError(w, http.StatusUnauthorized, "user not authorize")
			return
		}
	}

	helpers.RespondWithJson(w, http.StatusOK, dbInvoicetoInvoice(invoice))
}

func (cfg *Config) GetWorkersCreatedInvoicesHandler(w http.ResponseWriter, r *http.Request) {
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

	invoices, err := cfg.DBQueries.GetWorkersCreatedInvoices(r.Context(), database.GetWorkersCreatedInvoicesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
		UserID: user.ID,
	})
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting invoice")
		return
	}

	count, err := cfg.DBQueries.CountWorkersCreatedInvoices(r.Context(), user.ID)
	if err != nil {
		log.Println("DB ERROR error getting user invoices count: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting user invoices  count")
		return
	}
	res := struct {
		Invoices []Invoice `json:"invoices"`
		Total    int64     `json:"total"`
	}{
		Invoices: dbInvoicestoInvoices(invoices),
		Total:    count,
	}
	helpers.RespondWithJson(w, http.StatusOK, res)
}
func (cfg *Config) GetCustomerInvoicesHandler(w http.ResponseWriter, r *http.Request) {
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

	invoices, err := cfg.DBQueries.GetCustomerInvoices(r.Context(), database.GetCustomerInvoicesParams{
		Limit:         int32(limit),
		Offset:        int32(offset),
		CustomerEmail: user.Email,
	})
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting invoice")
		return
	}

	count, err := cfg.DBQueries.CountCustomersInvoices(r.Context(), user.Email)
	if err != nil {
		log.Println("DB ERROR error getting user invoices count: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting user invoices  count")
		return
	}
	res := struct {
		Invoices []Invoice `json:"invoices"`
		Total    int64     `json:"total"`
	}{
		Invoices: dbInvoicestoInvoices(invoices),
		Total:    count,
	}
	helpers.RespondWithJson(w, http.StatusOK, res)
}

func (cfg *Config) AdminListAllInvoicesHandler(w http.ResponseWriter, r *http.Request) {
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
	invoices, err := cfg.DBQueries.ListInvoices(r.Context(), database.ListInvoicesParams{
		Offset: int32(offset),
		Limit:  int32(limit),
	})
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting invoice")
		return
	}

	count, err := cfg.DBQueries.CountInvoices(r.Context())
	if err != nil {
		log.Println("DB ERROR error getting  invoices count: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting  invoices  count")
		return
	}
	res := struct {
		Invoices []Invoice `json:"invoices"`
		Total    int64     `json:"total"`
	}{
		Invoices: dbInvoicestoInvoices(invoices),
		Total:    count,
	}
	log.Println(res.Invoices)
	helpers.RespondWithJson(w, http.StatusOK, res)
}

func (cfg *Config) AdminSendInvoiceEmailHandler(w http.ResponseWriter, r *http.Request) {
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
	invoice, err := cfg.DBQueries.GetInvoice(r.Context(), parsedId)
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting invoice")
		return
	}

	inv := dbInvoicetoInvoice(invoice)

	// send a mail to the customer
	// go func() {
	// 	err := cfg.EmailSender.SendInvoice(mailer.InvoiceData{
	// 		InvoiceNumber: inv.InvoiceNumber,
	// 		CustomerName:  inv.CustomerName,
	// 		CustomerEmail: inv.CustomerEmail,
	// 		Items:         inv.Items,
	// 		Discounts:     inv.Discounts,
	// 		Total:         inv.Total,
	// 		Notes:         inv.Notes,
	// 		PaymentLink:   invoiceURL,
	// 		Date:          inv.CreatedAt.String(),
	// 	})
	// 	if err != nil {
	// 		log.Println("error sending invoice email reminder: " + err.Error())
	// 	} else {
	// 		// UPDATE
	// 		log.Printf("Invoice email reminder sent to %s for invoice %s", inv.CustomerEmail, inv.InvoiceNumber)
	// 	}
	// }()
	go sendInvoiceEmail(cfg, inv)

	helpers.RespondWithJson(w, http.StatusOK, map[string]string{"message": "reminder email dispatched"})
}
