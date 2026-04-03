package workspace

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"

	billingsvc "github.com/viktorHadz/goInvoice26/internal/service/billing"
	"github.com/viktorHadz/goInvoice26/internal/service/storage"
	"github.com/viktorHadz/goInvoice26/internal/transaction/authTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/billingTx"
)

var ErrDeleteBlockedByBilling = errors.New("workspace deletion blocked until billing is canceled")

type subscriptionCanceler interface {
	CancelSubscriptionImmediately(ctx context.Context, accountID int64) error
}

type Service struct {
	db      *sql.DB
	billing subscriptionCanceler
	store   *storage.LocalStore
}

func NewService(db *sql.DB, billing subscriptionCanceler, store *storage.LocalStore) *Service {
	return &Service{
		db:      db,
		billing: billing,
		store:   store,
	}
}

func (s *Service) DeleteAccount(ctx context.Context, accountID int64) error {
	record, err := billingTx.GetAccountBilling(ctx, s.db, accountID)
	if err != nil {
		return err
	}

	if strings.TrimSpace(record.StripeSubscriptionID) != "" {
		if s.billing == nil {
			return ErrDeleteBlockedByBilling
		}
		if err := s.billing.CancelSubscriptionImmediately(ctx, accountID); err != nil {
			if errors.Is(err, billingsvc.ErrNotConfigured) {
				return ErrDeleteBlockedByBilling
			}
			return err
		}
	}

	var (
		staged     storage.StagedAccountDirRemoval
		hasStaged  bool
		shouldUndo bool
	)
	if s.store != nil {
		staged, hasStaged, err = s.store.StageAccountDirRemoval(accountID)
		if err != nil {
			return err
		}
		shouldUndo = hasStaged
		defer func() {
			if !shouldUndo {
				return
			}
			if rollbackErr := staged.Rollback(); rollbackErr != nil {
				slog.ErrorContext(ctx, "workspace delete failed to restore staged uploads", "accountID", accountID, "err", rollbackErr)
			}
		}()
	}

	if err := authTx.DeleteAccount(ctx, s.db, accountID); err != nil {
		return err
	}

	shouldUndo = false
	if hasStaged {
		if err := staged.Commit(); err != nil {
			slog.ErrorContext(ctx, "workspace deleted but staged uploads cleanup failed", "accountID", accountID, "err", err)
		}
	}

	return nil
}
