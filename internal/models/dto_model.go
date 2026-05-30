package models

import "github.com/google/uuid"

type HealthCheckResponseDto struct {
	Status  bool   `json:"status"`
	Service string `json:"service"`
}

type OnboardUserRequestDto struct {
	FirstName string `json:"first_name" validate:"required,alpha,min=2,max=100,noSQLKeywords"`
	LastName  string `json:"last_name" validate:"required,alpha,min=2,max=100,noSQLKeywords"`
}

type OnboardUserResponseDto struct {
	ID        uuid.UUID `json:"id"`
	ApiKey    uuid.UUID `json:"api_key"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

type CreatePollRequestDto struct {
	Title       string   `json:"title" validate:"required,min=2,max=100,noSQLKeywords"`
	Description string   `json:"description" validate:"max=100,noSQLKeywords"`
	Public      bool     `json:"public"`
	Options     []string `json:"options" validate:"required,dive,required,min=2,max=100,noSQLKeywords"`
}

type CastVoteRequestDto struct {
	OptionIDs []uuid.UUID `json:"option_ids" validate:"required,min=1"`
}

type OptionSummaryDto struct {
	OptionID   uuid.UUID `json:"option_id"`
	Label      string    `json:"label"`
	Votes      uint      `json:"votes"`
	Percentage float64   `json:"percentage"`
}

type PollReportDto struct {
	PollID     uuid.UUID          `json:"poll_id"`
	Title      string             `json:"title"`
	TotalVotes uint               `json:"total_votes"`
	Options    []OptionSummaryDto `json:"options"`
}

type PollRequestDto struct {
	Title       string `json:"title" validate:"required,min=2,max=100,noSQLKeywords"`
	Description string `json:"description" validate:"max=100,noSQLKeywords"`
	Public      bool   `json:"public"`
}
