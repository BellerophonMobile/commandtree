package commandtree

import (
	"fmt"
)

type DuplicateCommandError struct {
	Command string
}

func (x DuplicateCommandError) Error() string {
	return fmt.Sprintf("Command '%v' defined more than once", x.Command)
}

type NoSuchCommandError struct {
	Command string
}

func (x NoSuchCommandError) Error() string {
	return fmt.Sprintf("No such command '%v'", x.Command)
}
