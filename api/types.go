package api

// list of data types throughout the code.

import (
	"bytes"
	"regexp"
	"sort"
)

// MetricKey is the logical name of a given metric.
// MetricKey should not contain any variable component in it.
type MetricKey string

// MetricKeys is an interface implementing sort.Interface to allow it to be sorted.
type MetricKeys []MetricKey

func (keys MetricKeys) Len() int {
	return len(keys)
}

func (keys MetricKeys) Less(i, j int) bool {
	return keys[i] < keys[j]
}

func (keys MetricKeys) Swap(i, j int) {
	keys[i], keys[j] = keys[j], keys[i]
}

// TagSet is the set of key-value pairs associated with a given metric.
type TagSet map[string]string

// NewTagSet creates a new instance of TagSet.
func NewTagSet() TagSet {
	return make(map[string]string)
}

// Merge two tagsets, and return a new tagset.
// If keys conflict, the first tag set is preferred.
func (tagSet TagSet) Merge(other TagSet) TagSet {
	result := NewTagSet()
	for key, value := range other {
		result[key] = value
	}
	for key, value := range tagSet {
		result[key] = value
	}
	return result
}

// ParseTagSet parses a given string to a tagset, nil
// if parsing failed.
func ParseTagSet(raw string) TagSet {
	result := NewTagSet()
	byteSlice := []byte(raw)
	stringPattern := `((?:[^=,\\]|\\[=,\\])+)`
	keyValuePattern := regexp.MustCompile("^" + stringPattern + "=" + stringPattern)
	for {
		matcher := keyValuePattern.FindSubmatchIndex(byteSlice)
		if matcher == nil {
			return nil
		}
		key := unescapeString(string(byteSlice[matcher[2]:matcher[3]]))
		value := unescapeString(string(byteSlice[matcher[4]:matcher[5]]))
		result[key] = value
		byteSlice = byteSlice[matcher[1]:]
		if len(byteSlice) == 0 {
			// end of input
			return result
		} else if byteSlice[0] == ',' {
			// progress to the next key-value pair.
			byteSlice = byteSlice[1:]
		} else {
			// invalid input.
			return nil
		}
	}
}

// Serialize transforms a given tagset to string-serialized form, following the spec.
func (tagSet TagSet) Serialize() string {
	var buffer bytes.Buffer
	sortedKeys := make([]string, len(tagSet))
	index := 0
	for key := range tagSet {
		sortedKeys[index] = key
		index++
	}
	sort.Strings(sortedKeys)

	for index, key := range sortedKeys {
		value := tagSet[key]
		if index != 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(escapeString(key))
		buffer.WriteString("=")
		buffer.WriteString(escapeString(value))
	}
	return buffer.String()
}

// HasKey returns true if a given tagset contains the given tag key.
func (tagSet TagSet) HasKey(key string) bool {
	_, hasTag := tagSet[key]
	return hasTag
}

// TaggedMetric is composition of a MetricKey and a TagSet.
// TaggedMetric should uniquely identify a single series of metric.
type TaggedMetric struct {
	MetricKey MetricKey
	TagSet    TagSet
}

// GraphiteMetric is a flat, dot-separated identifier to a series of metric.
type GraphiteMetric string
