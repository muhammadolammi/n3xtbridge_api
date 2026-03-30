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
	"strconv"
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

	// 2. Fetch Invoice from DB to get total
	inv, err := cfg.DBQueries.GetInvoice(r.Context(), uuid.MustParse(invoiceID))
	if err != nil {
		helpers.RespondWithError(w, http.StatusNotFound, "Invoice not found")
		return
	}
	// lets make some verification
	user, httpStatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpStatus, err.Error())
		return
	}
	if user.Email != inv.CustomerEmail {
		helpers.RespondWithError(w, http.StatusUnauthorized, "trying to pay another user's invoice")
		// ahh aha hahha hahhh we should probably leave this, you know :)
		return
	}
	if inv.Status == "paid" {
		helpers.RespondWithError(w, http.StatusConflict, "This invoice has already been settled.")
		return
	}

	existingPayment, err := cfg.DBQueries.GetLatestPendingPayment(r.Context(), inv.ID)
	if err == nil {
		// If the payment is recent (e.g., < 1 hour), we could technically
		// redirect them back, BUT Paystack links expire.
		// Safer approach: If they have a pending one, we let them create a
		// NEW reference but we mark the OLD one as 'cancelled'.
		log.Printf("Superseding old payment attempt: %s", existingPayment.Reference)
		_ = cfg.DBQueries.UpdatePaymentStatus(r.Context(), database.UpdatePaymentStatusParams{
			Reference: existingPayment.Reference,
			Status:    "cancelled",
		})
	}

	// 3. Generate a unique Reference (Very important for Paystack)
	// Format: N3XT-INV-UUID-TIMESTAMP
	reference := fmt.Sprintf("N3XT-%s-%d", inv.InvoiceNumber, time.Now().Unix())

	// 4. Create PENDING payment in your DB (SQLC query needed here)
	// This is the "Resiliency" step. If they fail to pay, we have the record.
	_, err = cfg.DBQueries.CreatePayment(r.Context(), database.CreatePaymentParams{
		InvoiceID: inv.ID,
		Amount:    inv.Total,
		Reference: reference,
		Status:    "pending",
	})

	// 5. Initialize with Paystack
	total, _ := strconv.ParseFloat(inv.Total, 64)
	callBackUrlBase := "http://localhost:5173"
	if cfg.IsProd {
		callBackUrlBase = "https://n3xtbridge.com"

	}

	paystackResp, err := cfg.Paystack.InitializeTransaction(payment.TransactionInitRequest{
		Email:     inv.CustomerEmail,
		Amount:    int64(total * 100), // Convert Naira to Kobo
		Reference: reference,
		Currency:  "NGN",
		Callback:  fmt.Sprintf("%s/dashboard/payment-success", callBackUrlBase),
	})

	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "Paystack init failed")
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
