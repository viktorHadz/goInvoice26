package team

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/authTx"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

type inviteInput struct {
	Email string `json:"email"`
}

func List(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID := accountscope.AccountID(r.Context())

		members, err := authTx.ListTeamMembers(r.Context(), a.DB, accountID)
		if err != nil {
			slog.ErrorContext(r.Context(), "list team members failed", "err", err, "account_id", accountID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load team members")
			return
		}

		invites, err := authTx.ListPendingInvites(r.Context(), a.DB, accountID)
		if err != nil {
			slog.ErrorContext(r.Context(), "list team invites failed", "err", err, "account_id", accountID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load team invites")
			return
		}

		res.JSON(w, http.StatusOK, models.TeamSummary{
			Members: members,
			Invites: invites,
		})
	}
}

func CreateInvite(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID := accountscope.AccountID(r.Context())
		actingUserID := userscope.UserID(r.Context())

		var input inviteInput
		if !res.DecodeJSON(w, r, &input) {
			return
		}

		input.Email = strings.TrimSpace(input.Email)
		if input.Email == "" {
			res.Validation(w, res.Required("email"))
			return
		}

		invite, err := authTx.CreateInvite(r.Context(), a.DB, accountID, actingUserID, input.Email)
		if err != nil {
			switch {
			case errors.Is(err, authTx.ErrInvalidEmail):
				res.Validation(w, res.Invalid("email", "must be a valid email address"))
				return
			case errors.Is(err, authTx.ErrInviteAlreadyExists):
				res.Error(w, http.StatusConflict, "TEAM_INVITE_EXISTS", "That email already has a pending invite.")
				return
			case errors.Is(err, authTx.ErrMemberAlreadyExists):
				res.Error(w, http.StatusConflict, "TEAM_MEMBER_EXISTS", "That teammate already has access.")
				return
			}

			slog.ErrorContext(r.Context(), "create team invite failed", "err", err, "account_id", accountID, "email", input.Email)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to create invite")
			return
		}

		res.JSON(w, http.StatusCreated, invite)
	}
}

func DeleteInvite(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID := accountscope.AccountID(r.Context())
		inviteID, ok := params.ValidateParam(w, r, "inviteID")
		if !ok {
			return
		}

		deleted, err := authTx.DeleteInvite(r.Context(), a.DB, accountID, inviteID)
		if err != nil {
			slog.ErrorContext(r.Context(), "delete team invite failed", "err", err, "account_id", accountID, "invite_id", inviteID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to revoke invite")
			return
		}
		if !deleted {
			res.NotFound(w, "Invite not found")
			return
		}

		res.NoContent(w)
	}
}

func DeleteMember(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID := accountscope.AccountID(r.Context())
		actingUserID := userscope.UserID(r.Context())
		memberUserID, ok := params.ValidateParam(w, r, "memberID")
		if !ok {
			return
		}

		removed, err := authTx.RemoveMember(r.Context(), a.DB, accountID, actingUserID, memberUserID)
		if err != nil {
			switch {
			case errors.Is(err, authTx.ErrCannotRemoveSelf):
				res.Error(w, http.StatusConflict, "TEAM_CANNOT_REMOVE_SELF", "Use sign out instead of removing your own account.")
				return
			case errors.Is(err, authTx.ErrCannotRemoveOwner):
				res.Error(w, http.StatusConflict, "TEAM_CANNOT_REMOVE_OWNER", "The owner account cannot be removed here.")
				return
			}

			slog.ErrorContext(r.Context(), "remove team member failed", "err", err, "account_id", accountID, "member_id", memberUserID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to remove teammate")
			return
		}
		if !removed {
			res.NotFound(w, "Teammate not found")
			return
		}

		res.NoContent(w)
	}
}
