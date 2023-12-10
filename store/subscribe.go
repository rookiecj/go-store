package store

import "fmt"

func DumpState[S State](state S, old S, action Action) {
	fmt.Printf("stat: %v <- %v, action: %v\n", state, old, action)
}
