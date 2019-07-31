package main

import (
	"fmt"
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

type set map[string]interface{}

func (s set) add(str string) {
	s[str] = nil
}
func (s set) exist(str string) bool {
	if _, exists := s[str]; exists {
		return true
	}
	return false
}
func (s set) keys() []string {
	i := 0
	arr := make([]string, len(s))
	for k := range s {
		arr[i] = k
		i++
	}
	return arr
}
func (s set) addSet(set2 set) {
	for _, v := range set2.keys() {
		s.add(v)
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
	input       []byte
	symMap      symbolMap
	startSymbol string
	firstSet    map[string]set
	followSet   map[string]set
	currPos     int
}

func New() *parser {
	p := &parser{}
	p.symMap = make(map[string][]symbolGroup)
	p.firstSet = make(map[string]set)
	p.followSet = make(map[string]set)
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

func (p *parser) buildSet(builder func(interface{}) set) {
	for key := range p.symMap {
		builder(symbol(key))
	}
}

func (p *parser) firstOf(sym interface{}) set {
	symStr := fmt.Sprintf("%v", sym)
	// already in the set
	if val, exists := p.firstSet[symStr]; exists {
		return val
	}
	for _, val := range p.symMap[symStr] {
		currSet := p.firstSet[symStr]
		if currSet == nil {
			currSet = set{}
		}
		switch v := val[0].(type) {
		case term:
			currSet.add(v.str())
		case nonTerm:
			currSet.addSet(p.firstOf(v))
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

func (p *parser) followOf(sym interface{}) set {
	symStr := fmt.Sprintf("%v", sym)
	if val, exists := p.followSet[symStr]; exists {
		return val
	}

	if symStr == p.startSymbol {
		p.followSet[symStr].add("$")
	}
	occurrences := findInMap(p.symMap, symStr)
	for _, val := range occurrences {
		// E -> T X
		// T -> ( E ) | int Y
		// X -> + E | ε
		// Y -> * T | ε
		// follow(E) = follow(X)
		if val.groupIndex == len(p.symMap[val.key][val.groupsIndex]) - 1 {
			p.followSet[symStr].addSet(p.followOf(val.key))
		} else {
			currSym := p.symMap[val.key][val.groupsIndex][val.groupIndex+1]
			currSymStr := fmt.Sprintf("%v", currSym)
			switch currSym.(type) {
			case term:
				p.followSet[symStr].add(currSymStr)
			case nonTerm:
				firstOfSet := p.firstOf(currSymStr)
				if firstOfSet.exist("ε") {
					p.followSet[symStr].addSet(p.followOf(currSym))
				}
				p.followSet[symStr].addSet(firstOfSet)
			}
		}
	}
	return nil
}

type symPosition struct {
	key         string
	groupsIndex int
	groupIndex  int
	val         string
}

func findInMap(symMap symbolMap, str string) []*symPosition {
	arr := make([]*symPosition, 0)
	for key, symGroups := range symMap {
		for groupsIndex, symGroup := range symGroups {
			for groupIndex, sym := range symGroup {
				symStr := fmt.Sprintf("%v", sym)
				if symStr == str {
					arr = append(arr, &symPosition{
						key, groupsIndex, groupIndex, str,
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
	p.startSymbol = "E"
	p.initSymbols("E -> T X", nil)
	p.initSymbols("X -> + T X | ε", nil)
	p.initSymbols("T -> F Y", nil)
	p.initSymbols("Y -> * F Y | ε", nil)
	p.initSymbols("F -> a | ( E )", nil)

	p.buildFirstSet()
	fmt.Println(p.firstSet["E"].keys())
	fmt.Println(p.firstSet["T"].keys())
	fmt.Println(p.firstSet["F"].keys())
	fmt.Println(p.firstSet["X"].keys())
	fmt.Println(p.firstSet["Y"].keys())
}
