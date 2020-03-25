package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"lxa/binchunk"
	"lxa/compiler"
	"lxa/runner"
	"os"
)

var (
	CLUA  bool
	GOLUA bool
	//DEBUG   bool
	COMPILE bool
)

func init() {
	flag.BoolVar(&COMPILE, "c", false, "compile lxa file to lua bytecode")
	//flag.BoolVar(&DEBUG, "g", false, "enable verbose logging and tracing")
	flag.BoolVar(&GOLUA, "golua", false, "use inner golua vm for excuting")
	flag.BoolVar(&CLUA, "clua", false, "use inner official clua 5.3.5 vm for excuting")
	flag.Parse()
}

func main() {
	if len(os.Args) > 1 {
		for _, filename := range flag.Args() {
			chunk, err := ioutil.ReadFile(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "reading %s: %v", filename, err)
			}
			var data []byte
			if binchunk.IsBinaryChunk(chunk) {
				data = chunk
			} else {
				proto := compiler.Compile(string(chunk), filename)
				data = binchunk.Dump(proto)
			}
			if COMPILE {
				ioutil.WriteFile(filename+".luac", data, 0666)
			} else {
				if GOLUA && !CLUA {
					runner.GoRunBinary(data, filename)
				} else {
					runner.CRunBinary(data, filename)
				}
			}
		}
	} else {
		fmt.Println("Oops! No input files given.")
		fmt.Println("Lxa 0.2.1 2020.03.25 Copyright (C) 2020 xaxys.")
		fmt.Println("See more Details at github.com/xaxys/lxa.")
		fmt.Println("Avaliable options are:")
		fmt.Println("  -c    ", "Compile a lxa file to lua bytecode without running")
		//fmt.Println("  -g", "Enable verbose logging and tracing")
		fmt.Println("  -golua", "Use inner golua vm for excuting (several stdlib unsupported yet)")
		fmt.Println("  -clua ", "Use inner official clua 5.3.5 vm for excuting (Default VM)")
	}
}
