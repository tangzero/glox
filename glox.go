// Package glox is A Go implementation of the Lox programming language from "Crafting Interpreters"
// by Robert Nystrom (http://www.craftinginterpreters.com/).
package glox

import (
	"bufio"
	"fmt"
	"os"
)

func RunFile(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read file: %v", err)
	}
	return Run(string(bytes))
}

func RunPrompt() error {
	fmt.Println("Glox REPL. Press Ctrl+C to exit.")
	prompt := "> "
	scanner := bufio.NewScanner(os.Stdin)
	for fmt.Print(prompt); scanner.Scan(); fmt.Print(prompt) {
		if err := Run(scanner.Text()); err != nil {
			fmt.Println(err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}
	return nil
}

func Run(source string) error {
	scanner := NewScanner(source)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		return err
	}
	program, err := NewParser[any](tokens).Parse()
	if err != nil {
		return err
	}
	return new(Interpreter).Interpret(program)
}
