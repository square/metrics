package query

import (
	"github.com/square/metrics-indexer/api"
)

// Predicate is a boolean function applied against the given
// metric alias and tagset. It determines whether the given metric
// should be included in the query.
type Predicate interface {
	// checks the matcher.
	Match(tagSet api.TagSet) bool
}

func (matcher *andPredicate) Match(tagSet api.TagSet) bool {
	for _, subPredicate := range matcher.predicates {
		if !subPredicate.Match(tagSet) {
			return false
		}
	}
	return true
}

func (matcher *orPredicate) Match(tagSet api.TagSet) bool {
	for _, subPredicate := range matcher.predicates {
		if subPredicate.Match(tagSet) {
			return true
		}
	}
	return false
}

func (matcher *notPredicate) Match(tagSet api.TagSet) bool {
	return !matcher.predicate.Match(tagSet)
}

func (matcher *listMatcher) Match(tagSet api.TagSet) bool {
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

func (matcher *regexMatcher) Match(tagSet api.TagSet) bool {
	if !matchPrecondition(matcher.tag, tagSet) {
		return false
	}
	return matcher.regex.MatchString(tagSet[matcher.tag])
}

func matchPrecondition(matcherTag string, tagSet api.TagSet) bool {
	if _, hasTag := tagSet[matcherTag]; !hasTag {
		return false
	}
	return true
}
