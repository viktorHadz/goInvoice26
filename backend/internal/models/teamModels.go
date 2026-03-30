package models

type TeamMember struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarUrl"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

type TeamInvite struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

type TeamSummary struct {
	Members []TeamMember `json:"members"`
	Invites []TeamInvite `json:"invites"`
}
