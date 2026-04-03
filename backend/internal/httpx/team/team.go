package team

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/billingplan"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	workspaceSvc "github.com/viktorHadz/goInvoice26/internal/service/workspace"
	"github.com/viktorHadz/goInvoice26/internal/transaction/authTx"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

type inviteInput struct {
	Email string `json:"email"`
}

func List(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, err := accountscope.Require(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "list team members missing account scope", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load team members")
			return
		}

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
		accountID, err := accountscope.Require(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "create team invite missing account scope", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to create invite")
			return
		}
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

		principal, ok := userscope.PrincipalFromContext(r.Context())
		if !ok {
			slog.ErrorContext(r.Context(), "create team invite missing principal")
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to create invite")
			return
		}
		if !billingplan.SupportsTeam(principal.BillingPlan) {
			res.Error(
				w,
				http.StatusPaymentRequired,
				"TEAM_PLAN_REQUIRED",
				"Inviting teammates requires an active team plan.",
			)
			return
		}

		invite, err := authTx.CreateInvite(
			r.Context(),
			a.DB,
			accountID,
			actingUserID,
			input.Email,
			billingplan.TeamSeatLimit,
		)
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
			case errors.Is(err, authTx.ErrTeamSeatLimitReached):
				res.Error(
					w,
					http.StatusConflict,
					"TEAM_SEAT_LIMIT_REACHED",
					"The team plan supports up to 5 people including pending invites.",
				)
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
		accountID, err := accountscope.Require(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "delete team invite missing account scope", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to revoke invite")
			return
		}
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
		accountID, err := accountscope.Require(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "remove team member missing account scope", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to remove teammate")
			return
		}
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

func DeleteWorkspace(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := userscope.PrincipalFromContext(r.Context())
		if !ok || principal.Role != authTx.UserRoleOwner {
			slog.WarnContext(r.Context(), "workspace deletion rejected for non-owner user")
			res.Error(w, http.StatusForbidden, "FORBIDDEN", "Only the workspace admin can delete the workspace")
			return
		}

		if err := a.Workspaces.DeleteAccount(r.Context(), principal.AccountID); err != nil {
			switch {
			case errors.Is(err, workspaceSvc.ErrDeleteBlockedByBilling):
				res.Error(
					w,
					http.StatusConflict,
					"WORKSPACE_DELETE_BILLING_BLOCKED",
					"Cancel the Stripe subscription before deleting this workspace.",
				)
				return
			case errors.Is(err, authTx.ErrAccountNotFound):
				res.NotFound(w, "Workspace not found")
				return
			default:
				slog.ErrorContext(r.Context(), "delete workspace failed", "err", err, "account_id", principal.AccountID)
				res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to delete workspace")
				return
			}
		}

		http.SetCookie(w, a.Auth.ClearSessionCookie())
		res.NoContent(w)
	}
}
