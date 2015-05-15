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

// SeriesType is a different aspect of data.
// For example, Blueflood may stores (min / max / average / count) during rollups,
// and these data are exposed via columns
type SeriesType string

// Timerange represents a range of time a given time series is defined in:
// it is 3-tuple of (start, end, resolution) with the following constraints:
// start <= end
// start = 0 mod resolution
// end =   0 mod resolution
//
// This range is inclusive of Start and End (i.e. [Start, End]). Start and End
// are Unix second timestamps. Resolution is in seconds.
// TODO: Make these ms
type Timerange struct {
	Start      int64
	End        int64
	Resolution int64
}

// IsValid determines whether the given timerange meets the constraint.
func (tr Timerange) IsValid() bool {
	return (tr.Start%tr.Resolution == 0 &&
		tr.End%tr.Resolution == 0 &&
		tr.Resolution > 0 &&
		tr.Start <= tr.End)
}

// Slots represent the total # of data points
// Behavior is undefined when operating on an invalid Timerange. There's a
// circular dependency here, but it all works out.
func (tr Timerange) Slots() int {
	return int((tr.End-tr.Start)/tr.Resolution) + 1
}

// Timeseries is a single time series, identified with the associated tagset.
type Timeseries struct {
	Values []float64
	Metric TaggedMetric
}

// SampleMethod determines how the given time series should be sampled.
// Note(This is currently unused).
type SampleMethod int

const (
	// SamplingMax chooses the maximum value.
	SampleMax SampleMethod = iota + 1
	// SamplingMin chooses the minimum value.
	SampleMin
	// SamplingMean chooses the average value.
	SampleMean
)

// SeriesList is a list of time series sharing the same time range.
type SeriesList struct {
	Series    []Timeseries
	Timerange Timerange
}

// IsValid determines whether the given time series is valid.
func (list SeriesList) IsValid() bool {
	if !list.Timerange.IsValid() {
		// timerange must be valid.
		return false
	}
	for _, series := range list.Series {
		// # of slots per series must be valid.
		if len(series.Values) != list.Timerange.Slots() {
			return false
		}
	}
	return true // validation is now successful.
}

// Sample converts the given serieslist to comform with the provided sampling strategy.
func (list SeriesList) Sample(timerange Timerange, sampleMethod SampleMethod) SeriesList {
	resultTimeseries := make([]float64, timerange.Slots())
	_ = resultTimeseries
	return list
}
