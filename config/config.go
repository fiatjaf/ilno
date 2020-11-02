package config

type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"INFO"`

	Host     string `envconfig:"HOST" default:"0.0.0.0"`
	Port     string `envconfig:"PORT" default:"11140"`
	Database string `envconfig:"DATABASE" default:"comments.db"`
	SeedAuth string `envconfig:"SEED_AUTH" default:"https://seed-auth.etleneum.com"`

	AllowedOrigins []string `envconfig:"ALLOWED_ORIGINS" default:"*"`
	AdminKey       string   `envconfig:"ADMIN_KEY"`
}
