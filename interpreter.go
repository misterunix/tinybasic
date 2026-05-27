package tinybasic

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"unicode"
)

const Version = "0.0.8"

// Interpreter represents a BASIC interpreter instance
type Interpreter struct {
	program        []Statement
	variables      map[string]float64
	labels         map[string]int // label name to statement index
	currentStmt    int
	callStack      []int // for GOSUB/RETURN
	done           bool
	inputCallback  func() (string, error)
	outputCallback func(string)
	dataValues     []string
	dataPointer    int
	forLoops       map[string]*ForLoopState
	ioPorts        map[uint]int64 // IO ports for IN/OUT functions
}

// ForLoopState tracks FOR loop state
type ForLoopState struct {
	variable   string
	endValue   float64
	stepValue  float64
	startIndex int
}

// Statement represents a parsed BASIC statement
type Statement struct {
	Label   string
	Command string
	Args    []Token
	Line    string
}

// Token types
const (
	TokenNumber     = "NUMBER"
	TokenString     = "STRING"
	TokenIdentifier = "IDENTIFIER"
	TokenOperator   = "OPERATOR"
	TokenFunction   = "FUNCTION"
	TokenComma      = "COMMA"
	TokenLeftParen  = "LPAREN"
	TokenRightParen = "RPAREN"
	TokenEOF        = "EOF"
)

// Token represents a lexical token
type Token struct {
	Type  string
	Value string
}

// New creates a new interpreter instance
func New() *Interpreter {
	return &Interpreter{
		program:     make([]Statement, 0),
		variables:   make(map[string]float64),
		labels:      make(map[string]int),
		currentStmt: 0,
		done:        false,
		forLoops:    make(map[string]*ForLoopState),
		ioPorts:     make(map[uint]int64),
		outputCallback: func(s string) {
			fmt.Print(s)
		},
	}
}

// SetInputCallback sets the callback for INPUT statements
func (interp *Interpreter) SetInputCallback(callback func() (string, error)) {
	interp.inputCallback = callback
}

// SetOutputCallback sets the callback for PRINT statements
func (interp *Interpreter) SetOutputCallback(callback func(string)) {
	interp.outputCallback = callback
}

// LoadProgram loads a program from source code
func (interp *Interpreter) LoadProgram(source string) error {
	lines := strings.Split(source, "\n")
	interp.program = make([]Statement, 0)
	interp.labels = make(map[string]int)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "REM") {
			continue
		}

		stmt, err := interp.parseLine(line)
		if err != nil {
			return fmt.Errorf("parse error: %s - %s", line, err)
		}

		if stmt.Label != "" {
			interp.labels[stmt.Label] = len(interp.program)
		}

		interp.program = append(interp.program, stmt)
	}

	return nil
}

// parseLine parses a single line of BASIC code
func (interp *Interpreter) parseLine(line string) (Statement, error) {
	stmt := Statement{Line: line}

	// Check for label (e.g., START: PRINT "Hello")
	if idx := strings.Index(line, ":"); idx > 0 {
		potential := strings.TrimSpace(line[:idx])
		if isValidLabel(potential) {
			stmt.Label = potential
			line = strings.TrimSpace(line[idx+1:])
		}
	}

	// Parse command and arguments
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return stmt, nil
	}

	stmt.Command = strings.ToUpper(parts[0])

	// Tokenize the rest of the line
	if len(line) > len(parts[0]) {
		argString := strings.TrimSpace(line[len(parts[0]):])
		tokens, err := tokenize(argString)
		if err != nil {
			return stmt, err
		}
		stmt.Args = tokens
	}

	// Check if this is an implicit LET (variable = expression)
	// If the command looks like a variable name and is followed by =, treat as LET
	basicKeywords := map[string]bool{
		"LET": true, "PRINT": true, "INPUT": true, "IF": true, "THEN": true,
		"GOTO": true, "GOSUB": true, "RETURN": true, "FOR": true, "NEXT": true,
		"TO": true, "STEP": true, "DATA": true, "READ": true, "END": true,
		"REM": true,
	}
	if stmt.Command != "" && !basicKeywords[stmt.Command] && isValidLabel(stmt.Command) {
		// Check if there's an = sign in the args
		hasEquals := false
		for _, tok := range stmt.Args {
			if tok.Type == TokenOperator && tok.Value == "=" {
				hasEquals = true
				break
			}
		}
		if hasEquals {
			// This is an implicit LET - prepend the variable name to args
			varToken := Token{Type: TokenIdentifier, Value: stmt.Command}
			stmt.Args = append([]Token{varToken}, stmt.Args...)
			stmt.Command = "LET"
		}
	}

	return stmt, nil
}

