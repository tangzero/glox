package glox

import "fmt"

func Report(line int, where string, message string) error {
	return fmt.Errorf("[line %d] Error%s: %s", line, where, message)
}

func Error(line int, message string) error {
	return Report(line, "", message)
}
