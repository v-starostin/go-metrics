package main

import "os"

func main() {
	os.Exit(0) // want "direct call to os.Exit in main function of main package is not allowed"
}