// isValidLabel checks if a string is a valid label
func isValidLabel(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, ch := range s {
		if i == 0 {
			if !unicode.IsLetter(ch) {
				return false
			}
		} else {
			if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
				return false
			}
		}
	}
	return true
}

// tokenize breaks a string into tokens
func tokenize(input string) ([]Token, error) {
	tokens := make([]Token, 0)
	i := 0

	for i < len(input) {
		ch := input[i]

		// Skip whitespace
		if unicode.IsSpace(rune(ch)) {
			i++
			continue
		}

		// String literal
		if ch == '"' {
			start := i + 1
			i++
			for i < len(input) && input[i] != '"' {
				i++
			}
			if i >= len(input) {
				return nil, fmt.Errorf("unterminated string")
			}
			tokens = append(tokens, Token{Type: TokenString, Value: input[start:i]})
			i++
			continue
		}

		// Number
		if unicode.IsDigit(rune(ch)) || (ch == '.' && i+1 < len(input) && unicode.IsDigit(rune(input[i+1]))) {
			start := i
			hasDecimal := false
			for i < len(input) {
				if unicode.IsDigit(rune(input[i])) {
					i++
				} else if input[i] == '.' && !hasDecimal {
					hasDecimal = true
					i++
				} else {
					break
				}
			}
			tokens = append(tokens, Token{Type: TokenNumber, Value: input[start:i]})
			continue
		}

		// Operators and comparisons
		if ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '^' || ch == '=' || ch == '<' || ch == '>' {
			start := i
			i++
			// Check for two-character operators
			if i < len(input) {
				twoChar := input[start : i+1]
				if twoChar == "<=" || twoChar == ">=" || twoChar == "<>" || twoChar == "==" {
					i++
				}
			}
			tokens = append(tokens, Token{Type: TokenOperator, Value: input[start:i]})
			continue
		}

		// Comma
		if ch == ',' {
			tokens = append(tokens, Token{Type: TokenComma, Value: ","})
			i++
			continue
		}

		// Parentheses
		if ch == '(' {
			tokens = append(tokens, Token{Type: TokenLeftParen, Value: "("})
			i++
			continue
		}
		if ch == ')' {
			tokens = append(tokens, Token{Type: TokenRightParen, Value: ")"})
			i++
			continue
		}

		// Identifier or function
		if unicode.IsLetter(rune(ch)) {
			start := i
			for i < len(input) && (unicode.IsLetter(rune(input[i])) || unicode.IsDigit(rune(input[i])) || input[i] == '_' || input[i] == '$') {
				i++
			}
			name := input[start:i]
			nameUpper := strings.ToUpper(name)

			// Check if it's a function
			if isMathFunction(nameUpper) {
				tokens = append(tokens, Token{Type: TokenFunction, Value: nameUpper})
			} else {
				tokens = append(tokens, Token{Type: TokenIdentifier, Value: name})
			}
			continue
		}

		return nil, fmt.Errorf("unexpected character: %c", ch)
	}

	return tokens, nil
}

