package flag

import (
	"fmt"
	"github.com/lemontvio/rtsp2img/version"
)

type Options struct {
	Config  string
	Damon   bool
	Help    bool
	Version bool
	Done    bool
}

func Help() {
	Usage()
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println(`  -h,-?        : this is help`)
	fmt.Println(`  -v           : show version and exit`)
	fmt.Println(`  -c config    : set configuration file`)
}

func Usage() {
	fmt.Printf("Usage: %v [-?hv] [-c config]\n", version.AppName)
}

func Version() {
	fmt.Printf("%v version: %v\n", version.AppName, version.Version)
}
