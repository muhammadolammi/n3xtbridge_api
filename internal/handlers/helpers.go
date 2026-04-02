package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/mailer"
	payment "github.com/muhammadolammi/n3xtbridge_api/internal/payments"
)

func (cfg *Config) getUserFromReq(r *http.Request) (database.User, int, error) {

	// authorization
	userIDStr, ok := r.Context().Value("user_id").(string)
	if !ok {
		return database.User{}, http.StatusUnauthorized, errors.New("user not found in context")

	}
	parsedID, err := uuid.Parse(userIDStr)
	if err != nil {
		return database.User{}, http.StatusInternalServerError, errors.New("error parsing user id")

	}
	user, err := cfg.DBQueries.GetUserByID(r.Context(), parsedID)
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		return database.User{}, http.StatusUnauthorized, errors.New("user not authenticated")

	}
	return user, http.StatusOK, nil
}

func validatePassword(password string) error {
	var (
		hasUpper  bool
		hasLower  bool
		hasSymbol bool
	)

	if len(password) < 10 {
		return fmt.Errorf("password must be at least 10 characters long")
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	if !hasUpper || !hasLower || !hasSymbol {
		return fmt.Errorf("password must include uppercase, lowercase, and a symbol")
	}

	return nil
}

// create a pending payment record and init with paystack for a checkout url

func (cfg *Config) getInvoiceCheckoutURL(inv database.Invoice, ctx context.Context) (*payment.TransactionInitResponse, int, error) {

	// Format: N3XT-INV-UUID-TIMESTAMP
	reference := fmt.Sprintf("N3XT-%s-%d", inv.InvoiceNumber, time.Now().Unix())

	// This is the "Resiliency" step. If they fail to pay, we have the record.
	_, err := cfg.DBQueries.CreatePayment(ctx, database.CreatePaymentParams{
		InvoiceID: inv.ID,
		Amount:    inv.Total,
		Reference: reference,
		Status:    "pending",
	})
	if err != nil {
		// helpers.RespondWithError(w, http.StatusInternalServerError, "Paystack init failed")
		log.Printf("DB ERROR  initializing payment: %v", err)
		return &payment.TransactionInitResponse{}, http.StatusInternalServerError, fmt.Errorf("ERROR creating payment record: ")
	}

	// 5. Initialize with Paystack
	total, _ := strconv.ParseFloat(inv.Total, 64)
	callBackUrlBase := getFrontendBaseURL(cfg.IsProd)

	paystackResp, err := cfg.Paystack.InitializeTransaction(payment.TransactionInitRequest{
		Email:     inv.CustomerEmail,
		Amount:    int64(total * 100), // Convert Naira to Kobo
		Reference: reference,
		Currency:  "NGN",
		Callback:  fmt.Sprintf("%s/payment-success", callBackUrlBase),
	})

	if err != nil {
		// helpers.RespondWithError(w, http.StatusInternalServerError, "Paystack init failed")
		log.Printf("Paystack ERROR  initializing payment: %v", err)
		return &payment.TransactionInitResponse{}, http.StatusInternalServerError, fmt.Errorf("Paystack ERROR: %v", err)
	}
	return paystackResp, http.StatusOK, nil
}

func getFrontendBaseURL(isProd bool) string {
	if isProd {
		return "https://n3xtbridge.com"
	}
	return "http://localhost:5173"
}

func sendInvoiceEmail(cfg *Config, inv Invoice) {
	frontendBaseUrl := getFrontendBaseURL(cfg.IsProd)
	invoiceURL := fmt.Sprintf("%s/invoice/%s?token=%s", frontendBaseUrl, inv.ID.String(), inv.PaymentToken)
	err := cfg.EmailSender.SendInvoice(mailer.InvoiceData{
		InvoiceNumber: inv.InvoiceNumber,
		CustomerName:  inv.CustomerName,
		CustomerEmail: inv.CustomerEmail,
		Items:         inv.Items,
		Discounts:     inv.Discounts,
		Total:         inv.Total,
		Notes:         inv.Notes,
		PaymentLink:   invoiceURL,
		Date:          inv.CreatedAt.String(),
	})
	if err != nil {
		log.Println("error sending invoice email reminder: " + err.Error())
	} else {

		_ = cfg.DBQueries.UpdateInvoiceReminderSentAt(context.Background(), inv.ID)
		log.Printf("Invoice email reminder sent to %s for invoice %s", inv.CustomerEmail, inv.InvoiceNumber)
	}
}
