package invoice

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func GeneratePDF(inv Invoice) ([]byte, error) {
	// 1. Define custom functions (like 'mul') for the template
	funcMap := template.FuncMap{
		"mul": func(qty int, price float64) float64 {
			return float64(qty) * price
		},
		"money":     formatMoney,
		"itemTotal": itemTotal,
		"date":      formatDate,
	}

	// 2. Parse and execute the template with the FuncMap
	tmpl, err := template.New("invoice.html").Funcs(funcMap).ParseFiles("templates/invoice.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var htmlBuilder strings.Builder
	if err := tmpl.Execute(&htmlBuilder, inv); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// 3. Setup Chromedp context (Headless is required for PDF)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var pdf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 4. Set the HTML content directly to the page
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			err = page.SetDocumentContent(frameTree.Frame.ID, htmlBuilder.String()).Do(ctx)
			if err != nil {
				return err
			}

			// 5. Generate the PDF
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).   // A4 Width in inches
				WithPaperHeight(11.69). // A4 Height in inches
				Do(ctx)
			if err != nil {
				return err
			}
			pdf = buf
			return nil
		}),
	)

	return pdf, err
}
