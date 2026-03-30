package helpers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
)

type AcceptQuoteAndCreateInvoiceParams struct {
	Quote         *database.Quote
	Customer      *database.User
	Db            *sql.DB
	Queries       *database.Queries
	InvoiceNumber string
	AdminID       uuid.UUID
}
type CreatePromotionAndLinkWithServiceParam struct {
	Db                   *sql.DB
	Queries              *database.Queries
	CreatePromotionParam database.CreatePromotionParams
	Service              database.Service
}

func AcceptQuoteAndCreateInvoice(ctx context.Context, params AcceptQuoteAndCreateInvoiceParams) error {
	// 1. Start the transaction
	tx, err := params.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Ensure rollback happens if any error occurs (or if we forget to commit)
	defer tx.Rollback()

	// 2. Bind the generated queries to this transaction
	qtx := params.Queries.WithTx(tx)

	// 3. Operation A: Accept the quote
	err = qtx.UpdateQuoteStatus(ctx, database.UpdateQuoteStatusParams{
		ID:     params.Quote.ID,
		Status: "accepted",
	})
	if err != nil {
		return fmt.Errorf("failed to accept quote: %w", err)
	}

	// 4. Operation B: Create the invoice
	// If this fails, the 'defer tx.Rollback()' ensures Operation A is undone

	_, err = qtx.CreateInvoiceWithQuote(ctx, database.CreateInvoiceWithQuoteParams{
		InvoiceNumber: params.InvoiceNumber,
		CustomerName:  fmt.Sprintf("%s %s", params.Customer.FirstName, params.Customer.LastName),
		CustomerEmail: params.Customer.Email,
		CustomerPhone: params.Customer.PhoneNumber,
		Total:         params.Quote.Amount,
		Notes:         params.Quote.Notes,
		Items:         params.Quote.Breakdown,
		Discounts:     params.Quote.Discounts,
		UserID:        params.AdminID,
		QuoteID: uuid.NullUUID{
			Valid: true,
			UUID:  params.Quote.ID,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// 5. Commit everything if both succeeded
	return tx.Commit()
}

func FinalizePayment(ctx context.Context, db *sql.DB, queries *database.Queries, reference string, externalID string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := queries.WithTx(tx)

	// 1. Update Payment Status
	payment, err := qtx.GetPaymentByReference(ctx, reference)

	if err != nil {
		return fmt.Errorf("payment ref not found: %w", err)
	}

	err = qtx.UpdatePaymentStatus(ctx, database.UpdatePaymentStatusParams{
		Reference:  reference,
		Status:     "success",
		ExternalID: sql.NullString{String: fmt.Sprintf("%s", externalID), Valid: true},
	})
	if err != nil {
		return err
	}

	// 2. Update Invoice Status
	err = qtx.MarkInvoiceAsPaid(ctx, payment.InvoiceID)
	if err != nil {
		return err
	}
	// update invoice quote as paid if invoice if for a quote
	inv, err := qtx.GetInvoice(ctx, payment.InvoiceID)
	if err != nil {
		return err
	}
	if inv.QuoteID.Valid {
		err = qtx.UpdateQuoteStatus(ctx, database.UpdateQuoteStatusParams{
			ID:     inv.QuoteID.UUID,
			Status: "paid",
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func CreatePromotionAndLinkWithService(ctx context.Context, params CreatePromotionAndLinkWithServiceParam) (database.Promotion, error) {
	// 1. Start the transaction
	tx, err := params.Db.BeginTx(ctx, nil)
	if err != nil {
		return database.Promotion{}, err
	}

	// Ensure rollback happens if any error occurs (or if we forget to commit)
	defer tx.Rollback()

	// 2. Bind the generated queries to this transaction
	qtx := params.Queries.WithTx(tx)
	promo, err := qtx.CreatePromotion(ctx, params.CreatePromotionParam)
	if err != nil {
		return database.Promotion{}, fmt.Errorf("failed to create  promotion: %w", err)
	}

	// 3. Operation A: Accept the quote
	currentPromos := params.Service.ActivePromoIds

	currentPromos = append(currentPromos, promo.ID.String())

	err = qtx.UpdateServicePromo(ctx, database.UpdateServicePromoParams{
		ID:             params.Service.ID,
		ActivePromoIds: currentPromos,
	})
	if err != nil {
		return database.Promotion{}, fmt.Errorf("failed to update service promo: %w", err)
	}

	// 5. Commit everything if both succeeded
	return promo, tx.Commit()
}
