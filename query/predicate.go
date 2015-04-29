package query

import (
	"github.com/square/metrics-indexer/api"
)

// Predicate is a boolean function applied against the given
// metric alias and tagset. It determines whether the given metric
// should be included in the query.
type Predicate interface {
	// checks the matcher.
	Match(alias string, tagSet api.TagSet) bool
}

func (matcher *andPredicate) Match(alias string, tagSet api.TagSet) bool {
	for _, subPredicate := range matcher.predicates {
		if !subPredicate.Match(alias, tagSet) {
			return false
		}
	}
	return true
}

func (matcher *orPredicate) Match(alias string, tagSet api.TagSet) bool {
	for _, subPredicate := range matcher.predicates {
		if subPredicate.Match(alias, tagSet) {
			return true
		}
	}
	return false
}

func (matcher *notPredicate) Match(alias string, tagSet api.TagSet) bool {
	return !matcher.predicate.Match(alias, tagSet)
}

func (matcher *listMatcher) Match(alias string, tagSet api.TagSet) bool {
	if !matchPrecondition(matcher.tag, matcher.alias, alias, tagSet) {
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

func (matcher *regexMatcher) Match(alias string, tagSet api.TagSet) bool {
	if !matchPrecondition(matcher.tag, matcher.alias, alias, tagSet) {
		return false
	}
	return matcher.regex.MatchString(tagSet[matcher.tag])
}

func matchPrecondition(
	matcherTag string,
	matcherAlias string,
	alias string,
	tagSet api.TagSet) bool {
	if _, hasTag := tagSet[matcherTag]; !hasTag {
		return false
	}
	if matcherAlias == "" || alias == matcherAlias {
		return true
	}
	return false
}
