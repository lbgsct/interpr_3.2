package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type stackLayer struct {
	variables map[string]varData
	functions map[string]funcData
}

type memoryStack []stackLayer

var memory memoryStack

type funcData struct {
	arguments []string
	body      []string
}

type varData struct {
	varType string
	value   string
}

func (m memoryStack) display(expression string) {
	fmt.Printf("\nVariables:\n")
	tokens := tokenize(expression[5:])
	if len(tokens) > 1 {
		m.displayVariable(tokens)
	} else {
		for lvl := len(m) - 1; lvl >= 0; lvl-- {
			for name, variable := range m[lvl].variables {
				fmt.Printf("%v = %v\n", name, variable.value)
			}
		}
	}
}

func (m memoryStack) displayVariable(tokens []string) {
	fmt.Printf("\n")

	variable := ""
	for i := 0; i < len(tokens); i += 2 {
		for lvl := len(m) - 1; lvl >= 0; lvl-- {
			if _, ok := m[lvl].variables[tokens[i]]; ok {
				variable = m[lvl].variables[tokens[i]].value
				break
			}
		}
		if variable != "" {
			fmt.Printf("%v = %v\n", tokens[i], variable)
			variable = ""
		}
	}
}

func initMemory() {
	memory = make(memoryStack, 0)
}

func addLayer() {
	memory = append(memory, stackLayer{make(map[string]varData), make(map[string]funcData)})
}

func storeFunc(expression string) {
	tokens := tokenize(expression)

	args := make([]string, 0)
	i := 2
	for ; tokens[i] != ")"; i++ {
		if tokens[i] != "," {
			args = append(args, tokens[i])
		}
	}

	body := make([]string, 0)
	i += 2
	for ; i < len(tokens)-1; i++ {
		body = append(body, tokens[i])
	}

	memory[len(memory)-1].functions[tokens[0]] = funcData{args, body}
}

func storeVar(expression string) string {
	tokens := tokenize(expression)
	var varType string
	switch tokens[2] {
	case "i":
		varType = "int"
	case "f":
		varType = "float"
	default:
		return "ERROR"
	}

	memory[len(memory)-1].variables[tokens[0]] = varData{varType, tokens[5]}

	return ""
}

func updateVariable(name string, newValue string) {
	for lvl := len(memory) - 1; lvl >= 0; lvl-- {
		variable := memory[lvl].variables[name]
		variable.value = newValue
		memory[lvl].variables[name] = variable
		return
	}
}

func solveInfixFunction(f funcData, tokens []string) []string {
	arguments := make([][]string, len(f.arguments))
	for i := range arguments {
		arguments[i] = make([]string, 0)
	}

	funcCounter := 0
	funcTokens := make([]string, 0)
	var innerFunc funcData
	argIndex := 0
	result := make([]string, 0)
	result = append(result, "(")

	for index, token := range tokens {
		if index == 0 || index == len(tokens)-1 {
			continue
		}
		if funcCounter == 0 {
			for _, stackLvl := range memory {
				if _, ok := stackLvl.functions[token]; ok {
					innerFunc = stackLvl.functions[token]
					funcCounter++
					break
				}
			}
			if funcCounter > 0 {
				continue
			}
			if token == "," {
				argIndex++
			} else {
				arguments[argIndex] = append(arguments[argIndex], token)
			}
		} else {
			funcTokens = append(funcTokens, token)
			if token == ")" {
				funcCounter--
				if funcCounter == 1 {
					parsedTokens := solveInfixFunction(innerFunc, funcTokens)
					arguments[argIndex] = append(arguments[argIndex], parsedTokens...)

					funcTokens = make([]string, 0)
					funcCounter = 0
				}
			} else if token == "(" {
				funcCounter++
			}
		}
	}

	argsMap := make(map[string][]string)
	for i := range f.arguments {
		argsMap[f.arguments[i]] = arguments[i]
	}

	for _, token := range f.body {
		if arg, ok := argsMap[token]; ok {
			result = append(result, "(")
			result = append(result, arg...)
			result = append(result, ")")
		} else {
			result = append(result, token)
		}
	}
	result = append(result, ")")

	return result
}

func handleMinus(prev string) string {
	if prev == "(" || prev == "/" || prev == "*" || prev == "+" || prev == "-" {
		return "~"
	}
	return "-"
}

func findVariable(token string) string {
	for lvl := len(memory) - 1; lvl >= 0; lvl-- {
		if variable, ok := memory[lvl].variables[token]; ok {
			return variable.value
		}
	}
	return "NO"
}

func tokenize(expression string) []string {
	expression = strings.ReplaceAll(expression, " ", "")

	tokens := make([]string, 0)
	var token strings.Builder

	for _, c := range expression {
		if c == '+' || c == '-' || c == '*' || c == '/' || c == '(' || c == ')' ||
			c == ',' || c == ':' || c == ';' || c == '=' {
			if token.Len() > 0 {
				tokens = append(tokens, token.String())
				token.Reset()
			}
			tokens = append(tokens, string(c))
		} else {
			token.Grow(1)
			token.WriteRune(c)
		}
	}

	for i, token := range tokens {
		if token == "-" {
			if i == 0 {
				tokens[i] = "~"
			} else {
				tokens[i] = handleMinus(tokens[i-1])
			}
		}
	}
	return tokens
}

