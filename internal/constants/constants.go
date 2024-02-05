package constants

const (
	// Error codes
	NotFound            = "RP-404"
	BadRequest          = "RP-400"
	Unauthorized        = "RP-401"
	Forbidden           = "RP-403"
	InternalServerError = "RP-500"

	// Endpoints
	ApiPrefix           = "/api"
	HealthCheckEndpoint = ApiPrefix + "/health/alive"
	ReadinessEndpoint   = ApiPrefix + "/health/status"
	OnboardUserEndpoint = ApiPrefix + "/onboard"
	GetPollsEndpoint    = ApiPrefix + "/polls"
	CreatePollEndpoint  = GetPollsEndpoint
	PollByIdEndpoint    = GetPollsEndpoint + "/{id}"
	ClosePollEndpoint   = PollByIdEndpoint
	UpdatePollEndpoint  = PollByIdEndpoint

	// Messages
	EntityNotFound        = "%s with id %d does not exist."
	GetEntityByIdMessage  = "Found %s with id %d."
	SaveEntityError       = "Error while saving %s."
	SuccessMessage        = "You have successfully %s!"
	GetAllEntitiesError   = "Error while fetching all %s."
	GetAllEntitiesMessage = "Found %d %s."
	CreateEntityError     = "Error while creating %s."
	CreateEntityMessage   = "Created %s successfully."
	UpdateEntityError     = "Error while updating %s."
	UpdateEntityMessage   = "Updated %s successfully."

	// Queries
	FindByIdQuery          = "id = ?"
	FindByApiKeyQuery      = "api_key = ?"
	FindByUserIdAndIdQuery = "user_id = ? AND id = ?"

	// Misc
	TimeFormat          = "2006-01-02 15:04:05"
	TraceIdHeader       = "x-trace-id"
	AuthorizationHeader = "x-api-key"
	HealthCheckMessage  = "Performing healthcheck for service: %s"
	DBTablePrefix       = "roly_poly_%s"
	Locale              = "en"
	LocalEnv            = "local"
	DevelopmentEnv      = "development"
	ProductionEnv       = "prod"
	StartMessage        = "Starting API Service on PORT=%s | ENV=%s"
	RequestIdCtxKey     = "request_id"
	ApiKeyCtxKey        = "api_key"
	UserIdCtxKey        = "user_id"

	// Errors
	HealthCheckError = "Error while performing healthcheck for service: %s"
)
