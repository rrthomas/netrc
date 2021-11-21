package main

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/rrthomas/go-netrc/netrc"
	"github.com/urfave/cli/v2"
)

const (
	ErrInvalidCommand = iota
	ErrInvalidNetrc
	ErrMachineNotFound
)

func main() {
	app := &cli.App{
		Name: "netrc",
		Usage: "Manage your netrc file.",
		Authors: []*cli.Author{
			&cli.Author{
				Name: "Naaman Newbold",
				Email: "naaman@heroku.com",
			},
		},
		Version: "0.0.2",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "netrc-path",
				Value: defaultNetrc(),
				Usage: "Path to the netrc file",
			},
			&cli.BoolFlag{
				Name:  "no-machine",
				Aliases: []string{"n"},
				Usage: "disable display of machine values",
			},
			&cli.BoolFlag{
				Name:  "login",
				Aliases: []string{"l"},
				Usage: "toggle display of login values",
			},
			&cli.BoolFlag{
				Name:  "password",
				Aliases: []string{"p"},
				Usage: "toggle display of password values",
			},
		},
		Action: listCommand,
	}
	app.CommandNotFound = func(c *cli.Context, cmd string) {
		exit(c, ErrInvalidCommand)
	}
	app.BashComplete = machineCompletion

	app.EnableBashCompletion = true
	app.Run(os.Args)
}

type formattableMachine struct {
	*netrc.Machine
	*machineFormat
}

type machineFormat struct {
	showMachine  bool
	showLogin    bool
	showPassword bool
}

func (m formattableMachine) Print() {
	var v []string

	if m.showMachine {
		v = append(v, m.Machine.Name)
	}
	if m.showLogin {
		v = append(v, m.Machine.Login)
	}
	if m.showPassword {
		v = append(v, m.Machine.Password)
	}

	fmt.Print(strings.Join(v, " "))
}

func listMachines(c *cli.Context, mf *machineFormat) {
	filter, netrcFile := cmdSetup(c)
	n, err := netrc.ParseFile(netrcFile)
	if err != nil {
		exit(c, ErrInvalidNetrc)
	}

	printMachines(filterMachines(n.Machines, filter), mf)
}

func listCommand(c *cli.Context) error {
	mf := &machineFormat{
		showMachine:  !c.Bool("no-machine"),
		showLogin:    c.Bool("login"),
		showPassword: c.Bool("password"),
	}
	listMachines(c, mf)
	return nil
}

func machineCompletion(c *cli.Context) {
	mf := &machineFormat{
		showMachine:  true,
		showLogin:    false,
		showPassword: false,
	}
	listMachines(c, mf)
}

func cmdSetup(c *cli.Context) (filter, netrcPath string) {
	return c.Args().First(), c.String("netrc-path")
}

func printMachines(machines []*netrc.Machine, mf *machineFormat) {
	printNewLine := false

	for _, m := range machines {
		if printNewLine {
			fmt.Println()
		}
		printNewLine = true

		formattableMachine{m, mf}.Print()
	}
}

func filterMachines(machines []*netrc.Machine, filter string) []*netrc.Machine {
	var filteredMachines []*netrc.Machine
	if filter == "" {
		return machines
	}
	for _, m := range machines {
		if m.Name == filter {
			filteredMachines = append(filteredMachines, m)
		}
	}
	return filteredMachines
}

func defaultNetrc() string {
	if u, err := user.Current(); err == nil {
		netrcPath := u.HomeDir + "/.netrc"
		if _, err := os.Stat(netrcPath); err == nil {
			return netrcPath
		}
	}
	return ""
}

func exit(c *cli.Context, e int) {
	cli.ShowAppHelp(c)
	os.Exit(e)
}
