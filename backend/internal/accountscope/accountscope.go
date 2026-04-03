package accountscope

import (
	"context"
	"errors"
)

type key struct{}

const DefaultAccountID int64 = 1

var ErrMissing = errors.New("account scope missing from context")

func WithAccountID(ctx context.Context, accountID int64) context.Context {
	if accountID <= 0 {
		return ctx
	}

	return context.WithValue(ctx, key{}, accountID)
}

func AccountID(ctx context.Context) int64 {
	accountID, _ := FromContext(ctx)
	return accountID
}

func FromContext(ctx context.Context) (int64, bool) {
	accountID, ok := ctx.Value(key{}).(int64)
	return accountID, ok && accountID > 0
}

func Require(ctx context.Context) (int64, error) {
	accountID, ok := FromContext(ctx)
	if !ok {
		return 0, ErrMissing
	}
	return accountID, nil
}
