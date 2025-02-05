package u

import "github.com/abiosoft/ishell/v2"

func EchoCmd(c *ishell.Context) {
	c.Println("RawArgs:", c.RawArgs)
	c.Println("Args:", c.Args)
}

func AddCmd(sh *ishell.Shell, name string) {
	switch name {
	case "echo":
		sh.AddCmd(&ishell.Cmd{
			Name: ".echo",
			Help: "echo the input",
			Func: EchoCmd,
		})

	case "sql":
		sh.AddCmd(&ishell.Cmd{
			Name: ".sql",
			Help: "execute sql",
			Func: SqlCmd,
		})

	default:
		sh.Println("Unknown command", name)
	}
}