// isMathFunction checks if a name is a math function
func isMathFunction(name string) bool {
	functions := []string{"SIN", "COS", "TAN", "ASIN", "ACOS", "ATAN",
		"LOG", "LN", "EXP", "SQRT", "ABS", "INT", "RND", "SGN",
		"FLOOR", "CEIL", "ROUND", "POW", "MIN", "MAX", "IN", "OUT"}
	for _, f := range functions {
		if name == f {
			return true
		}
	}
	return false
}

// Reset resets the interpreter state for a new run
func (interp *Interpreter) Reset() {
	interp.variables = make(map[string]float64)
	interp.currentStmt = 0
	interp.callStack = make([]int, 0)
	interp.done = false
	interp.dataPointer = 0
	interp.forLoops = make(map[string]*ForLoopState)
}

// Run executes the entire program
func (interp *Interpreter) Run() error {
	interp.Reset()
	for !interp.done && interp.currentStmt < len(interp.program) {
		if err := interp.Step(); err != nil {
			return err
		}
	}
	return nil
}

// Step executes a single statement
func (interp *Interpreter) Step() error {
	if interp.done || interp.currentStmt >= len(interp.program) {
		interp.done = true
		return nil
	}

	stmt := interp.program[interp.currentStmt]

	switch stmt.Command {
	case "LET":
		return interp.executeLet(stmt)
	case "PRINT":
		return interp.executePrint(stmt)
	case "INPUT":
		return interp.executeInput(stmt)
	case "IF":
		return interp.executeIf(stmt)
	case "GOTO":
		return interp.executeGoto(stmt)
	case "GOSUB":
		return interp.executeGosub(stmt)
	case "RETURN":
		return interp.executeReturn(stmt)
	case "FOR":
		return interp.executeFor(stmt)
	case "NEXT":
		return interp.executeNext(stmt)
	case "DATA":
		return interp.executeData(stmt)
	case "READ":
		return interp.executeRead(stmt)
	case "END":
		interp.done = true
		return nil
	case "REM":
		interp.currentStmt++
		return nil
	default:
		return fmt.Errorf("unknown command: %s", stmt.Command)
	}
}

// executeLet handles LET variable = expression
func (interp *Interpreter) executeLet(stmt Statement) error {
	if len(stmt.Args) < 3 {
		return fmt.Errorf("LET requires variable = expression")
	}

	if stmt.Args[0].Type != TokenIdentifier {
		return fmt.Errorf("LET requires a variable name")
	}

	varName := stmt.Args[0].Value

	// Find the equals sign
	eqIdx := -1
	for i, tok := range stmt.Args {
		if tok.Type == TokenOperator && tok.Value == "=" {
			eqIdx = i
			break
		}
	}

	if eqIdx == -1 {
		return fmt.Errorf("LET requires = operator")
	}

	// Evaluate expression after the equals sign
	value, err := interp.evaluateExpression(stmt.Args[eqIdx+1:])
	if err != nil {
		return err
	}

	interp.variables[varName] = value
	interp.currentStmt++
	return nil
}

// executePrint handles PRINT expressions
func (interp *Interpreter) executePrint(stmt Statement) error {
	output := ""
	i := 0

	for i < len(stmt.Args) {
		if stmt.Args[i].Type == TokenString {
			output += stmt.Args[i].Value
			i++
		} else if stmt.Args[i].Type == TokenComma {
			output += " "
			i++
		} else {
			// Find the next comma or end
			exprEnd := i
			for exprEnd < len(stmt.Args) && stmt.Args[exprEnd].Type != TokenComma {
				exprEnd++
			}

			value, err := interp.evaluateExpression(stmt.Args[i:exprEnd])
			if err != nil {
				return err
			}

			// Format output
			if value == float64(int64(value)) {
				output += fmt.Sprintf("%d", int64(value))
			} else {
				output += fmt.Sprintf("%g", value)
			}

			i = exprEnd
		}
	}

	interp.outputCallback(output + "\n")
	interp.currentStmt++
	return nil
}

