package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
	payment "github.com/muhammadolammi/n3xtbridge_api/internal/payments"
)

func (cfg *Config) InitializePaymentHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get Invoice ID from URL or Body
	invoiceID := chi.URLParam(r, "id")
	guestToken := r.URL.Query().Get("token") // Get token from query string

	inv, err := cfg.DBQueries.GetInvoice(r.Context(), uuid.MustParse(invoiceID))
	if err != nil {
		helpers.RespondWithError(w, http.StatusNotFound, "Invoice not found")
		return
	}

	isAuthorized := false

	// lets make some verification
	user, _, err := cfg.getUserFromReq(r)
	if err == nil && user.ID != uuid.Nil {
		// If logged in, email must match
		if user.Email == inv.CustomerEmail {
			isAuthorized = true
		} else {
			helpers.RespondWithError(w, http.StatusUnauthorized, "This invoice does not belong to your account.")
			return
		}
	}
	if !isAuthorized {
		if guestToken != "" && guestToken == inv.PaymentToken {
			isAuthorized = true
			log.Printf("Guest access granted for Invoice %s via token", inv.InvoiceNumber)
		}
	}

	if !isAuthorized {
		log.Println("Invalid session or payment token on init payment for invoice : ", inv.InvoiceNumber)
		helpers.RespondWithError(w, http.StatusUnauthorized, "Invalid session or payment token.")
		return
	}
	if inv.Status == "paid" {
		helpers.RespondWithError(w, http.StatusConflict, "This invoice has already been settled.")
		return
	}

	existingPayment, err := cfg.DBQueries.GetLatestPendingPayment(r.Context(), inv.ID)
	if err == nil {

		log.Printf("Superseding old payment attempt: %s", existingPayment.Reference)
		_ = cfg.DBQueries.UpdatePaymentStatus(r.Context(), database.UpdatePaymentStatusParams{
			Reference: existingPayment.Reference,
			Status:    "cancelled",
		})
	}
	paystackResp, httpStatus, err := cfg.getInvoiceCheckoutURL(inv, r.Context())
	if err != nil {
		helpers.RespondWithError(w, httpStatus, err.Error())
		return
	}

	// 6. Send the Paystack URL to frontend
	helpers.RespondWithJson(w, http.StatusOK, map[string]string{
		"checkout_url": paystackResp.Data.AuthorizationURL,
		"invoice_id":   invoiceID,
	})
}

func (cfg *Config) PaystackWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Read the raw body (we need it for both verification and forwarding)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 2. Verify Signature (Always verify at the entry point)
	hash := hmac.New(sha512.New, []byte(cfg.PaystackSecret))
	hash.Write(body)
	expectedSignature := hex.EncodeToString(hash.Sum(nil))
	receivedSignature := r.Header.Get("x-paystack-signature")

	if receivedSignature != expectedSignature {
		log.Println("⚠️ INVALID WEBHOOK SIGNATURE")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 3. Parse Event to check the Reference
	var event payment.WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reference := event.Data.Reference

	// 4. ROUTING LOGIC
	if strings.HasPrefix(reference, "N3XT-") {
		// --- PROCESS LOCALLY FOR N3XTBRIDGE ---
		if event.Event == "charge.success" {
			log.Printf("💰 N3xtbridge Payment Success: %s", reference)
			err := helpers.FinalizePayment(
				r.Context(),
				cfg.DBConn,
				cfg.DBQueries,
				reference,
				fmt.Sprintf("%d", event.Data.ID),
			)

			if err != nil {
				log.Println("❌ Error finalizing local payment:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

	} else {
		// --- FORWARD TO JOBMATCH ---
		log.Printf("🚀 Forwarding Webhook to JobMatch: %s", reference)
		jobMatchURL := "https://jobmatchapi.qtechconsults.com/api/webhook/paystack"

		// Use a background context so the Paystack response isn't delayed by the forward
		go forwardWebhook(jobMatchURL, body, receivedSignature)
	}

	// Always return 200 OK to Paystack immediately
	// Always return 200 OK to Paystack immediately
	w.WriteHeader(http.StatusOK)
}

func (cfg *Config) VerifyPaymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	reference := chi.URLParam(r, "ref")
	if reference == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Missing reference")
		return

	}

	payment, err := cfg.DBQueries.GetPaymentByReference(r.Context(), reference)
	if err != nil {
		// If not found, it might just be the database lag, return pending
		helpers.RespondWithJson(w, http.StatusOK, map[string]string{"status": "pending"})
		return
	}

	helpers.RespondWithJson(w, http.StatusOK, map[string]string{
		"status": string(payment.Status),
	})
}

func forwardWebhook(targetURL string, body []byte, signature string) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("❌ Failed to create forward request: %v", err)
		return
	}

	// Re-attach the critical Paystack headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-paystack-signature", signature)
	req.Header.Set("User-Agent", "N3xtbridge-Forwarder/1.0")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to forward webhook to %s: %v", targetURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ Forwarded webhook returned status: %d", resp.StatusCode)
	} else {
		log.Printf("✅ Webhook forwarded successfully to JobMatch")
	}
}
