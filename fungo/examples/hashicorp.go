package main

import (
	"github.com/paololu/funplugin/fungo"
)

// register functions and build to plugin binary
func main() {
	fungo.Register("sum_ints", SumInts)
	fungo.Register("sum_two_int", SumTwoInt)
	fungo.Register("sum", Sum)
	fungo.Register("sum_two_string", SumTwoString)
	fungo.Register("sum_strings", SumStrings)
	fungo.Register("concatenate", Concatenate)
	fungo.Register("setup_hook_example", SetupHookExample)
	fungo.Register("teardown_hook_example", TeardownHookExample)

	// if you want to run golang plugin over net/rpc, comment out the following line
	// os.Setenv("HRP_PLUGIN_TYPE", "rpc")
	fungo.Serve()
}
