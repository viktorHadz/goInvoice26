package userscope

import "context"

type key struct{}

type Principal struct {
	UserID    int64
	AccountID int64
	Email     string
	Role      string
	Name      string
}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, key{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(key{}).(Principal)
	if !ok {
		return Principal{}, false
	}

	return principal, principal.UserID > 0
}

func UserID(ctx context.Context) int64 {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return 0
	}

	return principal.UserID
}

func Role(ctx context.Context) string {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return ""
	}

	return principal.Role
}
