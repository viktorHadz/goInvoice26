package accountscope

import "context"

type key struct{}

const DefaultAccountID int64 = 1

func WithAccountID(ctx context.Context, accountID int64) context.Context {
	if accountID <= 0 {
		accountID = DefaultAccountID
	}

	return context.WithValue(ctx, key{}, accountID)
}

func AccountID(ctx context.Context) int64 {
	if accountID, ok := ctx.Value(key{}).(int64); ok && accountID > 0 {
		return accountID
	}

	return DefaultAccountID
}
