package internal

// List of errors returnd by the application

import (
	"fmt"
)

// RuleErrorCode is the error enum raised while YAML rule files is being compiled.
type RuleErrorCode int

const (
	// InvalidYaml is returned when the given YAML file fails to parse
	InvalidYaml RuleErrorCode = iota + 1
	// InvalidPattern is returned when an invalid rule pattern is provided.
	InvalidPattern
	// InvalidMetricKey is retruned when an invalid metric key is provided.
	InvalidMetricKey
	// InvalidCustomRegex is retruned when the custom regex is invalid.
	InvalidCustomRegex
)

// ConversionErrorCode is the error enum raised while the metrics are converted
// between the graphite format to tagged metrics format.
type ConversionErrorCode int

const (
	// MissingTag is returned during the reverse mapping, when a pattern required in the graphite key is not provided.
	MissingTag ConversionErrorCode = iota + 1
	// CannotInterpolate is returned when the tag interpolation fails.
	CannotInterpolate
	// NoMatch is returned when no rule can reverse the given tagged metric.
	NoMatch
)

// RuleError is the actual error object, wrapping RuleErrorCode and related metadata.
type RuleError interface {
	// Error code describing the error.
	Code() RuleErrorCode
	// Rule Metric Key, if applicable
	MetricKey() string

	Error() string
}

// ConversionError is the actual error object, wrapping ConversionErrorCode and related metadata.
type ConversionError interface {
	Code() ConversionErrorCode
	Error() string
}

// Implementations
// ===============

type ruleError struct {
	code      RuleErrorCode
	metricKey string
	message   string
}

type conversionError struct {
	code    ConversionErrorCode
	message string
}

func (err ruleError) Code() RuleErrorCode {
	return err.code
}

func (err ruleError) MetricKey() string {
	return err.metricKey
}

func (err ruleError) Error() string {
	return err.message
}

func newInvalidPattern(metricKey string) RuleError {
	return ruleError{InvalidPattern, metricKey, fmt.Sprintf("Invalid metric key '%s'", metricKey)}
}
func newInvalidMetricKey(metricKey string) RuleError {
	return ruleError{InvalidMetricKey, metricKey, fmt.Sprintf("Invalid pattern in key '%s'", metricKey)}
}
func newInvalidCustomRegex(metricKey string) RuleError {
	return ruleError{InvalidCustomRegex, metricKey, fmt.Sprintf("Invalid custom regex in key '%s'", metricKey)}
}

func (err conversionError) Code() ConversionErrorCode {
	return err.code
}

func (err conversionError) Error() string {
	return err.message
}

func newMissingTag(tag string) ConversionError {
	return conversionError{
		MissingTag,
		fmt.Sprintf("Missing tag '%s'", tag),
	}
}

func newCannotInterpolate() ConversionError {
	return conversionError{
		CannotInterpolate,
		"Cannot interpolate",
	}
}

func newNoMatch() ConversionError {
	return conversionError{
		NoMatch,
		"No match",
	}
}

// ensure interface
var _ RuleError = (*ruleError)(nil)
