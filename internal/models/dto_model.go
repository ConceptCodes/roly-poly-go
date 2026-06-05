package models

import "github.com/google/uuid"

// HealthCheckResponseDto represents the health status of a single service dependency.
type HealthCheckResponseDto struct {
	Status  bool   `json:"status"`
	Service string `json:"service"`
}

// OnboardUserRequestDto is the request body for creating a new user.
type OnboardUserRequestDto struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100" example:"Ada"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100" example:"Lovelace"`
}

// OnboardUserResponseDto is the response returned after onboarding a user.
type OnboardUserResponseDto struct {
	ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ApiKey    string    `json:"api_key" example:"rp_a1b2c3d4e5f6..." `
	FirstName string    `json:"first_name" example:"Ada"`
	LastName  string    `json:"last_name" example:"Lovelace"`
}

// CreatePollRequestDto is the request body for creating a poll.
type CreatePollRequestDto struct {
	Title       string   `json:"title" validate:"required,min=2,max=100" example:"Lunch"`
	Description string   `json:"description" validate:"max=100" example:"Where should we eat?"`
	Public      bool     `json:"public" example:"true"`
	Options     []string `json:"options" validate:"required,dive,required,min=2,max=100" example:"[\"Tacos\",\"Pizza\",\"Salad\"]"`
}

// CastVoteRequestDto is the request body for casting votes.
type CastVoteRequestDto struct {
	OptionIDs []uuid.UUID `json:"option_ids" validate:"required,min=1"`
}

// OptionSummaryDto contains the vote count and percentage for a single option.
type OptionSummaryDto struct {
	OptionID   uuid.UUID `json:"option_id"`
	Label      string    `json:"label" example:"Tacos"`
	Votes      uint      `json:"votes" example:"42"`
	Percentage float64   `json:"percentage" example:"66.67"`
}

// PollReportDto is the full vote report for a poll.
type PollReportDto struct {
	PollID     uuid.UUID          `json:"poll_id"`
	Title      string             `json:"title" example:"Lunch"`
	TotalVotes uint               `json:"total_votes" example:"100"`
	Options    []OptionSummaryDto `json:"options"`
}

// PollRequestDto is the request body for updating a poll.
type PollRequestDto struct {
	Title       string `json:"title" validate:"required,min=2,max=100" example:"Team Lunch"`
	Description string `json:"description" validate:"max=100" example:"Where should we eat today?"`
	Public      bool   `json:"public" example:"true"`
}
