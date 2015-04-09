// Package api holds common data types and public interface exposed by the indexer library.
package api

// Refer to the doc
// https://docs.google.com/a/squareup.com/document/d/1k0Wgi2wnJPQoyDyReb9dyIqRrD8-v0u8hz37S282ii4/edit
// for the terminology.

// MetricKey is the logical name of a given metric.
// MetricKey should not contain any variable component in it.
type MetricKey string

// TagSet is the set of key-value pairs associated with a given metric.
type TagSet map[string]string

// TaggedMetric is composition of a MetricKey and a TagSet.
// TaggedMetric should uniquely identify a single series of metric.
type TaggedMetric struct {
	MetricKey MetricKey
	TagSet    TagSet
}

// GraphiteMetric is a flat, dot-separated identifier to a series of metric.
type GraphiteMetric string

// API is the set of public methods exposed by the indexer library.
type API interface {
	// AddMetric adds the metric to the system.
	AddMetric(metric TaggedMetric) error

	// RemoveMetric removes the metric from the system.
	RemoveMetric(metric TaggedMetric) error

	// Convert the given tag-based metric name to graphite metric name,
	// using the configured rules. May error out.
	ToGraphiteName(metric TaggedMetric) (GraphiteMetric, error)

	// Converts the given graphite metric to the tag-based meric,
	// using the configured rules. May error out.
	ToTaggedName(metric GraphiteMetric) (TaggedMetric, error)

	// For a given MetricKey, retrieve all the tagsets associated with it.
	GetAllTags(metricKey MetricKey) []TagSet

	// For a given tag key-value pair, obtain the list of all the MetricKeys
	// associated with them.
	GetMericsForTag(tagKey string, tagValue string) []MetricKey
}
