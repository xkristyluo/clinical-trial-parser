// Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved.

package parser

import (
	"strings"

	"github.com/facebookresearch/clinical-trial-parser/src/ct/units"
	"github.com/facebookresearch/clinical-trial-parser/src/ct/variables"

	"github.com/golang/glog"
)

// Parser defines the parser logic for parsing clinical trial eligibility criteria.
type Parser struct {
	lexer  *Lexer
	tokens []*Token // lookahead for parser.
}

// NewParser creates a new parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses the input string to the list of criterion items.
func (p *Parser) Parse(input string) (criteria List) {
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("%v: %q\n", r, input)
			criteria = NewList()
		}
	}()
	p.lexer = NewLexer(input)
	p.tokens = make([]*Token, 0)
	criteria = p.parseSegment(tokenEOF)
	criteria.TrimItems()
	return
}

// next returns the next token.
func (p *Parser) next() *Token {
	if len(p.tokens) > 0 {
		t := p.tokens[0]
		p.tokens = p.tokens[1:]
		return t
	}
	return p.lexer.NextToken()
}

// peek returns but does not consume the next token.
func (p *Parser) peek(cnt int) *Token {
	for len(p.tokens) < cnt {
		p.tokens = append(p.tokens, p.lexer.NextToken())
	}
	return p.tokens[cnt-1]
}

// Parser methods:

func (p *Parser) parseSegmentWithoutR(tokenEnd tokenType) List {
	list := make(List, 0)
	nodes := NewItems()

loop:
	for {
		switch p.peek(1).typ {
		case tokenLeftParenthesis:
			p.next()
			if tokenEnd == tokenRightParenthesis {
				break loop
			}
			newNodes := p.parseSegment(tokenRightParenthesis)
			for i := 0; i < len(newNodes); i++ {
				if newNodes[i].Len() == 1 {
					// Skip one-token segment which most likely is an abbreviation.
				} else {
					list = append(list, newNodes[i])
				}
			}
		case tokenIdentifier:
			n := p.parseIdentifier()
			nodes.Add(n)
		case tokenNumber:
			if n := p.parseNumber(); n.Valid() {
				nodes.Add(n)
			}
		case tokenUnit:
			if n := p.parseUnit(); n.Valid() {
				nodes.Add(n)
			}
		case tokenNegation, tokenComparison, tokenLessComparison, tokenGreaterComparison:
			if n := p.parseComparison(); n.Valid() {
				nodes.Add(n)
			}
		case tokenConjunction:
			if n := p.parseConjunction(); n.Valid() {
				nodes.Add(n)
			}
		case tokenSlash:
			if nodes.LastType() == itemNumber {
				// Because a number preceded the slash, these tokens
				// may compose to a unit, such as '/ul'.
				n := p.parseIdentifier()
				nodes.Add(n)
			} else {
				if n := p.parseSlash(); n.Valid() {
					nodes.Add(n)
				}
			}
		case tokenDash:
			if n := p.parseDash(); n.Valid() {
				nodes.Add(n)
			}
		case tokenPunctuation:
			if n := p.parsePunctuation(); n.Valid() {
				nodes.Add(n)
			}
		case tokenEOF:
			break loop
		case tokenEnd:
			p.next()
			break loop
		default:
			p.next()
		}
	}
	if !nodes.Empty() {
		list = append(list, nodes)
	}
	return list
}