// executeInput handles INPUT variable
func (interp *Interpreter) executeInput(stmt Statement) error {
	if len(stmt.Args) == 0 {
		return fmt.Errorf("INPUT requires a variable name")
	}

	// Support optional prompt string
	varIdx := 0
	if stmt.Args[0].Type == TokenString {
		interp.outputCallback(stmt.Args[0].Value)
		varIdx = 1
		if varIdx < len(stmt.Args) && stmt.Args[varIdx].Type == TokenComma {
			varIdx++
		}
	}

	if varIdx >= len(stmt.Args) || stmt.Args[varIdx].Type != TokenIdentifier {
		return fmt.Errorf("INPUT requires a variable name")
	}

	varName := stmt.Args[varIdx].Value

	if interp.inputCallback == nil {
		return fmt.Errorf("no input callback set")
	}

	input, err := interp.inputCallback()
	if err != nil {
		return err
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
	if err != nil {
		return fmt.Errorf("invalid number input: %s", input)
	}

	interp.variables[varName] = value
	interp.currentStmt++
	return nil
}

// executeIf handles IF condition THEN statement
func (interp *Interpreter) executeIf(stmt Statement) error {
	// Find THEN keyword
	thenIdx := -1
	for i, tok := range stmt.Args {
		if tok.Type == TokenIdentifier && strings.ToUpper(tok.Value) == "THEN" {
			thenIdx = i
			break
		}
	}

	if thenIdx == -1 {
		return fmt.Errorf("IF requires THEN")
	}

	// Evaluate condition
	condition, err := interp.evaluateCondition(stmt.Args[:thenIdx])
	if err != nil {
		return err
	}

	if condition {
		// Check if THEN is followed by a label
		if thenIdx+1 < len(stmt.Args) && stmt.Args[thenIdx+1].Type == TokenIdentifier {
			label := stmt.Args[thenIdx+1].Value
			if idx, ok := interp.labels[label]; ok {
				interp.currentStmt = idx
				return nil
			}
		}
		// Otherwise just continue to next statement
	}

	interp.currentStmt++
	return nil
}

// executeGoto handles GOTO label
func (interp *Interpreter) executeGoto(stmt Statement) error {
	if len(stmt.Args) == 0 || stmt.Args[0].Type != TokenIdentifier {
		return fmt.Errorf("GOTO requires a label")
	}

	label := stmt.Args[0].Value
	idx, ok := interp.labels[label]
	if !ok {
		return fmt.Errorf("undefined label: %s", label)
	}

	interp.currentStmt = idx
	return nil
}

// executeGosub handles GOSUB label
func (interp *Interpreter) executeGosub(stmt Statement) error {
	if len(stmt.Args) == 0 || stmt.Args[0].Type != TokenIdentifier {
		return fmt.Errorf("GOSUB requires a label")
	}

	label := stmt.Args[0].Value
	idx, ok := interp.labels[label]
	if !ok {
		return fmt.Errorf("undefined label: %s", label)
	}

	// Push return address
	interp.callStack = append(interp.callStack, interp.currentStmt+1)
	interp.currentStmt = idx
	return nil
}

// executeReturn handles RETURN
func (interp *Interpreter) executeReturn(stmt Statement) error {
	if len(interp.callStack) == 0 {
		return fmt.Errorf("RETURN without GOSUB")
	}

	// Pop return address
	interp.currentStmt = interp.callStack[len(interp.callStack)-1]
	interp.callStack = interp.callStack[:len(interp.callStack)-1]
	return nil
}

// executeFor handles FOR variable = start TO end [STEP step]
func (interp *Interpreter) executeFor(stmt Statement) error {
	if len(stmt.Args) < 5 {
		return fmt.Errorf("FOR requires variable = start TO end")
	}

	varName := stmt.Args[0].Value

	// Find = sign
	eqIdx := -1
	for i, tok := range stmt.Args {
		if tok.Type == TokenOperator && tok.Value == "=" {
			eqIdx = i
			break
		}
	}
	if eqIdx == -1 {
		return fmt.Errorf("FOR requires = operator")
	}

	// Find TO
	toIdx := -1
	for i, tok := range stmt.Args {
		if tok.Type == TokenIdentifier && strings.ToUpper(tok.Value) == "TO" {
			toIdx = i
			break
		}
	}
	if toIdx == -1 {
		return fmt.Errorf("FOR requires TO")
	}

	// Find STEP (optional)
	stepIdx := -1
	for i, tok := range stmt.Args {
		if tok.Type == TokenIdentifier && strings.ToUpper(tok.Value) == "STEP" {
			stepIdx = i
			break
		}
	}

	// Evaluate start value
	startVal, err := interp.evaluateExpression(stmt.Args[eqIdx+1 : toIdx])
	if err != nil {
		return err
	}

	// Evaluate end value
	var endVal float64
	if stepIdx == -1 {
		endVal, err = interp.evaluateExpression(stmt.Args[toIdx+1:])
	} else {
		endVal, err = interp.evaluateExpression(stmt.Args[toIdx+1 : stepIdx])
	}
	if err != nil {
		return err
	}

	// Evaluate step value
	stepVal := 1.0
	if stepIdx != -1 {
		stepVal, err = interp.evaluateExpression(stmt.Args[stepIdx+1:])
		if err != nil {
			return err
		}
	}

	// Initialize loop
	interp.variables[varName] = startVal
	interp.forLoops[varName] = &ForLoopState{
		variable:   varName,
		endValue:   endVal,
		stepValue:  stepVal,
		startIndex: interp.currentStmt,
	}

	interp.currentStmt++
	return nil
}

// executeNext handles NEXT variable
func (interp *Interpreter) executeNext(stmt Statement) error {
	if len(stmt.Args) == 0 {
		return fmt.Errorf("NEXT requires a variable name")
	}

	varName := stmt.Args[0].Value
	loop, ok := interp.forLoops[varName]
	if !ok {
		return fmt.Errorf("NEXT without FOR: %s", varName)
	}

	// Increment variable
	interp.variables[varName] += loop.stepValue

	// Check if loop should continue
	currentVal := interp.variables[varName]
	shouldContinue := false
	if loop.stepValue > 0 {
		shouldContinue = currentVal <= loop.endValue
	} else {
		shouldContinue = currentVal >= loop.endValue
	}

	if shouldContinue {
		// Jump back to the statement after FOR
		interp.currentStmt = loop.startIndex + 1
	} else {
		// Exit loop
		delete(interp.forLoops, varName)
		interp.currentStmt++
	}

	return nil
}

// executeData handles DATA values
func (interp *Interpreter) executeData(stmt Statement) error {
	// Store data values for READ statements
	for _, tok := range stmt.Args {
		if tok.Type == TokenString {
			interp.dataValues = append(interp.dataValues, tok.Value)
		} else if tok.Type == TokenNumber {
			interp.dataValues = append(interp.dataValues, tok.Value)
		}
	}
	interp.currentStmt++
	return nil
}

// executeRead handles READ variables
func (interp *Interpreter) executeRead(stmt Statement) error {
	for _, tok := range stmt.Args {
		if tok.Type == TokenComma {
			continue
		}
		if tok.Type != TokenIdentifier {
			return fmt.Errorf("READ requires variable names")
		}

		if interp.dataPointer >= len(interp.dataValues) {
			return fmt.Errorf("out of DATA")
		}

		value, err := strconv.ParseFloat(interp.dataValues[interp.dataPointer], 64)
		if err != nil {
			return fmt.Errorf("invalid data value: %s", interp.dataValues[interp.dataPointer])
		}

		interp.variables[tok.Value] = value
		interp.dataPointer++
	}

	interp.currentStmt++
	return nil
}

// evaluateExpression evaluates a mathematical expression
func (interp *Interpreter) evaluateExpression(tokens []Token) (float64, error) {
	if len(tokens) == 0 {
		return 0, fmt.Errorf("empty expression")
	}

	// Convert to postfix notation using Shunting Yard algorithm
	postfix, err := interp.toPostfix(tokens)
	if err != nil {
		return 0, err
	}

	// Evaluate postfix expression
	return interp.evaluatePostfix(postfix)
}

// toPostfix converts infix expression to postfix
func (interp *Interpreter) toPostfix(tokens []Token) ([]Token, error) {
	output := make([]Token, 0)
	stack := make([]Token, 0)

	precedence := map[string]int{
		"^": 4,
		"*": 3,
		"/": 3,
		"+": 2,
		"-": 2,
	}

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		switch tok.Type {
		case TokenNumber:
			output = append(output, tok)

		case TokenIdentifier:
			// Variable
			output = append(output, tok)

		case TokenFunction:
			stack = append(stack, tok)

		case TokenLeftParen:
			stack = append(stack, tok)

		case TokenRightParen:
			// Pop until left paren
			for len(stack) > 0 && stack[len(stack)-1].Type != TokenLeftParen {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, fmt.Errorf("mismatched parentheses")
			}
			stack = stack[:len(stack)-1] // Remove left paren

			// If there's a function on the stack, pop it
			if len(stack) > 0 && stack[len(stack)-1].Type == TokenFunction {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}

		case TokenOperator:
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.Type != TokenOperator {
					break
				}
				if (tok.Value != "^" && precedence[tok.Value] <= precedence[top.Value]) ||
					(tok.Value == "^" && precedence[tok.Value] < precedence[top.Value]) {
					output = append(output, top)
					stack = stack[:len(stack)-1]
				} else {
					break
				}
			}
			stack = append(stack, tok)

		case TokenComma:
			// For function arguments - treat like right paren without removing paren
			for len(stack) > 0 && stack[len(stack)-1].Type != TokenLeftParen {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
		}
	}

	// Pop remaining operators
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		if top.Type == TokenLeftParen {
			return nil, fmt.Errorf("mismatched parentheses")
		}
		output = append(output, top)
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

// evaluatePostfix evaluates a postfix expression
func (interp *Interpreter) evaluatePostfix(tokens []Token) (float64, error) {
	stack := make([]float64, 0)

	for _, tok := range tokens {
		switch tok.Type {
		case TokenNumber:
			val, err := strconv.ParseFloat(tok.Value, 64)
			if err != nil {
				return 0, err
			}
			stack = append(stack, val)

		case TokenIdentifier:
			// Variable lookup
			val, ok := interp.variables[tok.Value]
			if !ok {
				return 0, fmt.Errorf("undefined variable: %s", tok.Value)
			}
			stack = append(stack, val)

		case TokenOperator:
			if len(stack) < 2 {
				return 0, fmt.Errorf("invalid expression")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			var result float64
			switch tok.Value {
			case "+":
				result = a + b
			case "-":
				result = a - b
			case "*":
				result = a * b
			case "/":
				if b == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				result = a / b
			case "^":
				result = math.Pow(a, b)
			default:
				return 0, fmt.Errorf("unknown operator: %s", tok.Value)
			}
			stack = append(stack, result)

		case TokenFunction:
			if len(stack) < 1 {
				return 0, fmt.Errorf("function requires argument")
			}

			// Handle functions with different arities
			var result float64
			switch tok.Value {
			case "MIN", "MAX", "POW", "OUT":
				if len(stack) < 2 {
					return 0, fmt.Errorf("%s requires 2 arguments", tok.Value)
				}
				b := stack[len(stack)-1]
				a := stack[len(stack)-2]
				stack = stack[:len(stack)-2]

				switch tok.Value {
				case "MIN":
					result = math.Min(a, b)
				case "MAX":
					result = math.Max(a, b)
				case "POW":
					result = math.Pow(a, b)
				case "OUT":
					// OUT(port, value) - write value to port
					port := uint(a)
					value := int64(b)
					interp.ioPorts[port] = value
					result = float64(value)
				}
			default:
				// Single argument functions
				arg := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				switch tok.Value {
				case "SIN":
					result = math.Sin(arg)
				case "COS":
					result = math.Cos(arg)
				case "TAN":
					result = math.Tan(arg)
				case "ASIN":
					result = math.Asin(arg)
				case "ACOS":
					result = math.Acos(arg)
				case "ATAN":
					result = math.Atan(arg)
				case "LOG":
					result = math.Log10(arg)
				case "LN":
					result = math.Log(arg)
				case "EXP":
					result = math.Exp(arg)
				case "SQRT":
					result = math.Sqrt(arg)
				case "ABS":
					result = math.Abs(arg)
				case "INT":
					result = float64(int64(arg))
				case "RND":
					// Simple random - could be improved
					result = rand.Float64() * math.Abs(arg)
					//result = math.Mod(math.Abs(arg)*9301+49297, 233280) / 233280
				case "SGN":
					if arg > 0 {
						result = 1
					} else if arg < 0 {
						result = -1
					} else {
						result = 0
					}
				case "FLOOR":
					result = math.Floor(arg)
				case "CEIL":
					result = math.Ceil(arg)
				case "ROUND":
					result = math.Round(arg)
				case "IN":
					// IN(port) - read value from port
					port := uint(arg)
					value, exists := interp.ioPorts[port]
					if exists {
						result = float64(value)
					} else {
						result = 0
					}
				default:
					return 0, fmt.Errorf("unknown function: %s", tok.Value)
				}
			}
			stack = append(stack, result)
		}
	}

	if len(stack) != 1 {
		return 0, fmt.Errorf("invalid expression")
	}

	return stack[0], nil
}

// evaluateCondition evaluates a boolean condition
func (interp *Interpreter) evaluateCondition(tokens []Token) (bool, error) {
	// Find comparison operator
	compIdx := -1
	compOp := ""
	for i, tok := range tokens {
		if tok.Type == TokenOperator {
			val := tok.Value
			if val == "=" || val == "==" || val == "<>" || val == "<" || val == ">" || val == "<=" || val == ">=" {
				compIdx = i
				compOp = val
				break
			}
		}
	}

	if compIdx == -1 {
		return false, fmt.Errorf("condition requires comparison operator")
	}

	// Evaluate left side
	left, err := interp.evaluateExpression(tokens[:compIdx])
	if err != nil {
		return false, err
	}

	// Evaluate right side
	right, err := interp.evaluateExpression(tokens[compIdx+1:])
	if err != nil {
		return false, err
	}

	// Compare
	switch compOp {
	case "=", "==":
		return left == right, nil
	case "<>":
		return left != right, nil
	case "<":
		return left < right, nil
	case ">":
		return left > right, nil
	case "<=":
		return left <= right, nil
	case ">=":
		return left >= right, nil
	}

	return false, fmt.Errorf("unknown comparison operator: %s", compOp)
}

// IsDone returns whether execution is complete
func (interp *Interpreter) IsDone() bool {
	return interp.done || interp.currentStmt >= len(interp.program)
}

// GetVariable returns the value of a variable
func (interp *Interpreter) GetVariable(name string) (float64, bool) {
	val, ok := interp.variables[name]
	return val, ok
}

// SetVariable sets the value of a variable
func (interp *Interpreter) SetVariable(name string, value float64) {
	interp.variables[name] = value
}

// GetAllVariables returns a copy of all variables
func (interp *Interpreter) GetAllVariables() map[string]float64 {
	vars := make(map[string]float64)
	for k, v := range interp.variables {
		vars[k] = v
	}
	return vars
}

// GetCurrentLine returns the current line number (statement index)
func (interp *Interpreter) GetCurrentLine() int {
	return interp.currentStmt
}

// GetProgramSize returns the number of statements in the program
func (interp *Interpreter) GetProgramSize() int {
	return len(interp.program)
}
