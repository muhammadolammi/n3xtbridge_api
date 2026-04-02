package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/shared"
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

func DbDiscountToDiscount(dbDiscount DBDiscount) shared.Discount {
	amount, _ := strconv.ParseFloat(dbDiscount.Amount, 64)

	return shared.Discount{
		Name:        dbDiscount.Name,
		Amount:      amount,
		Description: dbDiscount.Description,
		Type:        dbDiscount.Type,
		ItemName:    dbDiscount.ItemName,
	}
}

func DbDiscountsToDiscounts(dbDiscounts []DBDiscount) []shared.Discount {
	res := []shared.Discount{}
	for _, dbDiscount := range dbDiscounts {
		res = append(res, DbDiscountToDiscount(dbDiscount))
	}
	return res
}

func DbItemToItem(dbItem DBItem) shared.Item {

	price, _ := strconv.ParseFloat(dbItem.Price, 64)

	return shared.Item{
		Name:        dbItem.Name,
		Price:       price,
		Quantity:    dbItem.Quantity,
		Description: dbItem.Description,
	}
}
func DbItemsToItems(dbItems []DBItem) []shared.Item {
	res := []shared.Item{}
	for _, dbItem := range dbItems {
		res = append(res, DbItemToItem(dbItem))
	}
	return res
}
func dbInvoicetoInvoice(dbInvoice database.Invoice) Invoice {
	dbItems := []DBItem{}
	err := json.Unmarshal(dbInvoice.Items, &dbItems)
	if err != nil {
		log.Printf("Error unmarshaling items for invoice %s: %v", dbInvoice.ID, err)

	}
	dbDiscounts := []DBDiscount{}
	err = json.Unmarshal(dbInvoice.Discounts, &dbDiscounts)
	if err != nil {

		log.Printf("Error unmarshaling discounts for invoice %s: %v", dbInvoice.ID, err)
	}
	items := DbItemsToItems(dbItems)
	discounts := DbDiscountsToDiscounts(dbDiscounts)
	total, _ := strconv.ParseFloat(dbInvoice.Total, 64)

	return Invoice{
		ID:             dbInvoice.ID,
		QuoteID:        dbInvoice.QuoteID.UUID,
		InvoiceNumber:  dbInvoice.InvoiceNumber,
		UserId:         dbInvoice.UserID,
		CustomerName:   dbInvoice.CustomerName,
		CustomerEmail:  dbInvoice.CustomerEmail,
		CustomerPhone:  dbInvoice.CustomerPhone.String,
		Items:          items,
		Discounts:      discounts,
		Total:          float64(total),
		Notes:          dbInvoice.Notes,
		Status:         dbInvoice.Status,
		PaymentToken:   dbInvoice.PaymentToken,
		UpdatedAt:      dbInvoice.UpdatedAt,
		CreatedAt:      dbInvoice.CreatedAt,
		DeletedAt:      dbInvoice.DeletedAt,
		ReminderSentAt: dbInvoice.ReminderSentAt.Time,
	}

}

func dbInvoicestoInvoices(dbInvoices []database.Invoice) []Invoice {
	res := []Invoice{}
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
		PromoIDs:    dbService.ActivePromoIds,
		CreatedAt:   dbService.CreatedAt,
	}

}

