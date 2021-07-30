package studies

import (
	"strings"

	"github.com/facebookresearch/clinical-trial-parser/src/common/col/set"
	"github.com/facebookresearch/clinical-trial-parser/src/common/util/slice"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/criteria"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/parser"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/relation"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/variables"
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
	Id            string                 `json:"study_id,omitempty"`
	CriteriaCnt   int                    `json:"criteria_count"`
	ParsedCritera criteria.ParsedCritera `json:"parsed_critera,omitempty"`
}

func NewStudy(id, name string, conditions []string, eligibilityCriteria string) *Study {
	return &Study{Id: id, Name: name, Conditions: conditions, EligibilityCriteria: eligibilityCriteria}
}

func NewParsedStudy(id string, criteriaCnt int, parsedCritera criteria.ParsedCritera) *ParsedStudy {
	return &ParsedStudy{
		Id:            id,
		CriteriaCnt:   criteriaCnt,
		ParsedCritera: parsedCritera,
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
	for _, inclusion := range inclusions {
		lowercase := strings.ToLower(inclusion)
		orRelations, andRelations := interpreter.Interpret(lowercase)

		orRelations.Process()
		andRelations.Process()

		if !orRelations.Empty() {
			criterion := criteria.NewCriterion(inclusion, orRelations.MinScore(), orRelations)
			inclusionCriteria = append(inclusionCriteria, criterion)
		}
		if !andRelations.Empty() {
			for _, r := range andRelations {
				rs := relation.Relations{r}
				criterion := criteria.NewCriterion(inclusion, rs.MinScore(), rs)
				inclusionCriteria = append(inclusionCriteria, criterion)
			}
		}
	}
	s.InclusionCriteria = inclusionCriteria

	// Parse exclusion criteria:
	exclusionCriteria := criteria.NewCriteria()
	for _, exclusion := range exclusions {
		lowercase := strings.ToLower(exclusion)
		orRelations, andRelations := interpreter.Interpret(lowercase)
		orRelations.Process()
		andRelations.Process()
		orRelations.Negate()
		andRelations.Negate()

		if !andRelations.Empty() {
			criterion := criteria.NewCriterion(exclusion, andRelations.MinScore(), andRelations)
			exclusionCriteria = append(exclusionCriteria, criterion)
		}
		if !orRelations.Empty() {
			for _, r := range orRelations {
				rs := relation.Relations{r}
				criterion := criteria.NewCriterion(exclusion, rs.MinScore(), rs)
				exclusionCriteria = append(exclusionCriteria, criterion)
			}
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
func (s *Study) Relations() criteria.ParsedCritera {
	variableCatalog := variables.Get()
	var pc criteria.ParsedCritera
	cid := 0
	for _, c := range s.InclusionCriteria {
		for _, r := range c.Relations() {
			q := variableCatalog.Question(r.ID)
			p := criteria.NewParsedCriterion("inclusion", r.VariableType.String(), cid, c.String(), q, *r)
			pc = append(pc, p)
		}
		cid++
	}
	for _, c := range s.ExclusionCriteria {
		for _, r := range c.Relations() {
			q := variableCatalog.Question(r.ID)
			p := criteria.NewParsedCriterion("exclusion", r.VariableType.String(), cid, c.String(), q, *r)
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