func assignExpression(lineNum int, expression string, notation string) string {
	tokens := make([]string, 0)
	tokens = append(tokens, "(")
	tokens = append(tokens, tokenize(expression)...)
	tokens = append(tokens[:len(tokens)-1], ")")

	rpn := make([]string, 0)
	switch notation {
	case "prefix":

	case "infix":
		tokens = solveInfixFunction(memory[0].functions["expression"], tokens)

		stack := make([]string, 0)

		for _, token := range tokens {
			if _, ok := strconv.ParseFloat(token, 64); ok == nil {
				rpn = append(rpn, token)
			} else if findVariable(token) != "NO" {
				rpn = append(rpn, findVariable(token))
			} else if token == "~" || token == "+" || token == "-" || token == "*" || token == "/" || token == "(" || token == ")" {
				switch token {
				case "~", "(":
					stack = append(stack, token)
				case "*", "/":
					for len(stack) > 0 && stack[len(stack)-1] == "~" {
						rpn = append(rpn, stack[len(stack)-1])
						stack = stack[:len(stack)-1]
					}
					stack = append(stack, token)
				case "+", "-":
					for len(stack) > 0 && (stack[len(stack)-1] == "~" || stack[len(stack)-1] == "*" || stack[len(stack)-1] == "/") {
						rpn = append(rpn, stack[len(stack)-1])
						stack = stack[:len(stack)-1]
					}
					stack = append(stack, token)
				case ")":
					for len(stack) > 0 {
						if stack[len(stack)-1] == "(" {
							stack = stack[:len(stack)-1]
							break
						}
						rpn = append(rpn, stack[len(stack)-1])
						stack = stack[:len(stack)-1]
					}
				}
			} else {
				return "[" + strconv.Itoa(lineNum) + "] ERROR: unknown token: " + token
			}
		}
	case "postfix":

	default:
		return "[" + strconv.Itoa(lineNum) + "] ERROR: unknown notation: " + notation
	}

	index := 2
	if rpn[2] == "~" {
		index--
	}

	for len(rpn) > 1 {
		switch rpn[index] {
		case "~":
			if rpn[index-1][0] == '-' {
				rpn[index-1] = rpn[index-1][1:]
			} else {
				rpn[index-1] = "-" + rpn[index-1]
			}
			rpn = append(rpn[:index], rpn[index+1:]...)
			index--
		case "+", "-", "*", "/":
			a, ok := strconv.ParseFloat(rpn[index-2], 64)
			if ok != nil {
				return "[" + strconv.Itoa(lineNum) + "] ERROR: can't parse float " + rpn[index-2]
			}
			b, ok := strconv.ParseFloat(rpn[index-1], 64)
			if ok != nil {
				return "[" + strconv.Itoa(lineNum) + "] ERROR: can't parse float " + rpn[index-1]
			}

			switch rpn[index] {
			case "+":
				rpn[index-2] = fmt.Sprintf("%v", a+b)
			case "-":
				rpn[index-2] = fmt.Sprintf("%v", a-b)
			case "*":
				rpn[index-2] = fmt.Sprintf("%v", a*b)
			case "/":
				rpn[index-2] = fmt.Sprintf("%v", a/b)
			}
			rpn = append(rpn[:index-1], rpn[index+1:]...)
			index -= 2
		default:
			index++
		}
	}
	return strings.Join(rpn, "")
}

func Execute(filePath string) {
	initMemory()
	addLayer()
	memory[0].functions["expression"] = funcData{[]string{"x"}, []string{"x"}}

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Printf("Couldn't find the file \"%v\": %v\n", filePath, err)
	} else {
		lineNum := 0
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			lineNum++

			if strings.Contains(line, "(i)=") || strings.Contains(line, "(f)=") {
				if storeVar(line) == "ERROR" {
					fmt.Printf("[%v] Couldn't save variable, invalid type\n", lineNum)
				}
			} else if strings.Contains(line, "print") {
				memory.display(line)
			} else if strings.Contains(line, ":") {
				storeFunc(line)
			} else if strings.Contains(line, "=") {
				var saveLine strings.Builder
				saveLine.WriteString(line[:strings.Index(line, "=")+1])
				assignation := assignExpression(lineNum, line[strings.Index(line, "=")+1:], "infix")
				if assignation == "ERROR" {
					fmt.Printf("[line %v] Couldn't assign value\n", lineNum)
				} else {
					saveLine.WriteString(assignation)
					updateVariable(line[:strings.Index(line, "=")], assignation)
				}
			}
		}
	}
}

func main() {
	Execute("/home/sergey/micro/32/32.txt")
}