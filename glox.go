// Package glox is A Go implementation of the Lox programming language from "Crafting Interpreters"
// by Robert Nystrom (http://www.craftinginterpreters.com/).
package glox

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"
)

func RunFile(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read file: %v", err)
	}
	program, err := Parse[any](string(bytes))
	if err != nil {
		return err
	}
	return NewInterpreter(DefaultGlobals()).Interpret(program)
}

func RunPrompt() error {
	interpreter := NewInterpreter(DefaultGlobals())
	fmt.Println("Glox REPL. Press Ctrl+C to exit.")
	prompt := "> "
	scanner := bufio.NewScanner(os.Stdin)

	for fmt.Print(prompt); scanner.Scan(); fmt.Print(prompt) {
		program, err := Parse[any](scanner.Text())
		if err != nil {
			fmt.Println(err)
			continue
		}

		// if the input is a single expression, wrap it in a print statement.
		if expr, ok := program[0].(*ExpressionStmt[any]); ok && len(program) == 1 {
			program = Program[any]{&PrintStmt[any]{Expr: expr.Expr}}
		}

		if err := interpreter.Interpret(program); err != nil {
			fmt.Println(err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}
	return nil
}

func Parse[R any](source string) (Program[R], error) {
	scanner := NewScanner(source)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		return nil, err
	}
	return NewParser[R](tokens).Parse()
}

func DefaultGlobals() Env {
	env := NewEnvironment(nil)

	// add native functions here
	env.Define("clock", NewCallable("<native clock>", 0,
		func(*Interpreter, []any) (any, error) {
			return float64(time.Now().UnixMilli()), nil
		},
	))
	env.Define("prompt", NewCallable("<native prompt>", 1,
		func(i *Interpreter, args []any) (any, error) {
			fmt.Println(args[0])
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				return scanner.Text(), nil
			}
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("error reading input: %v", err)
			}
			return "", nil
		},
	))
	env.Define("number", NewCallable("<native number>", 1,
		func(i *Interpreter, args []any) (any, error) {
			arg := fmt.Sprintf("%v", args[0])
			number, err := strconv.ParseFloat(arg, 64)
			if err != nil {
				return nil, Error(0, fmt.Sprintf("could not convert '%s' to number", arg))
			}
			return number, nil
		},
	))

	return env
}
