// Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved.

package parser

import (
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
	for _, listVal := range list {
		for _, item := range listVal {
			for _, orR := range orRs {
				if item.val == orR.Name {
					orR.Start = int(item.pos)
					orR.End = int(item.pos) + len(item.val)
				}
			}
			for _, andR := range andRs {
				if item.val == andR.Name {
					andR.Start = int(item.pos)
					andR.End = int(item.pos) + len(item.val)
				}
			}
		}
	}
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
