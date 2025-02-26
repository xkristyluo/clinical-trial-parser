package studies

import (
	"log"
	"strings"

	"github.com/facebookresearch/clinical-trial-parser/src/common/col/set"
	"github.com/facebookresearch/clinical-trial-parser/src/common/util/slice"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/criteria"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/parser"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/relation"
)

type Study struct {
	Id                  string            `json:"study_id,omitempty"`
	Name                string            `json:"study_name,omitempty"`
	Conditions          []string          `json:"conditions,omitempty"`
	EligibilityCriteria string            `json:"eligibility_criteria,omitempty"`
	InclusionCriteria   criteria.Criteria `json:"-"`
	ExclusionCriteria   criteria.Criteria `json:"-"`
	CriteriaCnt         int               `json:"criteria_count"`
}

type ParsedStudy struct {
	Id             string                  `json:"study_id,omitempty"`
	CriteriaCnt    int                     `json:"criteria_count"`
	ParsedCriteria criteria.ParsedCriteria `json:"parsed_criteria,omitempty"`
}

func NewStudy(id, name string, conditions []string, eligibilityCriteria string) *Study {
	return &Study{Id: id, Name: name, Conditions: conditions, EligibilityCriteria: eligibilityCriteria}
}

func NewParsedStudy(id string, criteriaCnt int, ParsedCriteria criteria.ParsedCriteria) *ParsedStudy {
	return &ParsedStudy{
		Id:             id,
		CriteriaCnt:    criteriaCnt,
		ParsedCriteria: ParsedCriteria,
	}
}

func (s *Study) GetId() string {
	return s.Id
}

func (s *Study) GetName() string {
	return s.Name
}

func (s *Study) GetInclusionCriteria() criteria.Criteria {
	return s.InclusionCriteria
}

func (s *Study) GetExclusionCriteria() criteria.Criteria {
	return s.ExclusionCriteria
}

func (s *Study) CriteriaCount() int {
	return s.CriteriaCnt
}

// Parse parses eligibility criteria text to relations for the study s.
func (s *Study) Parse() *Study {
	interpreter := parser.Get()

	inclusions, exclusions := s.Criteria()
	s.CriteriaCnt = len(inclusions) + len(exclusions)

	// Parse inclusion criteria:
	inclusionCriteria := criteria.NewCriteria()
	for index, inclusion := range inclusions {
		lowercase := strings.ToLower(inclusion)
		orRelations, andRelations := interpreter.Interpret(lowercase)

		orRelations.Process()
		andRelations.Process()

		if !andRelations.Empty() {
			criterion := criteria.NewCriterion(inclusion, andRelations.MinScore(), andRelations, index)
			inclusionCriteria = append(inclusionCriteria, criterion)
		} else {
			criterion := criteria.NewCriterion(inclusion, orRelations.MinScore(), orRelations, index)
			inclusionCriteria = append(inclusionCriteria, criterion)
		}
	}
	s.InclusionCriteria = inclusionCriteria

	// Parse exclusion criteria:
	exclusionCriteria := criteria.NewCriteria()
	for index, exclusion := range exclusions {
		lowercase := strings.ToLower(exclusion)
		orRelations, andRelations := interpreter.Interpret(lowercase)
		orRelations.Process()
		andRelations.Process()
		orRelations.Negate()
		andRelations.Negate()

		if !orRelations.Empty() {
			criterion := criteria.NewCriterion(exclusion, orRelations.MinScore(), orRelations, index)
			exclusionCriteria = append(exclusionCriteria, criterion)
		} else {
			criterion := criteria.NewCriterion(exclusion, andRelations.MinScore(), andRelations, index)
			exclusionCriteria = append(exclusionCriteria, criterion)
		}

	}

	s.ExclusionCriteria = exclusionCriteria
	s.Transform()

	return s
}

// Criteria extracts inclusion and exclusion criteria from the eligibility criteria string.
func (s *Study) Criteria() ([]string, []string) {
	eligibilityCriteria := criteria.Normalize(s.EligibilityCriteria)

	// Parse inclusion criteria:
	var inclusions []string
	inclusionList := criteria.ExtractInclusionCriteria(eligibilityCriteria)
	for _, s := range inclusionList {
		inclusions = append(inclusions, criteria.Split(s)...)
	}

	for i, c := range inclusions {
		inclusions[i] = criteria.TrimCriterion(c)
	}
	inclusions = slice.RemoveEmpty(inclusions)

	// Parse exclusion criteria:
	var exclusions []string
	exclusionList := criteria.ExtractExclusionCriteria(eligibilityCriteria)
	for _, s := range exclusionList {
		exclusions = append(exclusions, criteria.Split(s)...)
	}

	for i, c := range exclusions {
		exclusions[i] = criteria.TrimCriterion(c)
	}
	exclusions = slice.RemoveEmpty(exclusions)

	return inclusions, exclusions
}

// Transform transforms criteria relations by converting parsed values to strings of valid literals.
// If a valid literal cannot be inferred, the confidence score of the relation is set to zero.
func (s *Study) Transform() {
	s.InclusionCriteria.Relations().Transform()
	s.ExclusionCriteria.Relations().Transform()
}

// Relations returns the string representation of the parsed criteria.
// Relations that are parsed from the same criterion and are conjoined
// by 'or' have the same criterion id (cid).
func (s *Study) Relations() criteria.ParsedCriteria {
	// variableCatalog := variables.Get()
	// r.VariableType.String()
	var pc criteria.ParsedCriteria
	cid := 0
	for _, c := range s.InclusionCriteria {
		relationR := c.Relations()
		log.Printf("========test%+v", c)
		if len(relationR) > 0 {
			p := criteria.NewParsedCriterion("inclusion", "", c.ClusterID, c.String(), "", relationR)
			pc = append(pc, p)
		} else {
			p := criteria.NewParsedCriterion("inclusion", "", c.ClusterID, c.String(), "", relation.Relations{})
			pc = append(pc, p)
		}
		cid++
	}
	for _, c := range s.ExclusionCriteria {
		relationR := c.Relations()
		if len(relationR) > 0 {
			p := criteria.NewParsedCriterion("exclusion", "", c.ClusterID, c.String(), "", relationR)
			pc = append(pc, p)
		} else {
			p := criteria.NewParsedCriterion("exclusion", "", c.ClusterID, c.String(), "", relation.Relations{})
			pc = append(pc, p)
		}
		cid++
	}

	return pc
	// return pc.JSON()
}

// ParsedCriteriaCount returns the number of parsed unique criteria.
func (s *Study) ParsedCriteriaCount() int {
	parsedCriteria := set.New()
	for _, c := range s.InclusionCriteria {
		for _, r := range c.Relations() {
			if r.Valid() {
				parsedCriteria.Add(c.String())
				break
			}
		}
	}
	for _, c := range s.ExclusionCriteria {
		for _, r := range c.Relations() {
			if r.Valid() {
				parsedCriteria.Add(c.String())
				break
			}
		}
	}
	return parsedCriteria.Size()
}

// RelationCount returns the number of parsed relations.
func (s *Study) RelationCount() int {
	cnt := 0
	for _, c := range s.InclusionCriteria {
		for _, r := range c.Relations() {
			if r.Valid() {
				cnt++
			}
		}
	}
	for _, c := range s.ExclusionCriteria {
		for _, r := range c.Relations() {
			if r.Valid() {
				cnt++
			}
		}
	}
	return cnt
}
