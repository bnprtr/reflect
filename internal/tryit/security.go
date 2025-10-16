package tryit

import (
	"fmt"
	"strings"
)

// SensitiveHeaders is a list of headers that should never be logged or displayed.
var SensitiveHeaders = []string{
	"authorization",
	"cookie",
	"set-cookie",
	"proxy-authorization",
	"www-authenticate",
	"x-api-key",
	"api-key",
}

// FilterHeaders filters headers through an allowlist.
// If the allowlist is empty, all headers are allowed.
// Returns a new map with only allowed headers (case-insensitive matching).
func FilterHeaders(headers map[string]string, allowlist []string) map[string]string {
	if len(allowlist) == 0 {
		// No allowlist means permit all
		return headers
	}

	// Build lowercase allowlist for case-insensitive matching
	allowedLower := make(map[string]bool)
	for _, h := range allowlist {
		allowedLower[strings.ToLower(h)] = true
	}

	filtered := make(map[string]string)
	for key, value := range headers {
		if allowedLower[strings.ToLower(key)] {
			filtered[key] = value
		}
	}

	return filtered
}

// RedactSensitiveHeaders removes sensitive header values from a header map.
// Returns a new map with sensitive values replaced with "[REDACTED]".
func RedactSensitiveHeaders(headers map[string][]string) map[string][]string {
	redacted := make(map[string][]string)

	// Build a set of sensitive header names (lowercase)
	sensitiveSet := make(map[string]bool)
	for _, h := range SensitiveHeaders {
		sensitiveSet[strings.ToLower(h)] = true
	}

	for key, values := range headers {
		if sensitiveSet[strings.ToLower(key)] {
			// Replace with redacted placeholder
			redacted[key] = []string{"[REDACTED]"}
		} else {
			// Copy values as-is
			redacted[key] = make([]string, len(values))
			copy(redacted[key], values)
		}
	}

	return redacted
}

// RedactSensitiveHeadersSingle is like RedactSensitiveHeaders but for map[string]string.
func RedactSensitiveHeadersSingle(headers map[string]string) map[string]string {
	redacted := make(map[string]string)

	// Build a set of sensitive header names (lowercase)
	sensitiveSet := make(map[string]bool)
	for _, h := range SensitiveHeaders {
		sensitiveSet[strings.ToLower(h)] = true
	}

	for key, value := range headers {
		if sensitiveSet[strings.ToLower(key)] {
			redacted[key] = "[REDACTED]"
		} else {
			redacted[key] = value
		}
	}

	return redacted
}

// MergeHeaders merges two header maps, with override taking precedence.
// Case is preserved from the override map.
func MergeHeaders(base, override map[string]string) map[string]string {
	result := make(map[string]string)

	// Copy base headers
	for k, v := range base {
		result[k] = v
	}

	// Override with provided headers
	for k, v := range override {
		result[k] = v
	}

	return result
}

// ValidateJSONSize checks if the JSON body size is within limits.
func ValidateJSONSize(jsonBody string, maxBytes int64) error {
	size := int64(len(jsonBody))
	if maxBytes > 0 && size > maxBytes {
		return fmt.Errorf("request body size %d bytes exceeds limit of %d bytes", size, maxBytes)
	}
	return nil
}

// IsSensitiveHeader returns true if the header name is considered sensitive.
func IsSensitiveHeader(name string) bool {
	nameLower := strings.ToLower(name)
	for _, sensitive := range SensitiveHeaders {
		if nameLower == strings.ToLower(sensitive) {
			return true
		}
	}
	return false
}
