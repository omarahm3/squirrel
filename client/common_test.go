package client

import (
	"flag"
	"io"
	"os"
)

// Globals that needs to be resetter on each round of test
func ResetTesting(oldArgs []string) {
  if oldArgs == nil {
    oldArgs = []string{"squirrel"}
  }
  
	// Reset args
	os.Args = oldArgs

	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
}
