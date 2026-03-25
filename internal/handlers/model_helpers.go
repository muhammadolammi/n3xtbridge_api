package handlers

import (
	"encoding/json"
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
