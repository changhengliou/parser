package recursive_descent

import (
	"fmt"
	"strings"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

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
type symbolMap map[string]symbolGroup

func (s *symbol) str() string {
	return string(*s)
}

func (g *symbolGroup) add(sym symbol) {
	*g = append(*g, sym)
}

type parser struct {
	input   []byte
	symMap  symbolMap
	currPos int
}

func (p *parser) initSymbols(str string) {
	strMap := strings.Split(str, "->")
	if len(strMap) != 2 {
		panic("Failed to parse symbol key")
	}

	strGrp := strings.Split(strMap[1], "|")
	var symGrp symbolGroup
	for _, s := range strGrp {
		symGrp.add([]byte(s))
	}
	p.symMap[strMap[0]] = symGrp
}

func (p *parser) leftFactor() {

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
