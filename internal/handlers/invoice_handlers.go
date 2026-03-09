package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
	invoicepackage "github.com/muhammadolammi/n3xtbridge_api/internal/invoice"
)

func CreateInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	var input invoicepackage.InvoiceInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}

	invoice := invoicepackage.CreateInvoice(input)

	pdf, err := invoicepackage.GeneratePDF(invoice)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to generate pdf")
		return
	}

	helpers.RespondWithPdf(w, http.StatusOK, pdf, invoice.InvoiceNumber)
}
