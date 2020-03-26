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
	CLUA     bool
	GOLUA    bool
	DEBUG    bool
	PARSE    bool
	COMPILE  bool
	PROGNAME string
)

func init() {
	PROGNAME = os.Args[0]
	flag.BoolVar(&COMPILE, "c", false, "compile lxa file to lua bytecode")
	flag.BoolVar(&DEBUG, "g", false, "enable verbose logging and tracing")
	flag.BoolVar(&PARSE, "p", false, "parse and print lua bytecode only")
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
			if PARSE {
				if binchunk.IsBinaryChunk(chunk) {
					runner.ParseBinary(chunk)
				} else {
					fmt.Fprintf(os.Stderr, "parsinging %s: %s", filename, "is not a lua bytecode file")
				}
				continue
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
				if GOLUA && !CLUA || DEBUG {
					runner.GoRunBinary(data, filename, DEBUG)
				} else {
					runner.CRunBinary(data, filename)
				}
			}
		}
	} else {
		fmt.Println("Oops! No input files given.")
		fmt.Println("Lxa 0.2.4 2020.03.26 Copyright (C) 2020 xaxys.")
		fmt.Println("usage:", PROGNAME, "[options] [script]")
		fmt.Println("avaliable options are:")
		fmt.Println("  -c    ", "Compile a lxa file to lua bytecode without running")
		fmt.Println("  -g    ", "Enable verbose logging and tracing (golua vm only)")
		fmt.Println("  -p    ", "Parse and Print lua bytecode without running")
		fmt.Println("  -golua", "Use inner golua vm for excuting (several stdlib unsupported yet)")
		fmt.Println("  -clua ", "Use inner official clua 5.3.5 vm for excuting (Default VM)")
	}
}
