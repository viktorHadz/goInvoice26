package models

type AuthUser struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarUrl"`
	Role      string `json:"role"`
}

type AuthAccount struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type AuthStatus struct {
	Authenticated bool         `json:"authenticated"`
	NeedsSetup    bool         `json:"needsSetup"`
	GoogleEnabled bool         `json:"googleEnabled"`
	User          *AuthUser    `json:"user,omitempty"`
	Account       *AuthAccount `json:"account,omitempty"`
	Billing       *AuthBilling `json:"billing,omitempty"`
}
