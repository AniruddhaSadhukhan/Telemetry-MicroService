package models

type contextKey string

const (

	// Context Keys
	UserContextKey = contextKey("User")

	// Users
	InternalUser = "__INTERNAL__"
	GuestUser    = "__GUEST__"

	//Collection Names
	TelemetryCollection       = "telemetry_events"
	SchemaMigrationCollection = "schema_migrations"
	TokenCollection           = "tokens"
)
