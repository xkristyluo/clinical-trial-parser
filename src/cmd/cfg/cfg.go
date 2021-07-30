package cfg

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

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

// LoadParameters loads variables and units from command line and a config file.
func (p *Parser) LoadParameters() error {
	configFname := flag.String("conf", "", "Config file")

	flag.Parse()
	if len(*configFname) == 0 {
		return fmt.Errorf("usage: %s -conf <config file>", os.Args[0])
	}

	parameters, err := conf.Load(*configFname)
	if err != nil {
		return err
	}
	p.parameters = parameters
	return nil
}

// Initialize initializes the parser by loading the resource data.
func (p *Parser) Initialize() error {
	fname := p.parameters.GetResourcePath("variable_file")
	variableDictionary, err := variables.Load(fname)
	if err != nil {
		return err
	}
	variables.Set(variableDictionary)

	fname = p.parameters.GetResourcePath("unit_file")
	unitDictionary, err := units.Load(fname)
	if err != nil {
		return err
	}
	units.Set(unitDictionary)

	return nil
}

// Ingest ingests eligibility criteria from json string input.
func (p *Parser) Ingest(data string) error {
	var ss []studies.Study
	err := json.Unmarshal([]byte(data), &ss)
	if err != nil {
		return err
	}

	// fmt.Printf("Studies : %+v", ss)

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
