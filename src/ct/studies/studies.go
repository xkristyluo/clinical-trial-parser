// Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved.

package studies

import "encoding/json"

// Studies defines a collection of clinical study records.
type Studies []*Study

type ParsedStudies []*ParsedStudy

func New() Studies {
	return make(Studies, 0)
}

func (s *ParsedStudies) JSON() string {
	if data, err := json.Marshal(s); err == nil {
		return string(data)
	}
	return ""
}

// Add adds the study to the studies.
func (ss *Studies) Add(s *Study) {
	*ss = append(*ss, s)
}

// Len returns the number of studies.
func (ss Studies) Len() int {
	return len(ss)
}

// Parse parses eligibility criteria text to relations for the studies ss.
func (ss Studies) Parse() {
	for _, s := range ss {
		s.Parse()
	}
}
