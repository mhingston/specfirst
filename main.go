package main

import "specfirst/cmd"

var version = "0.3.1"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
