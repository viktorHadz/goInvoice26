package editorTx

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

type InvoiceBookPageFilters struct {
	ClientID      int64
	SortBy        string
	SortDirection string
	PaymentState  string
}

func normalizeInvoiceBookPageFilters(filters InvoiceBookPageFilters) InvoiceBookPageFilters {
	out := InvoiceBookPageFilters{
		ClientID:      filters.ClientID,
		SortBy:        "date",
		SortDirection: "desc",
		PaymentState:  "all",
	}

	if filters.SortBy == "balance" {
		out.SortBy = "balance"
	}

	if filters.SortDirection == "asc" {
		out.SortDirection = "asc"
	}

	switch filters.PaymentState {
	case "paid", "unpaid":
		out.PaymentState = filters.PaymentState
	}

	return out
}

func invoiceBookWhereClause(filters InvoiceBookPageFilters) string {
	switch filters.PaymentState {
	case "paid":
		return `
		WHERE status <> 'void'
			AND balance_due_minor <= 0
		`
	case "unpaid":
		return `
		WHERE status <> 'void'
			AND balance_due_minor > 0
		`
	default:
		return ""
	}
}

func invoiceBookOrderClause(filters InvoiceBookPageFilters) string {
	direction := "DESC"
	if filters.SortDirection == "asc" {
		direction = "ASC"
	}

	if filters.SortBy == "balance" {
		return fmt.Sprintf(
			"ORDER BY balance_due_minor %s, issue_date DESC, base_number DESC",
			direction,
		)
	}

	return fmt.Sprintf("ORDER BY issue_date %s, base_number %s", direction, direction)
}

func invoiceBookBaseCTE(accountID int64, filters InvoiceBookPageFilters) (string, []any) {
	clientWhere := "WHERE i.account_id = ?"
	args := make([]any, 0, 2)
	args = append(args, accountID)
	if filters.ClientID > 0 {
		clientWhere += " AND i.client_id = ?"
		args = append(args, filters.ClientID)
	}

	baseCTE := fmt.Sprintf(`
		WITH paid_totals AS (
			SELECT
				p.invoice_id,
				COALESCE(SUM(p.amount_minor), 0) AS paid_minor
			FROM payments p
			WHERE p.payment_type = 'payment'
			GROUP BY p.invoice_id
		),
		invoice_page_rows AS (
			SELECT
				i.id,
				i.client_id,
				cur.client_name,
				cur.client_company_name,
				i.base_number,
				i.status,
				cur.revision_no,
				cur.issue_date,
				cur.due_by_date,
				cur.total_minor,
				cur.deposit_minor,
				COALESCE(pt.paid_minor, 0) AS paid_minor,
				CASE
					WHEN cur.total_minor - COALESCE(pt.paid_minor, 0) > 0
						THEN cur.total_minor - COALESCE(pt.paid_minor, 0)
					ELSE 0
				END AS balance_due_minor
			FROM invoices i
			JOIN invoice_revisions cur
				ON cur.id = i.current_revision_id
			LEFT JOIN paid_totals pt
				ON pt.invoice_id = i.id
			%s
		)
	`, clientWhere)

	return baseCTE, args
}

