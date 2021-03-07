package main

type InitCmd struct {
	Name string `arg help="Project name"`
}

func (cmd *InitCmd) Run() error {

}
