package cmd

import "fmt"

type VersionCmd struct{}

func (c *VersionCmd) Run(app *App) error {
	fmt.Println("linear", app.Version)
	return nil
}
