package runner

import (
	"lxa/state"
)

func GoRunBinary(b []byte, name string, debug bool) {
	ls := state.NewState(debug)
	ls.OpenLibs()
	ls.Load(b, name, "b")
	ls.Call(0, -1)
}
