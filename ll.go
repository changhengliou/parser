package main

import (
	"fmt"
	"parser/common"
	"strings"
)

var IS_TERM = func(s string) bool {
	if len(s) > 1 {
		return true
	}
	if s[0] >= 'A' && s[0] <= 'Z' {
		return false
	}
	return true
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

// R -> A + B C | A + B X | C X | C
// R -> A + B W | C Z
// W -> C | X
// Z -> X | ε
// recursive descent parser -> predictive parser LL(1) -> bottom up parser
// E -> T | T + E
// T -> int | int * T | (E)
// input = (int)
// -------------------------
// LL(1) parser
// left-factor: eliminate multiple valid choices
// E -> T X
// X -> + E | ε
// T -> int Y | (E)
// Y -> * T | ε
// -------------------------
// LL(1) is too weak, below are all not LL(1)
// parser table
// NOT left-factored
// left recursive
// ambiguous
type symbol string
type term symbol
type nonTerm symbol
type symbolGroup []interface{}
type symbolMap map[string][]symbolGroup

func (s *symbol) str() string {
	return string(*s)
}

func (s *term) str() string {
	return string(*s)
}
func (s *nonTerm) str() string {
	return string(*s)
}

func (g *symbolGroup) add(sym interface{}) {
	*g = append(*g, sym)
}

func (m symbolMap) add(key string, val symbolGroup) {
	m[key] = append(m[key], val)
}

type parser struct {
	input     []byte
	symMap    symbolMap
	nonTerms  []string
	firstSet  map[string]common.Set
	followSet map[string]common.Set
	currPos   int
}

func New() *parser {
	p := &parser{}
	p.symMap = make(map[string][]symbolGroup)
	p.firstSet = make(map[string]common.Set)
	p.followSet = make(map[string]common.Set)
	return p
}

func (p *parser) initSymbols(str string, isTerm func(s string) bool) {
	if isTerm == nil {
		isTerm = IS_TERM
	}
	strMap := strings.Split(str, "->")
	if len(strMap) != 2 {
		panic("Failed to parse symbol key")
	}

	key := strings.Trim(strMap[0], " ")
	p.nonTerms = append(p.nonTerms, key)

	strGrps := strings.Split(strMap[1], "|")
	for _, grp := range strGrps {
		var symGrp symbolGroup
		strSyms := strings.Split(grp, " ")
		for _, sym := range strSyms {
			if len(strings.Trim(sym, " ")) == 0 {
				continue
			}
			if isTerm(sym) {
				symGrp.add(term(sym))
			} else {
				symGrp.add(nonTerm(sym))
			}
		}
		p.symMap.add(key, symGrp)
	}
}

// R -> A + B C | A + B X | C X | C
// R -> A + B W | C Z
// W -> C | X
// Z -> X | ε
// -----------------
// A → α β | α γ
//
// A → α A'
// A' → β | γ
func (p *parser) leftFactor() {
	//for key, symGrps := range p.symMap {
	// find common prefix
	// merge string with common prefix
	//}
}

type stack []byte

func (s *stack) push(b byte) {
	*s = append(*s, b)
}
func (s *stack) pop() byte {
	c := s.top()
	*s = (*s)[:len(*s)-1]
	return c
}
func (s stack) top() byte {
	return s[len(s)-1]
}
func (s stack) empty() bool {
	return len(s) == 0
}

func (p *parser) buildFirstSet() {
	p.buildSet(p.firstOf)
}

func (p *parser) buildSet(builder func(interface{}) common.Set) {
	for _, nonTerm := range p.nonTerms {
		builder(symbol(nonTerm))
	}
}

func (p *parser) firstOf(sym interface{}) common.Set {
	symStr := fmt.Sprintf("%v", sym)
	// already in the set
	if val, exists := p.firstSet[symStr]; exists {
		return val
	}
	for _, val := range p.symMap[symStr] {
		currSet := p.firstSet[symStr]
		if currSet == nil {
			currSet = common.Set{}
		}
		switch v := val[0].(type) {
		case term:
			currSet.Add(v.str())
		case nonTerm:
			currSet.AddSet(p.firstOf(v))
		}
		p.firstSet[symStr] = currSet
	}
	return p.firstSet[symStr]
}

// S -> X T -> A B T
// X -> A B
// follow(A) = first(B)
// follow(X) = follow(B)
// if B -> ε, follow(X) = follow(A)
// if S is start symbol, $ -> follow(S)
func (p *parser) buildFollowSet() {
	p.buildSet(p.followOf)
}

// E -> T X
// T -> ( E ) | int Y
// X -> + E | ε
// Y -> * T | ε
// follow(E) = follow(X)
// ------------------------------
// follow(E) = { $, ), follow(X) } = { $, ) }
// follow(X) = { follow(E) } = { $, ) }
// follow(T) = { follow(Y), first(X) } = { follow(Y), +, follow(X) } = { follow(Y), +, $, ) } = { +, $, ) }
// follow(Y) = { follow(T) } = { +, $, ) }
// follow(() = { first(E) } = { (, int }
// follow()) = { follow(T) } = { +, $, ) }
// follow(+) = { first(E) } = { (, int }
// follow(*) = { first(T) } = { (, int }
// follow(int) = { first(Y) } = { *, follow(Y) } = { *, +, $, ) }
func (p *parser) followOf(sym interface{}) common.Set {
	symStr := fmt.Sprintf("%v", sym)
	if val, exists := p.followSet[symStr]; exists {
		return val
	}

	if p.followSet[symStr] == nil {
		p.followSet[symStr] = common.Set{}
	}
	if symStr == p.nonTerms[0] {
		p.followSet[symStr].Add("$")
	}
	occurrences := p.findOccurrences(symStr)
	for _, val := range occurrences {
		if val.groupIndex == len(p.symMap[val.key][val.groupsIndex])-1 {
			p.followSet[symStr].AddSet(p.followOf(val.key))
		} else {
			currSym := p.symMap[val.key][val.groupsIndex][val.groupIndex+1]
			currSymStr := fmt.Sprintf("%v", currSym)
			switch currSym.(type) {
			case term:
				p.followSet[symStr].Add(currSymStr)
			case nonTerm:
				firstOfSet := p.firstOf(currSymStr)
				if firstOfSet.Exist("ε") {
					p.followSet[symStr].AddSet(p.followOf(currSymStr))
				}
				for _, val := range firstOfSet.Keys() {
					if val != "ε" {
						p.followSet[symStr].Add(val)
					}
				}
			}
		}
	}
	return nil
}

type symPosition struct {
	key         string
	groupsIndex int
	groupIndex  int
}

func (p *parser) findOccurrences(str string) []*symPosition {
	arr := make([]*symPosition, 0)
	for _, nonTerm := range p.nonTerms {
		for groupsIndex, symGroup := range p.symMap[nonTerm] {
			for groupIndex, sym := range symGroup {
				symStr := fmt.Sprintf("%v", sym)
				if symStr == str {
					arr = append(arr, &symPosition{
						nonTerm, groupsIndex, groupIndex,
					})
				}
			}
		}
	}
	return arr
}

func parse(input string) bool {
	symbolStack := stack{}
	symbolStack.push('E')

	currPos := 0
	for !symbolStack.empty() {
		char := symbolStack.pop()
		// non terminal
		if char >= 'a' && char <= 'z' {
			if char != input[currPos] {
				return false
			}
		} else {
			// lookup table
			// push result from table to stack
		}
		currPos++
	}
	return true
}

// S->Sa|b
// S->bS'
// S'->aS'| ε
// https://en.wikipedia.org/wiki/Left_recursion#Removing_all_left_recursion
func main() {
	p := New()
	//p.initSymbols("E -> T X", nil)
	//p.initSymbols("X -> + T X | ε", nil)
	//p.initSymbols("T -> F Y", nil)
	//p.initSymbols("Y -> * F Y | ε", nil)
	//p.initSymbols("F -> a | ( E )", nil)
	//p.buildFirstSet()
	//fmt.Printf("E => %v\n", p.firstSet["E"].Keys())
	//fmt.Printf("T => %v\n", p.firstSet["T"].Keys())
	//fmt.Printf("F => %v\n", p.firstSet["F"].Keys())
	//fmt.Printf("X => %v\n", p.firstSet["X"].Keys())
	//fmt.Printf("T => %v\n", p.firstSet["Y"].Keys())
	p.initSymbols("E -> T X", nil)
	p.initSymbols("T -> ( E ) | int Y", nil)
	p.initSymbols("X -> + E | ε", nil)
	p.initSymbols("Y -> * T E & | ε", nil)
	p.buildFollowSet()
	// E -> T X
	// T -> ( E ) | int Y
	// X -> + E | ε
	// Y -> * T | ε

	fmt.Printf("E => %v\n", p.followSet["E"].Keys())
	fmt.Printf("X => %v\n", p.followSet["X"].Keys())
	fmt.Printf("T => %v\n", p.followSet["T"].Keys())
	fmt.Printf("Y => %v\n", p.followSet["Y"].Keys())
}
