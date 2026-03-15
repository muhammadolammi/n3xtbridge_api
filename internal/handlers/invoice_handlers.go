package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
	invoicepackage "github.com/muhammadolammi/n3xtbridge_api/internal/invoice"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
)

func (cfg *Config) CreateInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	var input invoicepackage.InvoiceInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}

	invoice := invoicepackage.CreateInvoice(input)

	// Convert to DB params
	discountStr := ""
	if input.Discount > 0 {
		discountStr = fmt.Sprintf("%.2f", input.Discount)
	}
	totalStr := fmt.Sprintf("%.2f", invoice.Total)

	dbParams := database.CreateInvoiceParams{
		InvoiceNumber: invoice.InvoiceNumber,
		CustomerName:  invoice.CustomerName,
		CustomerEmail: invoice.CustomerEmail,
		CustomerPhone: sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		Discount:      sql.NullString{String: discountStr, Valid: input.Discount > 0},
		Total:         totalStr,
		Notes:         sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	}

	// Save invoice to database
	ctx := context.Background()
	dbInvoice, err := cfg.DB.CreateInvoice(ctx, dbParams)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to save invoice: "+err.Error())
		return
	}

	// Save items to database
	for _, item := range input.Items {
		itemParams := database.CreateItemParams{
			InvoiceID: uuid.NullUUID{UUID: dbInvoice.ID, Valid: true},
			Name:      item.Name,
			Quantity:  int32(item.Quantity),
			Price:     fmt.Sprintf("%.2f", item.Price),
		}
		_, err := cfg.DB.CreateItem(ctx, itemParams)
		if err != nil {
			helpers.RespondWithError(w, http.StatusInternalServerError, "failed to save item: "+err.Error())
			return
		}
	}

	// Generate PDF with DB invoice data (including ID)
	pdf, err := invoicepackage.GeneratePDF(invoice)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to generate pdf")
		return
	}

	helpers.RespondWithPdf(w, http.StatusOK, pdf, invoice.InvoiceNumber)
}
