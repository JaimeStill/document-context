package config

import "fmt"

// ParseString extracts a string value from an options map with fallback support.
//
// This function provides type-safe extraction of string configuration values from
// the generic Options map used in configuration composition patterns. It validates
// that the value is a non-empty string if present.
//
// Parameters:
//   - options: The map[string]any to extract from
//   - key: The configuration key to look up
//   - fallback: Default value returned when key is not present
//
// Returns the string value from options, or fallback if key is absent.
// Returns an error if the key exists but the value is not a string or is empty.
//
// Example:
//
//	background, err := ParseString(cfg.Options, "background", "white")
//	if err != nil {
//	    return nil, fmt.Errorf("invalid background: %w", err)
//	}
func ParseString(options map[string]any, key, fallback string) (string, error) {
	if value, ok := options[key]; ok {
		result, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("%s must be a string", key)
		}
		if result == "" {
			return "", fmt.Errorf("%s cannot be empty", key)
		}
		return result, nil
	}
	return fallback, nil
}

// ParseNilIntRanged extracts an optional integer value with range validation.
//
// This function provides type-safe extraction of optional integer configuration
// values from the generic Options map. It handles JSON unmarshaling's float64
// representation of numbers and validates the value falls within the specified range.
//
// The function returns nil when the key is absent, enabling distinction between
// "not configured" (nil) and "explicitly set to a value" (non-nil pointer).
//
// Parameters:
//   - options: The map[string]any to extract from
//   - key: The configuration key to look up
//   - low: Minimum valid value (inclusive)
//   - high: Maximum valid value (inclusive)
//
// Returns nil if the key is absent (not configured).
// Returns a pointer to the integer value if present and within range.
// Returns an error if the key exists but:
//   - The value is not an integer or float64
//   - The value is outside the specified range [low, high]
//
// JSON Handling:
// JSON unmarshaling represents all numbers as float64. This function accepts
// both int and float64 types, converting float64 to int when appropriate.
//
// Example:
//
//	brightness, err := ParseNilIntRanged(cfg.Options, "brightness", 0, 200)
//	if err != nil {
//	    return nil, fmt.Errorf("invalid brightness: %w", err)
//	}
//	if brightness != nil {
//	    // Use *brightness value
//	}
func ParseNilIntRanged(options map[string]any, key string, low, high int) (*int, error) {
	if value, ok := options[key]; ok {
		result, ok := value.(int)
		if !ok {
			if f, ok := value.(float64); ok {
				result = int(f)
			} else {
				return nil, fmt.Errorf("%s must be an integer", key)
			}
		}
		if result < low || result > high {
			return nil, fmt.Errorf("%s must be %d-%d, got %d", key, low, high, result)
		}
		return &result, nil
	}
	return nil, nil
}
