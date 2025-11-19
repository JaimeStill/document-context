package logger

// Logger defines the interface for structured logging.
//
// This interface provides level-based logging methods that accept a message
// string followed by optional key-value pairs for structured logging context.
//
// Implementations must be safe for concurrent use by multiple goroutines.
//
// Example usage:
//
//	logger.Info("processing document", "path", docPath, "pages", pageCount)
//	logger.Error("conversion failed", "error", err, "document", docID)
//
// The variadic args parameter should contain alternating keys (strings) and
// values (any type) for structured logging. Most implementations use slog or
// similar structured logging libraries underneath.
type Logger interface {
	// Debug logs a debug-level message with optional key-value pairs.
	//
	// Debug messages are the most verbose and typically only enabled during
	// development or troubleshooting. They provide detailed information about
	// program execution flow.
	Debug(msg string, args ...any)

	// Info logs an informational message with optional key-value pairs.
	//
	// Info messages communicate normal operational events that are useful for
	// understanding system behavior in production.
	Info(msg string, args ...any)

	// Warn logs a warning message with optional key-value pairs.
	//
	// Warn messages indicate potential issues or unexpected conditions that
	// don't prevent the system from functioning but may require attention.
	Warn(msg string, args ...any)

	// Error logs an error message with optional key-value pairs.
	//
	// Error messages indicate failures or exceptional conditions that prevent
	// normal operation. Errors should always be logged with context to aid
	// in troubleshooting.
	Error(msg string, args ...any)
}
