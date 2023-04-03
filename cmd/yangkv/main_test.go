package main

import "testing"

func TestMain(t *testing.T) {
	err := assemblyComponents()
	if err != nil {
		t.Fatalf("AssemblyComponents=[%v], want=nil", err)
	}
}
