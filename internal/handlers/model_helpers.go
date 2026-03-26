package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/invoice"
)

func dbUserToUser(dbUser database.User) User {
	return User{
		ID: dbUser.ID,

		FirstName:   dbUser.FirstName,
		LastName:    dbUser.LastName,
		Email:       dbUser.Email,
		PhoneNumber: dbUser.PhoneNumber,
		Address:     dbUser.Address,
		Country:     dbUser.Country,
		State:       dbUser.State,
		Role:        dbUser.Role,
		CreatedAt:   dbUser.CreatedAt,
	}

}

func dbInvoicetoInvoice(dbInvoice database.Invoice) invoice.Invoice {
	items := []invoice.Item{}
	err := json.Unmarshal(dbInvoice.Items, &items)
	if err != nil {
		log.Printf("Error unmarshaling items for invoice %s: %v", dbInvoice.ID, err)

	}
	discounts := []invoice.Discount{}
	err = json.Unmarshal(dbInvoice.Discounts, &discounts)
	if err != nil {

		log.Printf("Error unmarshaling discounts for invoice %s: %v", dbInvoice.ID, err)
	}
	total, _ := strconv.ParseFloat(dbInvoice.Total, 64)
	return invoice.Invoice{
		ID:            dbInvoice.ID,
		InvoiceNumber: dbInvoice.InvoiceNumber,
		UserId:        dbInvoice.UserID,
		CustomerName:  dbInvoice.CustomerName,
		CustomerEmail: dbInvoice.CustomerEmail,
		CustomerPhone: dbInvoice.CustomerPhone.String,
		Items:         items,
		Discounts:     discounts,
		Total:         float64(total),
		Notes:         dbInvoice.Notes,
		Status:        dbInvoice.Status,
		UpdatedAt:     dbInvoice.UpdatedAt,
		CreatedAt:     dbInvoice.CreatedAt,
	}

}

func dbInvoicestoInvoices(dbInvoices []database.Invoice) []invoice.Invoice {
	res := []invoice.Invoice{}
	for _, dbInvoice := range dbInvoices {
		res = append(res, dbInvoicetoInvoice(dbInvoice))
	}
	return res
}

func dbServiceToService(dbService database.Service) Service {

	return Service{
		ID:          dbService.ID,
		Name:        dbService.Name,
		Description: dbService.Description,
		Category:    dbService.Category,
		IsActive:    dbService.IsActive,
		IsFeatured:  dbService.IsFeatured,
		Icon:        dbService.Icon,
		Image:       dbService.Image,
		Tags:        dbService.Tags,
		CreatedAt:   dbService.CreatedAt,
	}

}

func dbServicesstoServices(dbServices []database.Service) []Service {
	res := []Service{}
	for _, dbService := range dbServices {
		res = append(res, dbServiceToService(dbService))
	}
	return res
}

func DbQuoteRequestToQuoteRequest(dbReq database.QuoteRequest) QuoteRequest {
	return QuoteRequest{
		ID:          dbReq.ID,
		UserID:      dbReq.UserID,
		ServiceID:   dbReq.ServiceID,
		Description: dbReq.Description,
		Attachments: dbReq.Attachments,
		Status:      QuoteRequestStatus(dbReq.Status),
		CreatedAt:   dbReq.CreatedAt,
		UpdatedAt:   dbReq.UpdatedAt,
	}
}

func DbQuoteRequestsToQuoteRequests(dbReqs []database.QuoteRequest) []QuoteRequest {
	res := make([]QuoteRequest, 0, len(dbReqs))
	for _, req := range dbReqs {
		res = append(res, DbQuoteRequestToQuoteRequest(req))
	}
	return res
}

// Define a small struct for the breakdown items if you want typed data
type QuoteItem struct {
	Item string  `json:"item"`
	Cost float64 `json:"cost"`
}

func DbQuoteToQuote(dbQuote database.Quote) Quote {
	// Parse the decimal string to float64 for the frontend
	amount, _ := strconv.ParseFloat(dbQuote.Amount, 64)
	breakDowns := []QuoteBreakdown{}
	err := json.Unmarshal(dbQuote.Breakdown, &breakDowns)
	if err != nil {

		log.Printf("Error unmarshaling breakdowns for quote %s: %v", dbQuote.ID, err)
	}

	return Quote{
		ID:             dbQuote.ID,
		QuoteRequestID: dbQuote.QuoteRequestID,
		Amount:         fmt.Sprintf("%.2f", amount),
		Breakdown:      breakDowns,
		Notes:          dbQuote.Notes,
		Status:         QuoteStatus(dbQuote.Status),
		ExpiresAt:      dbQuote.ExpiresAt,
		CreatedAt:      dbQuote.CreatedAt,
		UpdatedAt:      dbQuote.UpdatedAt,
	}
}

func DbQuotesToQuotes(dbQuotes []database.Quote) []Quote {
	res := make([]Quote, 0, len(dbQuotes))
	for _, q := range dbQuotes {
		res = append(res, DbQuoteToQuote(q))
	}
	return res
}

func DbQuoteRequestRowToQuoteRequestRow(dbRow database.GetQuoteRequestsRow) GetQuoteRequestsRow {
	return GetQuoteRequestsRow{
		ID:          dbRow.ID,
		UserID:      dbRow.UserID,
		ServiceID:   dbRow.ServiceID,
		Description: dbRow.Description,
		Attachments: dbRow.Attachments,
		Status:      QuoteRequestStatus(dbRow.Status),
		CreatedAt:   dbRow.CreatedAt,
		UpdatedAt:   dbRow.UpdatedAt,
		UserEmail:   dbRow.UserEmail,
		UserName:    dbRow.UserName,
		ServiceName: dbRow.ServiceName,
	}
}

func DbQuoteRequestRowsToQuoteRequestsRow(dbRows []database.GetQuoteRequestsRow) []GetQuoteRequestsRow {
	res := make([]GetQuoteRequestsRow, 0, len(dbRows))
	for _, row := range dbRows {
		res = append(res, DbQuoteRequestRowToQuoteRequestRow(row))
	}
	return res
}
