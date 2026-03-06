package invoiceTx

// Check DB this is a new invoice
// Checks that a client exists in clients table and a current revision id exists in invoices
// If it doesnt exist proceed to creating invoice
// func CheckRevisionID(ctx context.Context, a *app.App, clientID int64) (int64, error) {
// 	// Check client exists first
// 	if _, err := clientsTx.Exists(ctx, a, clientID); err != nil {
// 		return 0, err
// 	}
// 	// query DB for Exisitng revision ID and if no revision id set it to null
// 	revID, err := a.DB.QueryContext(
// 		ctx,
// 		`SELECT current_revision_id FROM invoices WHERE client_id=?;`,
// 		clientID)
// 	if err != nil {
// 		return 0, err
// 	}

// 	return 1, err
// }

var (
	clientStmt string = `INSERT 
INTO invoice_revisions(
	client_name,
	client_company_name,
	client_address,
	client_email)
VALUES(?, ?, ?, ?)`
)

// func Create(ctx *context.Context, a *app.App, i *models.InvoiceCreateOut) (bool, error) {

// 	_, err := a.DB.ExecContext(*ctx, clientStmt, &i.ClientName, &i.ClientCompanyName, &i.ClientAddress, &i.ClientEmail)
// 	if err != nil {
// 		return false, err
// 	}
// 	return true, nil
// }
