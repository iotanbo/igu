package main

import (
	"fmt"
	"os"

	ec "github.com/iotanbo/igu/pkg/ec"
	ecfs "github.com/iotanbo/igu/pkg/ecfs"
	"github.com/iotanbo/igu/pkg/fu"
)

func main() {
	fmt.Println("Iotanbo Go Utils v.0.0.1")
	_, e := fu.FileExists("test")
	if e.Some() { // there was some kind of error
		// Switch error kind using if-else statements
		if e.Eq(ec.Type) {
			fmt.Println("Specified path is not a file")
		} else if e.Eq(ec.NotFound) {
			fmt.Println("Specified path does not exist")
		} else {
			fmt.Printf("%v\n", e)
		}
		// Alternatively switch error code directly (better readability)
		switch e.Code {
		case ec.NotFound:
			fmt.Println("Specified path does not exist")
		case ec.Type:
			fmt.Println("Specified path is not a file")
		case ec.NotImplemented:
		case ec.AlreadyExists:
		case ecfs.FileCorrupt:
		}
		os.Exit(0)
	}

	fmt.Println("Test was successful.")
}
