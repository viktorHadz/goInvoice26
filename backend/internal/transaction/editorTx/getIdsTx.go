package editorTx

import (
	"context"
	"fmt"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

func QueryInvoiceBookPage(
	a *app.App,
	ctx context.Context,
	clientID int64,
	limit int,
	offset int,
) (models.INVBookOut, error) {
	if clientID < 1 {
		return models.INVBookOut{}, fmt.Errorf("invalid clientID: %d", clientID)
	}
	if limit < 1 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// --------------------------------------------------
	// 1. Fetches paginated parent invoices by base number - 1, 2, 3...
	// --------------------------------------------------
	invoiceRows, err := a.DB.QueryContext(ctx, `
		SELECT
			i.id,
			i.base_number,
			i.status
		FROM invoices i
		WHERE i.client_id = ?
		ORDER BY i.base_number DESC
		LIMIT ? OFFSET ?;
	`, clientID, limit, offset)
	if err != nil {
		return models.INVBookOut{}, fmt.Errorf("query paged invoices: %w", err)
	}
	defer invoiceRows.Close()

	items := make([]models.INVBookInvoice, 0, limit)
	invoiceIDs := make([]int64, 0, limit)

	// map invoice_id -> index in items slice
	itemIndexByInvoiceID := make(map[int64]int, limit)

	for invoiceRows.Next() {
		var item models.INVBookInvoice
		if err := invoiceRows.Scan(
			&item.ID,
			&item.BaseNo,
			&item.Status,
		); err != nil {
			return models.INVBookOut{}, fmt.Errorf("scan paged invoice row: %w", err)
		}

		item.Revisions = make([]models.INVBookRevision, 0, 2)

		itemIndexByInvoiceID[item.ID] = len(items)
		invoiceIDs = append(invoiceIDs, item.ID)
		items = append(items, item)
	}

	if err := invoiceRows.Err(); err != nil {
		return models.INVBookOut{}, fmt.Errorf("iterate paged invoice rows: %w", err)
	}

	// No invoices on this page
	if len(items) == 0 {
		return models.INVBookOut{Items: items}, nil
	}

	// --------------------------------------------------
	// 2. Fetch all revisions for those invoices
	// --------------------------------------------------
	placeholders := make([]string, len(invoiceIDs))
	args := make([]any, 0, len(invoiceIDs))
	for i, id := range invoiceIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	revisionsSQL := fmt.Sprintf(`
		SELECT
			r.id,
			r.invoice_id,
			r.revision_no,
			r.issue_date,
			r.due_by_date,
			r.updated_at
		FROM invoice_revisions r
		WHERE r.invoice_id IN (%s)
		ORDER BY r.invoice_id DESC, r.revision_no ASC;
	`, strings.Join(placeholders, ","))

	revisionRows, err := a.DB.QueryContext(ctx, revisionsSQL, args...)
	if err != nil {
		return models.INVBookOut{}, fmt.Errorf("query invoice revisions for page: %w", err)
	}
	defer revisionRows.Close()

	for revisionRows.Next() {
		var (
			revision  models.INVBookRevision
			invoiceID int64
		)

		if err := revisionRows.Scan(
			&revision.ID,
			&invoiceID,
			&revision.RevisionNo,
			&revision.IssueDate,
			&revision.DueByDate,
			&revision.UpdatedAt,
		); err != nil {
			return models.INVBookOut{}, fmt.Errorf("scan invoice revision row: %w", err)
		}

		idx, ok := itemIndexByInvoiceID[invoiceID]
		if !ok {
			return models.INVBookOut{}, fmt.Errorf("revision references unexpected invoice_id: %d", invoiceID)
		}

		items[idx].Revisions = append(items[idx].Revisions, revision)
	}

	if err := revisionRows.Err(); err != nil {
		return models.INVBookOut{}, fmt.Errorf("iterate invoice revision rows: %w", err)
	}

	return models.INVBookOut{
		Items: items,
	}, nil
}
