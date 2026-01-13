package dto

import "time"

// PlanDTO represents a subscription plan
type PlanDTO struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Currency    string   `json:"currency"`
	Interval    string   `json:"interval"` // month, year
	Features    []string `json:"features"`
	IsPopular   bool     `json:"isPopular"`
	IsCurrent   bool     `json:"isCurrent"`
}

// BillingInfoDTO represents user billing information
type BillingInfoDTO struct {
	Plan          PlanDTO      `json:"plan"`
	Status        string       `json:"status"` // active, past_due, canceled
	NextBillingAt *time.Time   `json:"nextBillingAt,omitempty"`
	PaymentMethod *string      `json:"paymentMethod,omitempty"` // e.g., "Visa ending in 4242"
	Invoices      []InvoiceDTO `json:"invoices"`
}

// InvoiceDTO represents a billing invoice
type InvoiceDTO struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"` // paid, open, void
	CreatedAt time.Time `json:"createdAt"`
	PdfURL    string    `json:"pdfUrl"`
}

// UpdatePlanRequest represents a request to change subscription plan
type UpdatePlanRequest struct {
	PlanID string `json:"planId" validate:"required"`
}