func dbServicesToServices(dbServices []database.Service) []Service {
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
		PromoIDs:    dbReq.PromoIds,
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

func DbQuoteToQuote(dbQuote database.Quote) Quote {
	// Parse the decimal string to float64 for the frontend
	amount, _ := strconv.ParseFloat(dbQuote.Amount, 64)
	dbBreakDowns := []DBItem{}
	err := json.Unmarshal(dbQuote.Breakdown, &dbBreakDowns)
	if err != nil {

		log.Printf("Error unmarshaling breakdowns for quote %s: %v", dbQuote.ID, err)
	}
	dbDiscounts := []DBDiscount{}
	err = json.Unmarshal(dbQuote.Discounts, &dbDiscounts)
	if err != nil {

		log.Printf("Error unmarshaling discounts for quote %s: %v", dbQuote.ID, err)
	}
	breakDowns := DbItemsToItems(dbBreakDowns)
	discounts := DbDiscountsToDiscounts(dbDiscounts)

	return Quote{
		ID:             dbQuote.ID,
		UserID:         dbQuote.UserID,
		QuoteRequestID: dbQuote.QuoteRequestID,
		Amount:         fmt.Sprintf("%.2f", amount),
		Breakdown:      breakDowns,
		Discounts:      discounts,
		PromoIDs:       dbQuote.PromoIds,

		Notes:     dbQuote.Notes,
		Status:    QuoteStatus(dbQuote.Status),
		ExpiresAt: dbQuote.ExpiresAt,
		CreatedAt: dbQuote.CreatedAt,
		UpdatedAt: dbQuote.UpdatedAt,
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
		PromoIDs:    dbRow.PromoIds,
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

func DbUserQuoteRequestRowToUserQuoteRequestRow(dbRow database.GetUserQuoteRequestsRow) GetUserQuoteRequestsRow {
	return GetUserQuoteRequestsRow{
		ID:          dbRow.ID,
		UserID:      dbRow.UserID,
		QuoteID:     dbRow.QuoteID.UUID,
		ServiceID:   dbRow.ServiceID,
		Description: dbRow.Description,
		Attachments: dbRow.Attachments,
		PromoIDs:    dbRow.PromoIds,

		Status:    QuoteRequestStatus(dbRow.Status),
		CreatedAt: dbRow.CreatedAt,
		UpdatedAt: dbRow.UpdatedAt,

		ServiceName: dbRow.ServiceName,
	}
}

func DbUserQuoteRequestRowsToUserQuoteRequestsRow(dbRows []database.GetUserQuoteRequestsRow) []GetUserQuoteRequestsRow {
	res := make([]GetUserQuoteRequestsRow, 0, len(dbRows))
	for _, row := range dbRows {
		res = append(res, DbUserQuoteRequestRowToUserQuoteRequestRow(row))
	}
	return res
}

func DbUserQuotesWithServiceRowToUserQuotesWithServiceRow(dbQuote database.GetUserQuotesWithServiceRow) GetUserQuotesWithServiceRow {
	// Parse the decimal string to float64 for the frontend
	amount, _ := strconv.ParseFloat(dbQuote.Amount, 64)
	dbBreakDowns := []DBItem{}
	err := json.Unmarshal(dbQuote.Breakdown, &dbBreakDowns)
	if err != nil {

		log.Printf("Error unmarshaling breakdowns for quote %s: %v", dbQuote.ID, err)
	}

	dbDiscounts := []DBDiscount{}
	err = json.Unmarshal(dbQuote.Discounts, &dbDiscounts)
	if err != nil {
		log.Printf("Error unmarshaling discounts for quote %s: %v", dbQuote.ID, err)
	}
	breakDowns := DbItemsToItems(dbBreakDowns)
	discounts := DbDiscountsToDiscounts(dbDiscounts)

	return GetUserQuotesWithServiceRow{
		ID:             dbQuote.ID,
		UserID:         dbQuote.UserID,
		QuoteRequestID: dbQuote.QuoteRequestID,
		Amount:         fmt.Sprintf("%.2f", amount),
		Breakdown:      breakDowns,
		Discounts:      discounts,
		PromoIDs:       dbQuote.PromoIds,

		Notes:       dbQuote.Notes,
		Status:      QuoteStatus(dbQuote.Status),
		ExpiresAt:   dbQuote.ExpiresAt,
		CreatedAt:   dbQuote.CreatedAt,
		UpdatedAt:   dbQuote.UpdatedAt,
		ServiceIcon: dbQuote.ServiceIcon,
		ServiceName: dbQuote.ServiceName,
		ServiceID:   dbQuote.ServiceID,
	}
}

func DbUserQuotesWithServiceRowsToUserQuotesWithServiceRows(dbQuotes []database.GetUserQuotesWithServiceRow) []GetUserQuotesWithServiceRow {
	res := make([]GetUserQuotesWithServiceRow, 0, len(dbQuotes))
	for _, q := range dbQuotes {
		res = append(res, DbUserQuotesWithServiceRowToUserQuotesWithServiceRow(q))
	}
	return res
}

// invoice
func CalculateInvoiceTotal(dbItems []DBItem, dbDiscounts []DBDiscount) float64 {
	var itemsTotal float64
	var discountsTotal float64
	items := DbItemsToItems(dbItems)
	discounts := DbDiscountsToDiscounts(dbDiscounts)

	for _, item := range items {
		// price, _ := strconv.ParseFloat(item.Price, 64)

		itemsTotal += float64(item.Quantity) * item.Price
	}
	for _, discount := range discounts {

		discountsTotal += discount.Amount
	}
	total := itemsTotal - discountsTotal

	return total
}

func GenerateInvoiceNumber() string {
	year := time.Now().Year()

	counter := time.Now().Unix() % 100000

	return fmt.Sprintf("INV-%d-%05d", year, counter)
}

func dbPromoToPromo(dbPromo database.Promotion) Promotion {
	dbBreakdowns := []DBDiscount{}
	err := json.Unmarshal(dbPromo.Breakdown, &dbBreakdowns)
	if err != nil {
		log.Printf("Error unmarshaling breakdown for Promo %s: %v", dbPromo.ID, err)
	}
	discounts := DbDiscountsToDiscounts(dbBreakdowns)

	return Promotion{
		ID:          dbPromo.ID,
		Code:        dbPromo.Code,
		Name:        dbPromo.Name,
		Description: dbPromo.Description,
		Breakdown:   discounts,
		IsActive:    dbPromo.IsActive.Bool,
		StartsAt:    dbPromo.StartsAt.Time,
		ExpiresAt:   dbPromo.ExpiresAt.Time,
		CreatedAt:   dbPromo.CreatedAt.Time,
	}
}

func dbPromosToPromos(dbPromos []database.Promotion) []Promotion {
	res := make([]Promotion, 0, len(dbPromos))
	for _, p := range dbPromos {
		res = append(res, dbPromoToPromo(p))
	}
	return res
}
