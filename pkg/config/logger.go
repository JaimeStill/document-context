package config

// LogLevel represents the minimum severity level for log output.
//
// Log levels are ordered from most verbose (debug) to least verbose (error).
// The disabled level suppresses all logging output.
type LogLevel string

const (
	// LogLevelDebug enables all log output including debug messages.
	LogLevelDebug LogLevel = "debug"

	// LogLevelInfo enables informational messages and above (info, warn, error).
	LogLevelInfo LogLevel = "info"

	// LogLevelWarn enables warning messages and above (warn, error).
	LogLevelWarn LogLevel = "warn"

	// LogLevelError enables only error messages.
	LogLevelError LogLevel = "error"

	// LogLevelDisabled suppresses all log output for maximum performance.
	// When set, the logger automatically uses io.Discard.
	LogLevelDisabled LogLevel = "disabled"
)

// LoggerConfig defines configuration for logger initialization.
//
// This configuration follows the Configuration Transformation Pattern (Type 1).
// It is ephemeral data that transforms into a logger.Logger instance via
// logger.NewSlogger() and is discarded after initialization.
//
// The configuration supports JSON serialization for external configuration files.
type LoggerConfig struct {
	// Level specifies the minimum log level. Defaults to LogLevelInfo.
	Level LogLevel `json:"level,omitempty"`

	// Format specifies the log output format: "text" or "json".
	// Defaults to "text".
	Format string `json:"format,omitempty"`
}

// DefaultLoggerConfig returns a LoggerConfig with sensible defaults.
//
// Default configuration:
//   - Level: LogLevelInfo (informational messages and above)
//   - Format: "text" (human-readable output)
//
// This configuration is suitable for most production applications.
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:  LogLevelInfo,
		Format: "text",
	}
}

// DisabledLoggerConfig returns a LoggerConfig that suppresses all logging.
//
// When using this configuration, logger.NewSlogger automatically redirects
// output to io.Discard for maximum performance with minimal overhead.
//
// Use this for performance-critical applications or when logging is not needed.
func DisabledLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:  LogLevelDisabled,
		Format: "text",
	}
}

// Finalize applies default values to any unset configuration fields.
//
// This method implements the Configuration Transformation Pattern's finalization
// step, ensuring all configuration fields have valid values before validation
// and transformation to a logger instance.
//
// Fields that are already set (non-zero values) are preserved unchanged.
func (c *LoggerConfig) Finalize() {
	defaults := DefaultLoggerConfig()
	if c.Level == "" {
		c.Level = defaults.Level
	}
	if c.Format == "" {
		c.Format = defaults.Format
	}
}
