// Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved.

package parser

import (
	"strings"

	"github.com/facebookresearch/clinical-trial-parser/src/ct/parser/production"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/relation"
)

var interpreter *Interpreter

func init() {
	interpreter = NewInterpreter()
}

// Get gets the interpreter to parse strings to relations.
func Get() *Interpreter {
	return interpreter
}

// Interpreter defines the interpreter struct to convert
// unstructured criteria strings to structured relations.
type Interpreter struct {
	parser  *Parser
	grammar Grammar
}

// NewInterpreter creates a new interpreter.
func NewInterpreter() *Interpreter {
	return &Interpreter{parser: NewParser(), grammar: NewCFGrammar(production.CriterionRules)}
}

// Interpret interprets clinical trial criteria using parse trees and formal grammars.
func (i *Interpreter) Interpret(input string) (relation.Relations, relation.Relations) {
	list := i.parser.Parse(input)
	list.FixMissingVariable()
	trees := i.buildTrees(list)
	orRs, andRs := trees.Relations()
	// 	m := make(map[string]int)
	// 	for _, listVal := range list {
	// 		for _, item := range listVal {
	// 			for _, orR := range orRs {
	// 				if item.val == orR.Name {
	// 					orR.Start = UnicodeIndex(input, item.val, m)
	// 					orR.End = orR.Start + len(item.val)
	// 				}
	// 				if orR.Lower != nil && orR.Lower.Value == item.val {
	// 					if orR.Lower.Start == 0 {
	// 						orR.Lower.Start = UnicodeIndex(input, item.val, m)
	// 						orR.Lower.End = orR.Lower.Start + len(item.val)
	// 					}
	// 				}
	// 				if orR.Upper != nil && orR.Upper.Value == item.val {
	// 					if orR.Upper.Start == 0 {
	// 						orR.Upper.Start = UnicodeIndex(input, item.val, m)
	// 						orR.Upper.End = orR.Upper.Start + len(item.val)
	// 					}
	// 				}
	// 				if orR.Unit != nil && orR.Unit.Value == item.val && item.val != "" {
	// 					startIndex := UnicodeIndex(input, item.val, m)
	// 					orR.Unit.Start = append(orR.Unit.Start, startIndex)
	// 					orR.Unit.End = append(orR.Unit.End, startIndex + len(item.val))
	// 				}
	// 			}
	// 			for _, andR := range andRs {
	// 				if item.val == andR.Name {
	// 					andR.Start = UnicodeIndex(input, item.val, m)
	// 					andR.End = andR.Start + len(item.val)
	// 				}
	// 				if andR.Lower != nil && andR.Lower.Value == item.val {
	// 					if andR.Lower.Start == 0 {
	// 						andR.Lower.Start = UnicodeIndex(input, item.val, m)
	// 						andR.Lower.End = andR.Lower.Start + len(item.val)
	// 					}
	// 				}
	// 				if andR.Upper != nil && andR.Upper.Value == item.val {
	// 					if andR.Upper.Start == 0 {
	// 						andR.Upper.Start = UnicodeIndex(input, item.val, m)
	// 						andR.Upper.End = andR.Upper.Start + len(item.val)
	// 					}
	// 				}
	//
	// 				if andR.Unit != nil && andR.Unit.Value == item.val && item.val != "" {
	// 					startIndex := UnicodeIndex(input, item.val, m)
	// 					andR.Unit.Start = append(andR.Unit.Start, startIndex)
	// 					andR.Unit.End = append(andR.Unit.End, startIndex + len(item.val))
	// 				}
	// 			}
	// 		}
	// 	}
	return orRs, andRs
}

// buildTrees builds trees from the parsed items. Trees represent criteria.
func (i *Interpreter) buildTrees(list List) Trees {
	trees := NewTrees()
	for _, items := range list {
		ts := i.grammar.BuildTrees(items)
		trees = append(trees, ts...)
	}
	trees.Dedupe()
	return trees
}

//get unicode index of substring
func UnicodeIndex(str, substr string, m map[string]int) int {
	lastIndex := m[substr]
	if lastIndex != 0 {
		lastIndex += len(substr)
	}
	str = str[lastIndex:]
	result := strings.Index(str, substr)
	if result >= 0 {
		prefix := []byte(str)[0:result]
		rs := []rune(string(prefix))
		result = len(rs) + lastIndex
		m[substr] = result
	}
	return result
}
