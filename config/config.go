package config

type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"INFO"`

	Host     string `envconfig:"HOST" default:"0.0.0.0"`
	Port     string `envconfig:"PORT" default:"11140"`
	Database string `envconfig:"DATABASE" default:"comments.db"`

	AllowedOrigins []string `envconfig:"ALLOWED_ORIGINS" default:"*"`
	AdminKey       string   `envconfig:"ADMIN_KEY"`
}
