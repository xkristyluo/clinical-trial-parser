package cfg

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/facebookresearch/clinical-trial-parser/src/common/conf"
	"github.com/facebookresearch/clinical-trial-parser/src/common/util/timer"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/studies"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/units"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/variables"

	"github.com/golang/glog"
)

// Parser defines the struct for processing eligibility criteria.
type Parser struct {
	parameters conf.Config
	registry   studies.Studies
	clock      timer.Timer
}

// NewParser creates a new parser to parse eligibility criteria.
func NewParser() *Parser {
	return &Parser{clock: timer.New()}
}

// Main function to parse eligibility criteria
func CfgParse(registry, configFname string) (string, error) {
	p := NewParser()
	if err := p.LoadParameters(configFname); err != nil {
		return "", fmt.Errorf("cfg parser failed to load config file: %v", err)
	}
	if err := p.InitParameters(); err != nil {
		return "", fmt.Errorf("cfg parser failed to initialize variables and inputs: %v", err)
	}
	if err := p.UnmarshalInput(registry); err != nil {
		return "", fmt.Errorf("cfg parser failed to unmarshal the json input: %v", err)
	}
	result := p.Parse()
	p.Close()

	return result, nil
}

// LoadParameters loads variables and units from command line and a config file.
func (p *Parser) LoadParameters(configFname string) error {
	log.Printf("config file path: %s", configFname)
	if len(configFname) == 0 {
		return fmt.Errorf("no configuration file found: %s", configFname)
	}

	parameters, err := conf.Load(configFname)
	if err != nil {
		return err
	}
	p.parameters = parameters
	return nil
}

// InitParameters initializes the parser by loading the resource data.
func (p *Parser) InitParameters() error {
	fname := p.parameters.GetResourcePath("variable_file")
	log.Printf("variable file path: %v", fname)
	variableDictionary, err := variables.Load(fname)
	if err != nil {
		return err
	}
	variables.Set(variableDictionary)

	fname = p.parameters.GetResourcePath("unit_file")
	log.Printf("unit file path: %v", fname)
	unitDictionary, err := units.Load(fname)
	if err != nil {
		return err
	}
	units.Set(unitDictionary)

	return nil
}

// UnmarshalInput ingests eligibility criteria from json string input.
func (p *Parser) UnmarshalInput(data string) error {
	var ss []studies.Study
	err := json.Unmarshal([]byte(data), &ss)
	if err != nil {
		return err
	}

	// fmt.Printf("cfg got Studies : %+v", ss)

	registry := studies.New()
	for _, study := range ss {
		registry.Add(studies.NewStudy(study.Id, study.Name, study.Conditions, study.EligibilityCriteria))
	}
	p.registry = registry

	return nil
}

// Parse parses the ingested eligibility criteria and writes the results to a file.
func (p *Parser) Parse() string {
	relationCnt := 0
	criteriaCnt := 0
	parsedCriteriaCnt := 0

	var ps studies.ParsedStudies

	for _, study := range p.registry {
		r := study.Parse().Relations()
		s := studies.NewParsedStudy(study.Id, study.CriteriaCnt, r)
		ps = append(ps, s)

		relationCnt += study.RelationCount()
		criteriaCnt += study.CriteriaCount()
		parsedCriteriaCnt += study.ParsedCriteriaCount()
	}

	ratio := 0.0
	if criteriaCnt > 0 {
		ratio = 100 * float64(relationCnt) / float64(criteriaCnt)
	}

	glog.Infof("Ingested studies: %d, Extracted criteria: %d, Parsed criteria: %d, Relations: %d, Relations per criteria: %.1f%%\n",
		p.registry.Len(), criteriaCnt, parsedCriteriaCnt, relationCnt, ratio)

	return ps.JSON()
}

// Close closes the parser.
func (p *Parser) Close() {
	glog.Info(p.clock.Elapsed())
	glog.Flush()
}
