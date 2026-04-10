package productsTx_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/productsTx"
)

func newTestApp(t *testing.T) (*app.App, func()) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")

	d, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := db.Migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	a := &app.App{DB: d}
	cleanup := func() {
		_ = d.Close()
		_ = os.Remove(dbPath)
	}
	return a, cleanup
}

func insertClient(t *testing.T, a *app.App) int64 {
	t.Helper()

	res, err := a.DB.Exec(`INSERT INTO clients (name) VALUES (?)`, "Test Client")
	if err != nil {
		t.Fatalf("insert client: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("client lastInsertId: %v", err)
	}
	return id
}

func insertClientForAccount(t *testing.T, a *app.App, accountID int64, name string) int64 {
	t.Helper()

	res, err := a.DB.Exec(`INSERT INTO clients (account_id, name) VALUES (?, ?)`, accountID, name)
	if err != nil {
		t.Fatalf("insert client for account %d: %v", accountID, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("client lastInsertId: %v", err)
	}
	return id
}

func TestProducts_CreateUpdateDelete(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)

	// ---- CREATE (sample hourly) ----
	create := models.ProductCreate{
		ProductType:     "sample",
		PricingMode:     "hourly",
		ProductName:     "Pattern Adjustment",
		HourlyRateMinor: ptrI64(3000),
		MinutesWorked:   ptrI64(90),
		ClientID:        clientID,
	}

	created, err := productsTx.InsertTx(a, ctx, create)
	if err != nil {
		t.Fatalf("CreateTx: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created.ID != 0")
	}
	if created.ClientID != clientID {
		t.Fatalf("clientID: got %d want %d", created.ClientID, clientID)
	}
	if created.HourlyRateMinor == nil || *created.HourlyRateMinor != 3000 {
		t.Fatalf("hourlyRateMinor: got %#v want 3000", created.HourlyRateMinor)
	}

	// ---- UPDATE (switch to flat sample) ----
	update := models.ProductCreate{
		ProductType:    "sample",
		PricingMode:    "flat",
		ProductName:    "Updated Flat Sample",
		FlatPriceMinor: ptrI64(1250),
		ClientID:       clientID,
	}

	updated, err := productsTx.UpdateTx(a, ctx, created.ID, update)
	if err != nil {
		t.Fatalf("UpdateTx: %v", err)
	}
	if updated.PricingMode != "flat" {
		t.Fatalf("pricingMode: got %q want %q", updated.PricingMode, "flat")
	}
	if updated.FlatPriceMinor == nil || *updated.FlatPriceMinor != 1250 {
		t.Fatalf("flatPriceMinor: got %#v want 1250", updated.FlatPriceMinor)
	}
	if updated.HourlyRateMinor != nil || updated.MinutesWorked != nil {
		t.Fatalf("expected hourly fields cleared, got hourly=%v minutes=%v", updated.HourlyRateMinor, updated.MinutesWorked)
	}
	if updated.UpdatedAt == nil || *updated.UpdatedAt == "" {
		t.Fatalf("expected UpdatedAt set")
	}

	// ---- DELETE ----
	if err := productsTx.DeleteTx(a, ctx, created.ID, clientID); err != nil {
		t.Fatalf("DeleteTx: %v", err)
	}

	// verify deleted
	var count int
	if err := a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE id = ?`, created.ID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected deleted product, count=%d", count)
	}
}

func TestProducts_AreScopedPerAccount(t *testing.T) {
	defaultCtx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	secondCtx := accountscope.WithAccountID(context.Background(), 2)
	a, cleanup := newTestApp(t)
	defer cleanup()

	if _, err := a.DB.Exec(`INSERT INTO accounts (id, name) VALUES (2, 'Second account')`); err != nil {
		t.Fatalf("insert second account: %v", err)
	}

	defaultClientID := insertClientForAccount(t, a, accountscope.DefaultAccountID, "Default Client")
	secondClientID := insertClientForAccount(t, a, 2, "Second Client")

	defaultProduct, err := productsTx.InsertTx(a, defaultCtx, models.ProductCreate{
		ProductType:    "style",
		PricingMode:    "flat",
		ProductName:    "Default Style",
		FlatPriceMinor: ptrI64(900),
		ClientID:       defaultClientID,
	})
	if err != nil {
		t.Fatalf("insert default product: %v", err)
	}

	secondProduct, err := productsTx.InsertTx(a, secondCtx, models.ProductCreate{
		ProductType:    "style",
		PricingMode:    "flat",
		ProductName:    "Second Style",
		FlatPriceMinor: ptrI64(1200),
		ClientID:       secondClientID,
	})
	if err != nil {
		t.Fatalf("insert second product: %v", err)
	}

	defaultProducts, err := productsTx.ListAll(a, defaultCtx, defaultClientID)
	if err != nil {
		t.Fatalf("list default products: %v", err)
	}
	if len(defaultProducts) != 1 || defaultProducts[0].ID != defaultProduct.ID {
		t.Fatalf("default products = %+v, want only %d", defaultProducts, defaultProduct.ID)
	}

	secondProducts, err := productsTx.ListAll(a, secondCtx, secondClientID)
	if err != nil {
		t.Fatalf("list second products: %v", err)
	}
	if len(secondProducts) != 1 || secondProducts[0].ID != secondProduct.ID {
		t.Fatalf("second products = %+v, want only %d", secondProducts, secondProduct.ID)
	}

	_, err = productsTx.UpdateTx(a, defaultCtx, secondProduct.ID, models.ProductCreate{
		ProductType:    "style",
		PricingMode:    "flat",
		ProductName:    "Hijacked",
		FlatPriceMinor: ptrI64(1300),
		ClientID:       secondClientID,
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-account update error = %v, want sql.ErrNoRows", err)
	}

	err = productsTx.DeleteTx(a, defaultCtx, secondProduct.ID, secondClientID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-account delete error = %v, want sql.ErrNoRows", err)
	}
}

func TestProducts_RequireAccountScope(t *testing.T) {
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)

	_, err := productsTx.InsertTx(a, context.Background(), models.ProductCreate{
		ProductType:    "style",
		PricingMode:    "flat",
		ProductName:    "Missing Scope",
		FlatPriceMinor: ptrI64(500),
		ClientID:       clientID,
	})
	if !errors.Is(err, accountscope.ErrMissing) {
		t.Fatalf("InsertTx missing scope error = %v, want ErrMissing", err)
	}
}

func TestProducts_BulkInsertTx_InsertsAllRows(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)
	rows := []models.ProductCreate{
		{
			ProductType:    "style",
			PricingMode:    "flat",
			ProductName:    "Hemline",
			FlatPriceMinor: ptrI64(1250),
			ClientID:       clientID,
		},
		{
			ProductType:     "sample",
			PricingMode:     "hourly",
			ProductName:     "Pattern Adjustment",
			HourlyRateMinor: ptrI64(3000),
			MinutesWorked:   ptrI64(90),
			ClientID:        clientID,
		},
	}

	inserted, err := productsTx.BulkInsertTx(a, ctx, rows)
	if err != nil {
		t.Fatalf("BulkInsertTx: %v", err)
	}
	if inserted != len(rows) {
		t.Fatalf("inserted = %d, want %d", inserted, len(rows))
	}

	listed, err := productsTx.ListAll(a, ctx, clientID)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("listed len = %d, want 2", len(listed))
	}
}

func TestProducts_BulkInsertTx_RollsBackOnError(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)
	rows := []models.ProductCreate{
		{
			ProductType:    "style",
			PricingMode:    "flat",
			ProductName:    "Hemline",
			FlatPriceMinor: ptrI64(1250),
			ClientID:       clientID,
		},
		{
			ProductType:    "style",
			PricingMode:    "flat",
			ProductName:    "Broken Row",
			FlatPriceMinor: nil,
			ClientID:       clientID,
		},
	}

	_, err := productsTx.BulkInsertTx(a, ctx, rows)
	if err == nil {
		t.Fatal("expected bulk insert error")
	}

	listed, listErr := productsTx.ListAll(a, ctx, clientID)
	if listErr != nil {
		t.Fatalf("ListAll: %v", listErr)
	}
	if len(listed) != 0 {
		t.Fatalf("listed len = %d, want 0 after rollback", len(listed))
	}
}

func TestProducts_ListAll_ReturnsEmptySliceForClientWithoutProducts(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)

	listed, err := productsTx.ListAll(a, ctx, clientID)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if listed == nil {
		t.Fatal("ListAll returned nil slice, want empty slice")
	}
	if len(listed) != 0 {
		t.Fatalf("listed len = %d, want 0", len(listed))
	}
}

func ptrI64(v int64) *int64 { return &v }
