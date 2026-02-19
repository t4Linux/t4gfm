package main

import (
	"fmt"

	variable "github.com/t4Linux/t4gfm/src/config"
	"github.com/t4Linux/t4gfm/src/internal/common"
)

func main() {
	fmt.Println("ConfigFile:", variable.ConfigFile)
	fmt.Println("HotkeysFile:", variable.HotkeysFile)
	common.LoadConfigFile()
	common.LoadHotkeysFile(false)
	fmt.Println("OK")
}
