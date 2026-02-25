package productsTx_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

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

func TestProducts_CreateUpdateDelete(t *testing.T) {
	ctx := context.Background()
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

func ptrI64(v int64) *int64 { return &v }
