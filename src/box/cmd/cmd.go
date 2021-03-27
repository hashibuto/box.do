package cmd

import (
	"bufio"
	"fmt"
	"os/exec"
)

type Command struct {
	CWD  string
	Echo bool
	Cmd  *exec.Cmd
}

// New returns a new command using the supplied arguments
func New(command ...string) (*Command, error) {
	executable := command[0]
	args := command[1:]

	return &Command{
		Cmd:  exec.Command(executable, args...),
		Echo: true,
	}, nil
}

// Run executes the command
func (c *Command) Run() error {
	var scanner *bufio.Scanner
	if c.CWD != "" {
		c.Cmd.Dir = c.CWD
	}
	if c.Echo {
		output, err := c.Cmd.StdoutPipe()
		c.Cmd.Stderr = c.Cmd.Stdout
		if err != nil {
			return fmt.Errorf("cmd.Run: %w", err)
		}

		scanner = bufio.NewScanner(output)
	}

	c.Cmd.Start()

	if c.Echo {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}

	return c.Cmd.Wait()
}
