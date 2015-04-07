package api

type MetricKey string

type TagSet map[string]string

type TaggedMetric struct {
	MetricKey MetricKey
	TagSet    TagSet
}

type GraphiteMetric string

type Api struct {
}

func (api *Api) AddMetric(metric TaggedMetric) error {
  return nil
}

func (api *Api) ToGraphiteName(metric TaggedMetric) GraphiteMetric {
  return ""
}

func (api *Api) ToTaggedName(metric GraphiteMetric) TaggedMetric {
  return TaggedMetric{}
}

func (api *Api) GetAllTags(metricKey MetricKey) []TagSet {
  return nil
}

func (api *Api) GetMericsForTag(tagKey string, tagValue string) []MetricKey {
  return nil
}
