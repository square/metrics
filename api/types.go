package api

// list of data types throughout the code.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"time"
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

// Equals check whether two tags are equal.
func (left TagSet) Equals(right TagSet) bool {
	if len(left) != len(right) {
		return false
	}
	for k := range left {
		_, ok := right[k]
		if !ok || left[k] != right[k] {
			return false
		}
	}
	return true
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
// are Unix milliseconds timestamps. Resolution is in milliseconds.
// (Millisecond precision allows an effective range of 290 million years in each direction)
type Timerange struct {
	start      int64
	end        int64
	resolution int64
}

// Start returns the .start field
func (tr Timerange) Start() int64 {
	return tr.start
}

// End returns the .end field
func (tr Timerange) End() int64 {
	return tr.end
}

func (tr Timerange) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Start      int64 `json:"start"`
		End        int64 `json:"end"`
		Resolution int64 `json:"resolution"`
	}{tr.start, tr.end, tr.resolution})
}

// Resolution() returns the .resolution field
func (tr Timerange) Resolution() int64 {
	return tr.resolution
}

// NewTimerange creates a timerange which is validated, providing error otherwise.
func NewTimerange(start, end, resolution int64) (Timerange, error) {
	if resolution <= 0 {
		return Timerange{}, fmt.Errorf("invalid resolution %d", resolution)
	}

	if start%resolution != 0 {
		return Timerange{}, errors.New(fmt.Sprintf("start %% resolution (mod) must be 0 (start=%d, resolution=%d)", start, resolution))
	}
	if end%resolution != 0 {
		return Timerange{}, errors.New(fmt.Sprintf("end %% resolution (mod) must be 0 (end=%d, resolution=%d)", end, resolution))
	}
	if start > end {
		return Timerange{}, errors.New(fmt.Sprintf("start must be <= end (start=%d, end=%d)", start, end))
	}
	return Timerange{start: start, end: end, resolution: resolution}, nil
}

func NewSnappedTimerange(start, end, resolution int64) (Timerange, error) {
	if resolution <= 0 {
		return Timerange{}, fmt.Errorf("invalid resolution %d", resolution)
	}
	if start > end {
		return Timerange{}, errors.New(fmt.Sprintf("start must be <= end (start=%d, end=%d)", start, end))
	}
	return Timerange{start: start, end: end, resolution: resolution}.Snap(), nil
}

func snap(n, boundary int64) int64 {
	if n < 0 {
		return -snap(-n, boundary)
	}
	// This performs a round.
	// Dividing by `boundary` truncates towards zero.
	// The resulting integer is then multiplied by `boundary` again.
	// Thus the result is a multiple of `boundary`.
	// For integer division, x/r*r = (x/r)*r in general rounds to a multiple of r towards 0.
	// Adding `boundary/2` changes this instead to a "round to nearest" rather than "round towards 0".
	// (Where "up" is the round for values exactly halfway between).
	// These halfway points round "away from zero" (rather than "towards -infinity").
	return (n + boundary/2) / boundary * boundary
}

// Round() will fix some invalid timeranges by rounding their starts and ends.
func (tr Timerange) Snap() Timerange {
	if tr.resolution == 0 {
		panic("AAH")
		return tr
	}
	tr.start = snap(tr.start, tr.resolution)
	tr.end = snap(tr.end, tr.resolution)
	return tr
}

// Later() returns a timerange which is forward in time by the amount given
func (tr Timerange) Shift(shift time.Duration) Timerange {
	tr.start += int64(shift / time.Millisecond)
	tr.end += int64(shift / time.Millisecond)
	return tr.Snap()
}

// Slots represent the total # of data points
// Behavior is undefined when operating on an invalid Timerange. There's a
// circular dependency here, but it all works out.
func (tr Timerange) Slots() int {
	return int((tr.end-tr.start)/tr.resolution) + 1
}

// Timeseries is a single time series, identified with the associated tagset.
type Timeseries struct {
	Values []float64
	TagSet TagSet
}

// This function exists to manually encode floats.
func (ts Timeseries) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	var scratch [64]byte
	buffer.WriteByte('{')
	buffer.WriteString("\"tagset\":")
	tagset, err := json.Marshal(ts.TagSet)
	if err != nil {
		return []byte{}, err
	}
	buffer.Write(tagset)
	buffer.WriteByte(',')
	buffer.WriteString("\"values\":")
	buffer.WriteByte('[')
	n := len(ts.Values)
	for i := 0; i < n; i++ {
		if i > 0 {
			buffer.WriteByte(',')
		}
		f := ts.Values[i]
		if math.IsInf(f, 1) {
			buffer.WriteString("null") // TODO - positive infinity
		} else if math.IsInf(f, -1) {
			buffer.WriteString("null") // TODO - negative infinity
		} else if math.IsNaN(f) {
			buffer.WriteString("null")
		} else {
			b := strconv.AppendFloat(scratch[:0], f, 'g', -1, 64)
			buffer.Write(b)
		}
	}
	buffer.WriteByte(']')
	buffer.WriteByte('}')
	return buffer.Bytes(), err
}

// SampleMethod determines how the given time series should be sampled.
type SampleMethod int

const (
	// SamplingMax chooses the maximum value.
	SampleMax SampleMethod = iota + 1
	// SamplingMin chooses the minimum value.
	SampleMin
	// SamplingMean chooses the average value.
	SampleMean
)

func (sm SampleMethod) String() string {
	switch sm {
	case SampleMax:
		return "SampleMax"
	case SampleMin:
		return "SampleMin"
	case SampleMean:
		return "SampleMean"
	}

	return "unknown"
}

// SeriesList is a list of time series sharing the same time range.
type SeriesList struct {
	Series    []Timeseries `json:"series"`
	Timerange Timerange    `json:"timerange"`
	Name      string       `json:"name"` // human-readable description of the given time series.
}

// IsValid determines whether the given time series is valid.
func (list SeriesList) isValid() bool {
	for _, series := range list.Series {
		// # of slots per series must be valid.
		if len(series.Values) != list.Timerange.Slots() {
			return false
		}
	}
	return true // validation is now successful.
}