// parseSegment parses a segment of tokens from the beginning to the end of the input string
// or delimited by a pair of left and right parenthesises.
func (p *Parser) parseSegment(tokenEnd tokenType) List {
	list := make(List, 0)
	nodes := NewItems()

loop:
	for {
		switch p.peek(1).typ {
		case tokenLeftParenthesis:
			p.next()
			if tokenEnd == tokenRightParenthesis {
				break loop
			}
			newNodes := p.parseSegmentWithoutR(tokenRightParenthesis)
			for i := 0; i < len(newNodes); i++ {
				if newNodes[i].Len() == 1 {
					// Skip one-token segment which most likely is an abbreviation.
				} else {
					list = append(list, newNodes[i])
				}
			}
		case tokenRightParenthesis:
			p.next()
			newNodes := p.parseSegment(tokenEnd)
			for i := 0; i < len(newNodes); i++ {
				if newNodes[i].Len() == 1 {
					// Skip one-token segment which most likely is an abbreviation.
				} else {
					list = append(list, newNodes[i])
				}
			}
		case tokenIdentifier:
			n := p.parseIdentifier()
			nodes.Add(n)
		case tokenNumber:
			if n := p.parseNumber(); n.Valid() {
				nodes.Add(n)
			}
		case tokenUnit:
			if n := p.parseUnit(); n.Valid() {
				nodes.Add(n)
			}
		case tokenNegation, tokenComparison, tokenLessComparison, tokenGreaterComparison:
			if n := p.parseComparison(); n.Valid() {
				nodes.Add(n)
			}
		case tokenConjunction:
			if n := p.parseConjunction(); n.Valid() {
				nodes.Add(n)
			}
		case tokenSlash:
			if nodes.LastType() == itemNumber {
				// Because a number preceded the slash, these tokens
				// may compose to a unit, such as '/ul'.
				n := p.parseIdentifier()
				nodes.Add(n)
			} else {
				if n := p.parseSlash(); n.Valid() {
					nodes.Add(n)
				}
			}
		case tokenDash:
			if n := p.parseDash(); n.Valid() {
				nodes.Add(n)
			}
		case tokenPunctuation:
			if n := p.parsePunctuation(); n.Valid() {
				nodes.Add(n)
			}
		case tokenEOF:
			break loop
		case tokenEnd:
			p.next()
			break loop
		default:
			p.next()
		}
	}
	if !nodes.Empty() {
		list = append(list, nodes)
	}
	return list
}

func (p *Parser) parseUnit() *Item {
	// 	n := UnknownItem()
	// 	t := p.next()
	// 	n.pos = t.pos
	// 	n.name = t.val
	//
	// 	if t.typ == tokenUnit {
	// 		if p.peek(1).typ == tokenUnit {
	// 			n.name = n.name +" "+ p.peek(1).val
	// 			p.next()
	// 			n.Set(itemUnit, n.name)
	// 		} else {
	// 			n.Set(itemUnit, n.name)
	// 		}
	// 	}
	// 		return n
	if t := p.next(); t.typ == tokenUnit {
		n := NewItem(itemUnit, t.val)
		n.pos = t.pos
		n.name = t.val
		return n
	}
	return UnknownItem()
}

func (p *Parser) parseNumber() *Item {
	if t := p.next(); t.typ == tokenNumber {
		n := UnknownItem()
		if p.peek(1).typ == tokenSlash && p.peek(2).typ == tokenNumber {
			n.Set(itemNumber, t.val+"/"+p.peek(2).val)
			p.next()
			p.next()
		} else {
			n.Set(itemNumber, t.val)
		}
		n.pos = t.pos
		n.name = t.val
		return n
	}
	return UnknownItem()
}

func (p *Parser) parseIdentifier() *Item {
	n := UnknownItem()
	t := p.next()
	n.pos = t.pos

	if t.val == "to" {
		n.Set(itemRange, t.val)
		return n
	}

	variable := ""
	unit := ""
	candidate := t.val
	var candidateBackup string

	variableMatchCnt := 0
	unitMatchCnt := 0
	identifierCnt := 0

	variableCatalog := variables.Get()
	unitCatalog := units.Get()

	isIdentifier := true

loop:
	for isIdentifier {
		identifierCnt++
		switch {
		case variableCatalog.Match(candidate):
			if name, ok := variableCatalog.Get(candidate); ok {
				variable = name
				variableMatchCnt = identifierCnt
			}
			fallthrough
		case unitCatalog.Match(candidate):
			if name, ok := unitCatalog.Get(candidate); ok {
				unit = name
				unitMatchCnt = identifierCnt
			}
		default:
			break loop
		}

		t = p.peek(identifierCnt)

		if t.typ == tokenLeftParenthesis {
			for {
				t = p.peek(identifierCnt)
				identifierCnt++
				if t.typ == tokenRightParenthesis || t.typ == tokenEOF {
					break
				}
			}
			if t.typ == tokenRightParenthesis {
				t = p.peek(identifierCnt)
			}
		}
		candidateBackup = candidate
		candidate += " " + t.val
		isIdentifier = t.typ == tokenIdentifier || t.typ == tokenConjunction || t.typ == tokenSlash
	}

	switch {
	case variableMatchCnt == 0 && unitMatchCnt == 0:
		n.name = t.val

		// swallow
	case variableMatchCnt < unitMatchCnt:
		n.Set(itemUnit, unit)
		n.name = candidateBackup
		for i := 1; i < unitMatchCnt; i++ {
			p.next()
		}
	default:
		n.Set(itemVariable, variable)
		n.name = candidateBackup
		for i := 1; i < variableMatchCnt; i++ {
			p.next()
		}
	}

	return n
}

