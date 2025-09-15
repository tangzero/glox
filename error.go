package glox

import "fmt"

func Report(line int, where string, message string) {
	fmt.Printf("[line %d] Error%s: %s\n", line, where, message)
}

func Error(line int, message string) {
	Report(line, "", message)
}
