package main

import (
	"fmt"
	"lxa/golua/lua"
	"lxa/golua/std"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		// ls := state.New()
		// ls.OpenLibs()
		// ls.LoadFile(os.Args[1])
		// ls.Call(0, -1)
		var opts = []lua.Option{lua.WithTrace(false), lua.WithVerbose(false)}
		state := lua.NewState(opts...)
		defer state.Close()
		std.Open(state)

		if err := state.Main(os.Args[1:]...); err != nil {
			fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
			os.Exit(1)
		}
	}
}