func (p *Parser) parseComparison() *Item {
	n := UnknownItem()
	t := p.next()
	n.pos = t.pos
	n.name = t.val

	negate := false
	if t.typ == tokenNegation && p.peek(1).typ != tokenEOF {
		negate = true
		t = p.next()
	}

	switch {
	case containsStrings(t.val, "<", "=") || containsStrings(t.val, "≤"):
		n.Set(itemComparison, "≤")
	case containsStrings(t.val, "<"):
		if p.peek(1).val == "or" {
			p.next()
			if p.peek(1).val == "=" {
				p.next()
				n.Set(itemComparison, "≤")
			}
		} else {
			n.Set(itemComparison, "<")
		}
	case containsStrings(t.val, ">", "=") || containsStrings(t.val, "≥"):
		n.Set(itemComparison, "≥")
	case containsStrings(t.val, ">"):
		if p.peek(1).val == "or" {
			p.next()
			n.name = n.name + " or"
			if p.peek(1).val == "=" {
				p.next()
				n.name = n.name + " ="
				n.Set(itemComparison, "≥")
			}
		} else {
			n.Set(itemComparison, ">")
		}
	case t.typ == tokenLessComparison:
		switch {
		case p.hasEqual():
			n.Set(itemComparison, "≤")
		default:
			if p.peek(1).val == "the" {
				p.next()
				if p.peek(1).val == "last" || p.peek(1).val == "next" || p.peek(1).val == "past" || p.peek(1).val == "first" {
					p.next()
					n.Set(itemComparison, "≤")
				} else {
					n.Set(itemComparison, "≤")
				}
			} else {
				n.Set(itemComparison, "<")
			}
		}
	case t.typ == tokenGreaterComparison:
		switch {
		case p.hasEqual():
			n.Set(itemComparison, "≥")
		default:
			n.Set(itemComparison, ">")
		}
	case t.typ == tokenComparison:
		switch t.val {
		case "between":
			n.Set(itemRange, t.val)
		case "at":
			if p.peek(1).val == "least" {
				p.next()
				n.Set(itemComparison, "≥")
			}
		case "within":
			if p.peek(1).val == "the" {
				p.next()
				if p.peek(1).val == "last" {
					p.next()
					n.Set(itemComparison, "≤")
				} else {
					n.Set(itemComparison, "≤")
				}
			}
		}
	}
	if negate {
		n = n.Negate()
	}
	return n
}

func (p *Parser) parseConjunction() *Item {
	n := UnknownItem()
	t := p.next()
	n.pos = t.pos
	n.name = t.val
	if t.typ == tokenConjunction {
		switch t.val {
		case "or", "and/or":
			switch p.peek(1).typ {
			case tokenLessComparison:
				p.hasEqual()
				n.Set(itemComparison, "≤")
			case tokenGreaterComparison:
				p.hasEqual()
				n.Set(itemComparison, "≥")
			default:
				n.Set(itemOr, "or")
			}
		default:
			n.Set(itemAnd, "and")
		}
	}
	return n
}

func (p *Parser) parsePunctuation() *Item {
	if t := p.next(); t.typ == tokenPunctuation {
		n := NewItem(itemPunctuation, t.val)
		n.pos = t.pos
		n.name = t.val
		return n
	}
	return UnknownItem()
}

func (p *Parser) parseSlash() *Item {
	if t := p.next(); t.typ == tokenSlash {
		n := NewItem(itemSlash, t.val)
		n.pos = t.pos
		n.name = t.val
		return n
	}
	return UnknownItem()
}

func (p *Parser) parseDash() *Item {
	if t := p.next(); t.typ == tokenDash {
		n := NewItem(itemRange, t.val)
		n.pos = t.pos
		n.name = t.val
		return n
	}
	return UnknownItem()
}

func (p *Parser) hasEqual() bool {
	equal := false
	if p.peek(1).val == "than" {
		p.next()
	}
	if p.peek(1).val == "or" {
		p.next()
	}
	if p.peek(1).val == "equal" {
		p.next()
		equal = true
		if p.peek(1).val == "to" {
			p.next()
		}
	}
	return equal
}

func containsStrings(s string, subs ...string) bool {
	for _, si := range subs {
		if !strings.Contains(s, si) {
			return false
		}
	}
	return true
}
