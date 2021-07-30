package main

import (
	"github.com/facebookresearch/clinical-trial-parser/src/cmd/cfg"
	"github.com/golang/glog"
)

func main() {
	input := `[{
		"study_id": "NCT04342793",
		"study_name": "A Study to Evaluate the Efficacy and Safety of ALS-L1023 in Subjects With NASH",
		"conditions": ["Nonalcoholic", "Steatohepatitis"], 
		"eligibility_criteria": "Inclusion Criteria:\n\n- Men or women ages 19 and over, under 75 years of age\n\n- Patients diagnosed with NAFLD on abdominal ultrasonography and MRI\n\n- Patients show presence of hepatic fat fraction as defined by ≥ 8% on MRI-PDFF and\nliver stiffness as defined by ≥ 2.5 kPa on MRE at Screening\n\nExclusion Criteria:\n\n- Any subject with current, significant alcohol consumption or a history of significant\nalcohol consumption for a period of more than 3 consecutive months any time within 2\nyear prior to screening will be excluded\n\n- Chronic liver disease (including hemochromatosis, liver cancer, autoimmune liver\ndisease, viral hepatitis A, B, alcoholic liver disease\n\n- Uncontrolled diabetes mellitus as defined by a HbA1c ≥ 9.0％ at Screening\n\n- Patients who are allergic or hypersensitive to the drug or its constituents\n\n- Pregnant or lactating women"
		}]`

	p := cfg.NewParser()
	if err := p.LoadParameters(); err != nil {
		glog.Fatal(err)
	}
	if err := p.Initialize(); err != nil {
		glog.Fatal(err)
	}
	if err := p.Ingest(input); err != nil {
		glog.Fatal(err)
	}

	output := p.Parse()
	glog.Infof("parsed eligibility criteria: \n%s", output)
	p.Close()
}
