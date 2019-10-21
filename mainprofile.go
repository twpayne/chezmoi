//+build profile

package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/twpayne/chezmoi/cmd"
)

func main() {
	cpuProfileFile, err := os.Create("chezmoi.prof")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer cpuProfileFile.Close()

	pprof.StartCPUProfile(cpuProfileFile)
	// FIXME pprof.StopCPUProfile is not run if os.Exit is called
	defer pprof.StopCPUProfile()

	cmd.Execute()
}
