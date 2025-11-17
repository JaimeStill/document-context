// Package logger provides structured logging interfaces and implementations.
//
// This package defines the Logger interface for application-wide structured logging,
// along with a standard implementation (Slogger) using Go's log/slog package.
//
// # Design Philosophy
//
// The logger package follows the Interface-Based Layer Interconnection principle:
//   - Logger interface defines the contract for structured logging
//   - Slogger provides a concrete implementation using log/slog
//   - Applications can provide custom implementations if needed
//
// The interface ensures logging abstraction throughout the application, preventing
// direct dependencies on specific logging libraries.
//
// # Configuration Pattern
//
// Logger initialization follows the Configuration Transformation Pattern (Type 1):
//  1. Define config.LoggerConfig with desired settings
//  2. Transform to Logger interface via NewSlogger()
//  3. Configuration is discarded; logger interface persists
//
// # Usage Example
//
// Basic usage with default configuration:
//
//	cfg := config.DefaultLoggerConfig()
//	logger, err := logger.NewSlogger(cfg, os.Stderr)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	logger.Info("application started", "version", version)
//	logger.Debug("processing document", "path", docPath, "pages", pageCount)
//	logger.Error("conversion failed", "error", err, "document", docID)
//
// Disabled logging for performance-critical code:
//
//	cfg := config.DisabledLoggerConfig()
//	logger, _ := logger.NewSlogger(cfg, io.Discard)
//	// All logging calls become no-ops with minimal overhead
//
// Custom configuration with JSON output:
//
//	cfg := config.LoggerConfig{
//	    Level:  config.LogLevelDebug,
//	    Format: "json",
//	}
//	file, _ := os.Create("app.log")
//	logger, _ := logger.NewSlogger(cfg, file)
//
// # Log Levels
//
// The Logger interface supports four severity levels:
//   - Debug: Detailed diagnostic information for development
//   - Info: General operational events and status messages
//   - Warn: Unexpected conditions that don't prevent operation
//   - Error: Failures and exceptional conditions
//
// Plus a special "disabled" level that suppresses all output.
//
// # Thread Safety
//
// All Logger implementations must be safe for concurrent use by multiple
// goroutines. The Slogger implementation achieves this by wrapping slog.Logger,
// which is itself thread-safe.
//
// # Structured Logging
//
// Logger methods accept key-value pairs for structured context:
//
//	logger.Info("message", "key1", value1, "key2", value2)
//
// The variadic args should contain alternating keys (strings) and values (any type).
// This enables rich structured logging for analysis and debugging.
package logger
