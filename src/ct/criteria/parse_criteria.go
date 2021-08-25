// Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved.

package criteria

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"

	"github.com/facebookresearch/clinical-trial-parser/src/common/util/text"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/relation"
)

var (
	reDeleteCriterion = regexp.MustCompile(`(?i)([^\n]+meet inclusion criteria|[^\n]*inclusion/exclusion criteria)\W? *(\n|$)`)
	reMatchInclusions = regexp.MustCompile(`(?is)inclusions?(?: *:| criteria(?:[^:\n]*?:| *\n))(.*?)(?:[^\n]*\bexclusions?(?: *:| criteria(?:[^:\n]*?:| *\n))|$)`)
	reMatchExclusions = regexp.MustCompile(`(?is)exclusions?(?: *:| criteria(?:[^:\n]*?:| *\n))(.*?)(?:[^\n]*\binclusions?(?: *:| criteria(?:[^:\n]*?:| *\n))|$)`)

	reCriteriaSplitter = regexp.MustCompile(`\n\n`)
	reTrimmer          = regexp.MustCompile(`^(\s*-\s*)?(\s*\d+\.?\s*)?`)

	reMatchTabs       = regexp.MustCompile(`the following(\s+criteria)?(\s*:)?\s*\n\s*(-|\d+\.|[a-z]\s)\s*`)
	reMatchTabLine    = regexp.MustCompile(`the following`)
	reMatchBulletLine = regexp.MustCompile(`^\s*(-|\d+\.|[a-z]\s)\s*`)

	empty = []string{}
)

type ParsedCriterion struct {
	EligibilityType string            `json:"eligibility_type,omitempty"` // inclusion or exclusion
	VariableType    string            `json:"variable_type,omitempty"`    // numerical or ordinal
	CriterionIndex  int               `json:"criterion_index"`
	Criterion       string            `json:"criterion,omitempty"`
	Question        string            `json:"question,omitempty"`
	Relation        relation.Relation `json:"relation,omitempty"`
}

type ParsedCriteria []*ParsedCriterion

func NewParsedCriterion(eligibilityType, variableType string, criterionIndex int, criterion, question string, relation relation.Relation) *ParsedCriterion {
	return &ParsedCriterion{
		EligibilityType: eligibilityType,
		VariableType:    variableType,
		CriterionIndex:  criterionIndex,
		Criterion:       criterion,
		Question:        question,
		Relation:        relation,
	}
}

func (p *ParsedCriteria) JSON() string {
	if data, err := json.Marshal(p); err == nil {
		return string(data)
	}
	return ""
}

// Normalize normalizes eligibility criteria text. For now, non-informative
// "Does not meet inclusion criteria" like criteria are removed.
func Normalize(s string) string {
	s = reDeleteCriterion.ReplaceAllString(s, "")
	return s
}

func PrintNew(text string, input []string) {
	log.Printf("%s===============", text)
	for index, value := range input {
		log.Printf("%d: %v", index, value)
	}
}

// ExtractInclusionCriteria extracts a block of inclusion criteria from the string.
func ExtractInclusionCriteria(s string) []string {
	return extractCriteria(s, reMatchInclusions)
}

// ExtractExclusionCriteria extracts a block of exclusion criteria from the string.
func ExtractExclusionCriteria(s string) []string {
	return extractCriteria(s, reMatchExclusions)
}

func extractCriteria(s string, r *regexp.Regexp) []string {
	values := r.FindAllStringSubmatch(s, -1)
	if len(values) == 0 {
		return empty
	}
	c := []string{}
	for _, value := range values {
		if len(value) == 2 {
			if v := strings.TrimSpace(value[1]); len(v) > 0 {
				c = append(c, v)
			}
		}
	}
	return c
}

// Split splits eligibility criteria numberings into individual criteria.
func Split(s string) []string {
	rules := reCriteriaSplitter.Split(s, -1)
	PrintNew("Split", rules)
	numTabs, header, foundTab := initLine(s)
	if numTabs == 0 {
		return rules
	}
	var newRules []string
	for _, rule := range rules {
		if rule, header, foundTab = checkLine(rule, header, foundTab); len(rule) > 0 {
			newRules = append(newRules, rule)
		}
	}
	return newRules
}

// TrimCriterion normalizes the criterion by removing leading bullets,
// numberings, and all leading and trailing punctuation.
func TrimCriterion(s string) string {
	s = reTrimmer.ReplaceAllString(s, "")
	s = text.NormalizeWhitespace(s)
	s = strings.Trim(s, ` ,.;:/"`)
	return s
}

func checkLine(rule string, header string, foundTab bool) (string, string, bool) {
	// found a bullet for a previously seen header
	if foundTab && reMatchBulletLine.MatchString(rule) {
		rule = header + " " + TrimCriterion(rule)

		// found a header
	} else if reMatchTabLine.MatchString(rule) {
		foundTab = true
		header = rule
		rule = ""

		// normal criteria
	} else {
		foundTab = false
		header = ""
	}

	return rule, header, foundTab
}

func initLine(eligibilities string) (int, string, bool) {
	numTabs := len(reMatchTabs.FindAllStringIndex(eligibilities, -1))
	header := ""
	foundTab := false
	return numTabs, header, foundTab
}
