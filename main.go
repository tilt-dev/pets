package main

import (
	"fmt"
	"os"

	"github.com/windmilleng/pets/cmd/pets"
)

func main() {
	if err := pets.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
