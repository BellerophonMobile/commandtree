package commandtree

import (
	"fmt"
	"bytes"
	"github.com/BellerophonMobile/qsplit"
)

type CommandMap map[string]*Command

type parameter interface {
	setdefault()
	parse(words []string) ([]string,error)
	usagetype() string
	usagedefault() string
	usagedescription() string
}

type Variables struct {
	parameters map[string]parameter
}

func (x *Variables) setdefaults() {
	for _,p := range(x.parameters) {
		p.setdefault()
	}
}

func Parameters() *Variables {
	return &Variables{
		parameters: make(map[string]parameter),
	}
}

func (x *Variables) String(label, defaultval, description string, ref *string) *Variables {
	x.parameters[label] = stringparameter{label, defaultval, description, ref}
	return x
}

type stringparameter struct {
	label string
	defaultval string
	description string
	dest *string
}

func (x stringparameter) parse(words []string) ([]string,error) {
	if len(words) <= 0 {
		return nil,MissingValueError{x.label}
	}
	(*x.dest) = words[0]
	return words[1:],nil
}

func (x stringparameter) setdefault() {
	(*x.dest) = x.defaultval
}

func (x stringparameter) usagetype() string {
	return "string"
}
func (x stringparameter) usagedefault() string {
	return x.defaultval
}
func (x stringparameter) usagedescription() string {
	return x.description
}

type MissingValueError struct {
	label string
}

func (x MissingValueError) Error() string {
	return fmt.Sprintf("Missing value for parameter %v", x.label)
}

type Action func([]string) error

type Command struct {
	Command string
	Parameters *Variables
	Action Action
	Description string
	Usage string
	subcommands CommandMap
}

type CommandTree struct {
	Commands CommandMap
}

func New() *CommandTree {
	return &CommandTree{
		Commands: make(CommandMap),
	}
}

func (x *CommandTree) Add(command *Command) error {

	if x.Commands == nil {
		x.Commands = make(CommandMap)
	}

	_,ok := x.Commands[command.Command]
	if ok {
		return DuplicateCommandError{command.Command}
	}
	
	x.Commands[command.Command] = command

	return nil

}

func (x *CommandTree) Usage() string {
	var buff = &bytes.Buffer{}
	fmt.Fprintf(buff, "   %-12v   %v\n", "Command", "Description")
	fmt.Fprintf(buff, "   ------------   ------------------------------------------------\n")
	for _,cmd := range(x.Commands) {
		fmt.Fprintf(buff, "   %-12v   %v\n", cmd.Command, cmd.Description)
	}
	return buff.String()
}

func (x *CommandTree) Help(words []string) (string,error) {

	cmd,_,err := x.findcommand(words, true)
	if err != nil {
		return "",err
	}

	var buff = &bytes.Buffer{}
	fmt.Fprintf(buff, "%v", cmd.Command)
	fmt.Fprintf(buff, "\n   %v", cmd.Description)
	if cmd.Usage != "" {
		fmt.Fprintf(buff, "\n   %v", cmd.Usage)
	}
	
	if cmd.Parameters == nil || len(cmd.Parameters.parameters) <= 0 {
		fmt.Fprintf(buff, "\n\n      <no named parameters>")
	} else {
		fmt.Fprintf(buff, "\n\n      %-12v   %-8v   %-24v   %v", "Parameter", "Type", "Default", "Description")
		fmt.Fprintf(buff, "\n      ------------   --------   ------------------------   ------------------")
		for v,p := range(cmd.Parameters.parameters) {
			fmt.Fprintf(buff, "\n      %-12v   %-8v   %-24v   %v", v, p.usagetype(), p.usagedefault(), p.usagedescription())
		}
	}

	if cmd.subcommands == nil || len(cmd.subcommands) <= 0 {
		fmt.Fprintf(buff, "\n\n      <no subcommands>")
	} else {
		fmt.Fprintf(buff, "\n\n      %-12v   %v", "Subcommand", "Description")
		fmt.Fprintf(buff, "\n      ------------   ------------------")
		for s,c := range(cmd.subcommands) {
			fmt.Fprintf(buff, "\n      %-12v   %v", s, c.Description)
		}
	}
	
	return buff.String(),nil	

}

func (x *CommandTree) findcommand(words []string, strict bool) (*Command,[]string,error) {

	var cmd *Command
	menu := x.Commands
	
	for len(words) > 0 && len(menu) > 0 {
		c,ok := menu[words[0]]
		if !ok {
			if strict {
				return nil,nil,NoSuchCommandError{words[0]}				
			}
			break
		}
		cmd = c
		menu = cmd.subcommands
		words = words[1:]
	}

	if cmd == nil {
		return nil,nil,NoSuchCommandError{words[0]}
	}

	return cmd,words,nil
	
}

func (x *CommandTree) Execute(line string) error {

	words,err := qsplit.Split(line)
	if err != nil {
		return err
	}

	if len(words) <= 0 {
		return nil
	}

	return x.ExecuteWords(words)
	
}

func (x *CommandTree) ExecuteWords(words []string) error {

	cmd,args,err := x.findcommand(words, false)
	if err != nil {
		return err
	}

	if cmd.Parameters != nil {
		cmd.Parameters.setdefaults()

		for len(args) > 0 {
			p := args[0]
			v,ok := cmd.Parameters.parameters[p]
			if !ok {
				break
			}
			args,err = v.parse(args[1:])
			if err != nil {
				return err
			}
		}
	}

	if cmd.Action == nil {
		if len(args) > 0 {
			return NoSuchCommandError{args[0]}
		}

		return nil
	}
	
	return cmd.Action(args)

}

func (x *Command) Add(command *Command) error {

	if x.subcommands == nil {
		x.subcommands = make(CommandMap)
	}

	_,ok := x.subcommands[command.Command]
	if ok {
		return DuplicateCommandError{command.Command}
	}
	
	x.subcommands[command.Command] = command

	return nil

}
