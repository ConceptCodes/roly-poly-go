package models

type HealthCheckResponseDto struct {
	Up      bool   `json:"status"`
	Service string `json:"service"`
}

type OnboardUserRequestDto struct {
	FirstName string `json:"first_name" validate:"required,alpha,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,alpha,min=2,max=100"`
}

type CreatePollRequestDto struct {
	Title       string   `json:"title" validate:"required,min=2,max=100"`
	Description string   `json:"description" validate:"max=100"`
	Public      bool     `json:"public"`
	Options     []string `json:"options" validate:"required,dive,required,min=2,max=100"`
}

type PollRequestDto struct {
	Title       string `json:"title" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=100"`
	Public      bool   `json:"public"`
}
