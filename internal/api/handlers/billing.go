package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
)

// BillingHandler handles billing and subscription related API endpoints
type BillingHandler struct {
	logger    *logger.Logger
	validator *validator.Validator
}

// NewBillingHandler creates a new BillingHandler
func NewBillingHandler(log *logger.Logger, val *validator.Validator) *BillingHandler {
	return &BillingHandler{
		logger:    log,
		validator: val,
	}
}

// ListPlans returns available subscription plans
// @Summary List subscription plans
// @Description Get a list of available subscription plans
// @Tags Billing
// @Produce json
// @Success 200 {array} dto.PlanDTO "List of plans"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /billing/plans [get]
func (h *BillingHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	// Mock data
	plans := []dto.PlanDTO{
		{
			ID:          "free",
			Name:        "Free",
			Description: "For individuals and small projects",
			Price:       0,
			Currency:    "USD",
			Interval:    "month",
			Features: []string{
				"Up to 3 cloud accounts",
				"Basic security scanning",
				"Daily drift detection",
				"Community support",
			},
			IsPopular: false,
			IsCurrent: true, // Default for new users
		},
		{
			ID:          "pro",
			Name:        "Pro",
			Description: "For growing teams and startups",
			Price:       29,
			Currency:    "USD",
			Interval:    "month",
			Features: []string{
				"Unlimited cloud accounts",
				"Advanced security scanning",
				"Real-time drift detection",
				"AI-powered recommendations",
				"Priority support",
				"Webhook integrations",
			},
			IsPopular: true,
			IsCurrent: false,
		},
		{
			ID:          "enterprise",
			Name:        "Enterprise",
			Description: "For large organizations with strict compliance needs",
			Price:       99,
			Currency:    "USD",
			Interval:    "month",
			Features: []string{
				"Everything in Pro",
				"SSO / SAML",
				"Custom compliance frameworks",
				"Dedicated success manager",
				"SLA guarantees",
				"On-premise deployment option",
			},
			IsPopular: false,
			IsCurrent: false,
		},
	}

	utils.WriteSuccess(w, http.StatusOK, plans)
}

// GetBillingInfo returns user billing information
// @Summary Get billing info
// @Description Get current user billing information including plan and invoices
// @Tags Billing
// @Produce json
// @Success 200 {object} dto.BillingInfoDTO "Billing information"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /billing/info [get]
func (h *BillingHandler) GetBillingInfo(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	h.logger.Infof("Getting billing info for user %d", userID)

	// Mock data based on the "free" plan as default
	nextBill := time.Now().Add(30 * 24 * time.Hour)

	info := dto.BillingInfoDTO{
		Plan: dto.PlanDTO{
			ID:          "free",
			Name:        "Free",
			Description: "For individuals and small projects",
			Price:       0,
			Currency:    "USD",
			Interval:    "month",
		},
		Status:        "active",
		NextBillingAt: &nextBill,
		Invoices:      []dto.InvoiceDTO{}, // Empty for free plan
	}

	utils.WriteSuccess(w, http.StatusOK, info)
}

// UpdatePlan updates the user's subscription plan
// @Summary Update subscription plan
// @Description Upgrade or downgrade user subscription plan
// @Tags Billing
// @Accept json
// @Produce json
// @Param request body dto.UpdatePlanRequest true "Update plan request"
// @Success 200 {object} utils.SuccessResponse "Plan updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /billing/subscription [post]
func (h *BillingHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	if errs := h.validator.Validate(req); len(errs) > 0 {
		utils.WriteError(w, errors.ValidationError("Validation failed", errs))
		return
	}

	userID, _ := middleware.GetUserID(r)
	h.logger.Infof("Updating plan for user %d to %s", userID, req.PlanID)

	// In a real implementation, this would interact with Stripe/payment provider
	utils.WriteSuccessWithMessage(w, http.StatusOK, "Subscription updated successfully", nil)
}

// CreateCheckoutSession creates a checkout session for upgrading
// @Summary Create checkout session
// @Description Create a Stripe checkout session for plan upgrade
// @Tags Billing
// @Accept json
// @Produce json
// @Param request body dto.UpdatePlanRequest true "Plan to upgrade to"
// @Success 200 {object} map[string]string "Checkout URL"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /billing/checkout [post]
func (h *BillingHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	// Mock response
	resp := map[string]string{
		"url": "https://checkout.stripe.com/test-session",
	}
	utils.WriteSuccess(w, http.StatusOK, resp)
}
