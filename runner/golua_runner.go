package runner

import (
	"lxa/state"
)

func GoRunBinary(b []byte, name string) {
	ls := state.New()
	ls.OpenLibs()
	ls.Load(b, name, "b")
	ls.Call(0, -1)
}
