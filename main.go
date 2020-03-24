package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"lxa/binchunk"
	"lxa/compiler"
	"lxa/golua/lua"
	"lxa/golua/std"
	"os"
)

var (
	DEBUG   bool
	COMPILE bool
)

func init() {
	flag.BoolVar(&COMPILE, "c", false, "compile .lxa to lua bytecode")
	flag.BoolVar(&DEBUG, "debug", false, "enable verbose logging and tracing")
	flag.Parse()
}

func main() {
	if len(os.Args) > 1 {
		if !COMPILE {
			// ls := state.New()
			// ls.OpenLibs()
			// ls.LoadFile(os.Args[1])
			// ls.Call(0, -1)
			var opts = []lua.Option{lua.WithTrace(DEBUG), lua.WithVerbose(DEBUG)}
			state := lua.NewState(opts...)
			defer state.Close()
			std.Open(state)
			if err := state.Main(flag.Args()...); err != nil {
				fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
				os.Exit(1)
			}
		} else {
			for _, filename := range flag.Args() {
				chunk, err := ioutil.ReadFile(filename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "reading %s: %v", filename, err)
				}
				proto := compiler.Compile(string(chunk), filename)
				data := binchunk.Dump(proto)
				ioutil.WriteFile(filename+".luac", data, 0666)
			}
		}
	}
}
