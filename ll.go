package main

import (
	"fmt"
	"parser/trie"
	"strings"
)

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
type symbol []byte
type term symbol
type nonTerm symbol
type symbolGroup []symbol
type symbolMap map[string][]symbolGroup

func (s *symbol) str() string {
	return string(*s)
}

func (g *symbolGroup) add(sym symbol) {
	*g = append(*g, sym)
}

func (m symbolMap) add(key string, val symbolGroup) {
	m[key] = append(m[key], val)
}

type parser struct {
	input   []byte
	symMap  symbolMap
	currPos int
}

func New() *parser {
	p := &parser{}
	p.symMap = make(map[string][]symbolGroup)
	return p
}

func (p *parser) initSymbols(str string) {
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
			symGrp.add([]byte(sym))
		}
		p.symMap.add(key, symGrp)
	}
}

// R -> A + B C | A + B X | C X | C
// R -> A + B W | C Z
// W -> C | X
// Z -> X | ε
func (p *parser) leftFactor() {
	for key, symGrps := range p.symMap {
		// find common prefix
		// merge string with common prefix
	}
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

func firstSet() {

}

// S -> X T -> A B T
// X -> A B
// follow(A) = first(B)
// follow(X) = follow(B)
// if B -> ε, follow(X) = follow(A)
// if S is start symbol, $ -> follow(S)
func followSet() {

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
	var s symbolGroup
	s.add([]byte("abc"))
	s.add([]byte("def"))

	fmt.Println(s)
}

