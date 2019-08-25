package main

import (
	"fmt"
	"strconv"
)

// *+?()|
// \+ for escape
// e1 matches s and e2 matches t, then e1|e2 matches s or t, and e1e2 matches st.
// The metacharacters *, +, and ? are repetition operators:
// e1* matches a sequence of zero or more (possibly different) strings, each of which match e1; e1+ matches one or more;
// e1? matches zero or one.
// The sequence of execution is first alternation, then concatenation, and finally the repetition operators.
// A backreference like \1 or \2 matches the string matched by a previous parenthesized expression,
// and only that string: (cat|dog)\1 matches catcat and dogdog but not catdog nor dogcat.
// A deterministic finite automaton (DFA), because in any input, each possible input letter leads to at most one new input

type NfaFrag struct {
	input *NfaState
	next  *NfaState
}
type NfaState struct {
	state rune
	next  *NfaState
	next2 *NfaState
}

type stack struct {
	val []interface{}
}

func (s *stack) push(val interface{}) {
	s.val = append(s.val, val)
}

func (s *stack) pop() interface{} {
	val := s.val[len(s.val)-1]
	s.val = s.val[:len(s.val)-1]
	return val
}

func (s *stack) top() interface{} {
	return s.val[len(s.val)-1]
}

func (s *stack) isEmpty() bool {
	return len(s.val) == 0
}

func toNfa(input []rune) {
	var (
		fragStack = stack{}
	)
	for _, c := range input {
		switch c {
		case ALTERN:
			e2 := fragStack.pop().(*NfaFrag)
			e1 := fragStack.pop().(*NfaFrag)
			fragStack.push(&NfaFrag{
				input: &NfaState{
					state: -1,
					next:  e1.next,
					next2: e2.next,
				},
				next:  ,
			})
		case CONCAT:
		case KLEENE:
		case '?':
		case '+':
		default:
			fragStack.push(&NfaFrag{
				input: &NfaState{
					state: c,
					next:  nil,
					next2: nil,
				},
				next: nil,
			})
		}
	}
}

// 1. Collation-related bracket symbols [==] [::] [..]
// 2. Escaped characters \
// 3. Character set (bracket expression) []
// 4. Grouping ()
// 5. Single-character-ERE duplication * + ? {m,n}
// 6. Concatenation
// 7. Anchoring ^$
// 8. Alternation |
const (
	LPAREN = iota + 1
	RPAREN
	ALTERN
	CONCAT
	KLEENE
)

// (a|b)*a
func parse(str string) []rune {
	parenCount := 0
	isConcat := false
	arr := make([]rune, 0)
	for _, c := range str {
		switch c {
		case '(':
			parenCount += 1
			isConcat = false
			arr = append(arr, LPAREN)
			continue
		case ')':
			parenCount -= 1
			arr = append(arr, RPAREN)
		case '*':
			arr = append(arr, KLEENE)
		case '|':
			arr = append(arr, ALTERN)
			isConcat = false
			continue
		default:
			if isConcat {
				arr = append(arr, CONCAT)
			}
			arr = append(arr, c)
		}
		isConcat = true
		if parenCount < 0 {
			panic("Unbalanced regular expression")
		}
	}
	return arr
}

// The compilation of the . that follows pops the two b NFA fragment from the stack and pushes an NFA fragment for the
// concatenation bb.. Each NFA fragment is defined by its start input and its outgoing arrow
// https://www.boost.org/doc/libs/1_56_0/libs/regex/doc/html/boost_regex/syntax/basic_extended.html#boost_regex.syntax.basic_extended.operator_precedence
func infixToPostfix(infix []rune) []rune {
	postfix := make([]rune, 0)
	opStack := &stack{}
	for _, c := range infix {
		switch {
		case c == LPAREN:
			opStack.push(c)
		case c == RPAREN:
			for opStack.top().(int32) != LPAREN {
				postfix = append(postfix, opStack.pop().(rune))
			}
			opStack.pop()
		case c == ALTERN:
			fallthrough
		case c == CONCAT:
			fallthrough
		case c == KLEENE:
			for !opStack.isEmpty() && c <= opStack.top().(int32) && opStack.top() != LPAREN {
				postfix = append(postfix, opStack.pop().(int32))
			}
			opStack.push(c)
		default:
			postfix = append(postfix, c)
		}
	}
	for !opStack.isEmpty() {
		postfix = append(postfix, opStack.pop().(rune))
	}
	return postfix
}

//3 + 4 × 2 ÷ ( 1 − 5 ) ^ 2 ^ 3
//right associate rule is not handled yet
func shuntingYard(input string) string {
	var (
		operatorStack []rune
		currNum       int
		output        string
		precedence    = map[rune]int{
			'+': 1,
			'-': 1,
			'×': 2,
			'÷': 2,
			'^': 3,
		}
	)
	str := []rune(input)
	for i := 0; i < len(str); i++ {
		if str[i] == ' ' {
			continue
		}
		isNum := false
		for i < len(str) && str[i] >= '0' && str[i] <= '9' {
			isNum = true
			currNum = int(str[i]-'0') + currNum*10
			i++
		}
		if isNum {
			output += strconv.Itoa(currNum)
			currNum = 0
			continue
		}
		if str[i] == '(' {
			operatorStack = append(operatorStack, str[i])
		} else if str[i] == ')' {
			for operatorStack[len(operatorStack)-1] != '(' {
				output += string(operatorStack[len(operatorStack)-1])
				operatorStack = operatorStack[:len(operatorStack)-1]
			}
			operatorStack = operatorStack[:len(operatorStack)-1]
		} else {
			currOrder := precedence[str[i]]
			for len(operatorStack) != 0 && currOrder <= precedence[operatorStack[len(operatorStack)-1]] && operatorStack[len(operatorStack)-1] != '(' {
				output += string(operatorStack[len(operatorStack)-1])
				operatorStack = operatorStack[:len(operatorStack)-1]
			}
			operatorStack = append(operatorStack, str[i])
		}
	}
	for len(operatorStack) != 0 {
		output += string(operatorStack[len(operatorStack)-1])
		operatorStack = operatorStack[:len(operatorStack)-1]
	}
	return output
}

func main() {
	fmt.Println(infixToPostfix(parse("(ab)*c")))    // "ab.c*"
	fmt.Println(infixToPostfix(parse("(a(b|d))*"))) // "abd|.*"
	fmt.Println(infixToPostfix(parse("a(bb)+c")))   // "abb..c.+"
}
