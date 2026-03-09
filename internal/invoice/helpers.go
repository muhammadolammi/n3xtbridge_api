package invoice

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func CalculateTotal(input InvoiceInput) float64 {

	var itemsTotal float64

	for _, item := range input.Items {
		itemsTotal += float64(item.Quantity) * item.Price
	}
	total := itemsTotal - input.Discount

	return total
}

func CreateInvoice(input InvoiceInput) Invoice {
	total := CalculateTotal(input)
	return Invoice{
		ID:            uuid.NewString(),
		InvoiceNumber: GenerateInvoiceNumber(),
		CustomerName:  input.CustomerName,
		CustomerEmail: input.CustomerEmail,
		CustomerPhone: input.CustomerPhone,
		Items:         input.Items,
		Discount:      input.Discount,
		Total:         total,
		CreatedAt:     time.Now(),
	}
}

func GenerateInvoiceNumber() string {
	year := time.Now().Year()

	counter := time.Now().Unix() % 100000

	return fmt.Sprintf("INV-%d-%05d", year, counter)
}

func formatDate(t time.Time) string {
	return t.Format("02 Jan 2006")
}

func formatMoney(n float64) string {
	s := fmt.Sprintf("%.0f", n)

	nInt, _ := strconv.ParseInt(s, 10, 64)

	return fmt.Sprintf("₦%s", humanize(nInt))
}

func humanize(n int64) string {
	in := strconv.FormatInt(n, 10)
	out := ""

	for i, c := range reverse(in) {
		if i != 0 && i%3 == 0 {
			out = "," + out
		}
		out = string(c) + out
	}

	return out
}

func reverse(s string) string {
	runes := []rune(s)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

func itemTotal(q int, p float64) float64 {
	return float64(q) * p
}