func QueryInvoiceBookPage(
	a *app.App,
	ctx context.Context,
	clientID int64,
	limit int,
	offset int,
	filters InvoiceBookPageFilters,
) (models.INVBookOut, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	filters = normalizeInvoiceBookPageFilters(filters)
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return models.INVBookOut{}, err
	}
	filters.ClientID = clientID
	baseCTE, baseArgs := invoiceBookBaseCTE(accountID, filters)
	whereClause := invoiceBookWhereClause(filters)

	var total int
	countSQL := baseCTE + fmt.Sprintf(`
		SELECT COUNT(*)
		FROM invoice_page_rows
		%s;
	`, whereClause)
	if err := a.DB.QueryRowContext(ctx, countSQL, baseArgs...).Scan(&total); err != nil {
		return models.INVBookOut{}, fmt.Errorf("count invoices: %w", err)
	}

	pageSQL := baseCTE + fmt.Sprintf(`
		SELECT
			id,
			client_id,
			client_name,
			client_company_name,
			base_number,
			status,
			revision_no,
			issue_date,
			due_by_date,
			total_minor,
			deposit_minor,
			paid_minor,
			balance_due_minor
		FROM invoice_page_rows
		%s
		%s
		LIMIT ? OFFSET ?;
	`, whereClause, invoiceBookOrderClause(filters))

	pageArgs := append(append([]any{}, baseArgs...), limit, offset)
	invoiceRows, err := a.DB.QueryContext(ctx, pageSQL, pageArgs...)
	if err != nil {
		return models.INVBookOut{}, fmt.Errorf("query paged invoices: %w", err)
	}
	defer invoiceRows.Close()

	items := make([]models.INVBookInvoice, 0, limit)
	invoiceIDs := make([]int64, 0, limit)
	itemIndexByInvoiceID := make(map[int64]int, limit)

	for invoiceRows.Next() {
		var item models.INVBookInvoice
		if err := invoiceRows.Scan(
			&item.ID,
			&item.ClientID,
			&item.ClientName,
			&item.ClientCompanyName,
			&item.BaseNo,
			&item.Status,
			&item.LatestRevisionNo,
			&item.IssueDate,
			&item.DueByDate,
			&item.TotalMinor,
			&item.DepositMinor,
			&item.PaidMinor,
			&item.BalanceDueMinor,
		); err != nil {
			return models.INVBookOut{}, fmt.Errorf("scan paged invoice row: %w", err)
		}

		item.Revisions = make([]models.INVBookRevision, 0, 2)
		item.History = make([]models.INVBookHistoryItem, 0, 4)

		itemIndexByInvoiceID[item.ID] = len(items)
		invoiceIDs = append(invoiceIDs, item.ID)
		items = append(items, item)
	}

	if err := invoiceRows.Err(); err != nil {
		return models.INVBookOut{}, fmt.Errorf("iterate paged invoice rows: %w", err)
	}

	if len(items) == 0 {
		return models.INVBookOut{
			Items:   []models.INVBookInvoice{},
			Limit:   limit,
			Offset:  offset,
			Count:   0,
			Total:   total,
			HasMore: false,
		}, nil
	}

	// --------------------------------------------------
	// 2. Fetch all revisions for invoices on this page
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

	historyRows, err := invoiceTx.QueryInvoiceHistoryForInvoices(ctx, a.DB, invoiceIDs)
	if err != nil {
		return models.INVBookOut{}, fmt.Errorf("query invoice history for page: %w", err)
	}
	for _, history := range historyRows {
		idx, ok := itemIndexByInvoiceID[history.InvoiceID]
		if !ok {
			return models.INVBookOut{}, fmt.Errorf("history references unexpected invoice_id: %d", history.InvoiceID)
		}

		items[idx].History = append(items[idx].History, toInvoiceBookHistoryItem(history))
	}

	count := len(items)

	return models.INVBookOut{
		Items:   items,
		Limit:   limit,
		Offset:  offset,
		Count:   count,
		Total:   total,
		HasMore: offset+count < total,
	}, nil
}

func toInvoiceBookHistoryItem(row invoiceTx.InvoiceHistoryRow) models.INVBookHistoryItem {
	item := models.INVBookHistoryItem{
		ID:        row.ID,
		Type:      row.Type,
		CreatedAt: row.CreatedAt,
		Label:     optionalStringPtr(row.Label),
	}

	if row.RevisionNo.Valid {
		revisionNo := int(row.RevisionNo.Int64)
		item.RevisionNo = &revisionNo
	}
	if row.ReceiptNo.Valid {
		receiptNo := int(row.ReceiptNo.Int64)
		item.ReceiptNo = &receiptNo
	}
	if row.IssueDate.Valid {
		issueDate := row.IssueDate.String
		item.IssueDate = &issueDate
	}
	if row.DueByDate.Valid {
		dueByDate := row.DueByDate.String
		item.DueByDate = &dueByDate
	}
	if row.PaymentDate.Valid {
		paymentDate := row.PaymentDate.String
		item.PaymentDate = &paymentDate
	}
	if row.AmountMinor.Valid {
		amountMinor := row.AmountMinor.Int64
		item.AmountMinor = &amountMinor
	}

	return item
}

func optionalStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	out := value.String
	return &out
}
