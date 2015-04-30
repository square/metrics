package query

import (
	"github.com/square/metrics/api"
)

// Predicate is a boolean function applied against the given
// metric alias and tagset. It determines whether the given metric
// should be included in the query.
type Predicate interface {
	// checks the matcher.
	Apply(tagSet api.TagSet) bool
}

func (matcher *andPredicate) Apply(tagSet api.TagSet) bool {
	for _, subPredicate := range matcher.predicates {
		if !subPredicate.Apply(tagSet) {
			return false
		}
	}
	return true
}

func (matcher *orPredicate) Apply(tagSet api.TagSet) bool {
	for _, subPredicate := range matcher.predicates {
		if subPredicate.Apply(tagSet) {
			return true
		}
	}
	return false
}

func (matcher *notPredicate) Apply(tagSet api.TagSet) bool {
	return !matcher.predicate.Apply(tagSet)
}

func (matcher *listMatcher) Apply(tagSet api.TagSet) bool {
	if !matchPrecondition(matcher.tag, tagSet) {
		return false
	}
	tagValue := tagSet[matcher.tag]
	for _, match := range matcher.matches {
		if match == tagValue {
			return true
		}
	}
	return false
}

func (matcher *regexMatcher) Apply(tagSet api.TagSet) bool {
	if !matchPrecondition(matcher.tag, tagSet) {
		return false
	}
	return matcher.regex.MatchString(tagSet[matcher.tag])
}

func matchPrecondition(matcherTag string, tagSet api.TagSet) bool {
	return tagSet.HasKey(matcherTag)
}
