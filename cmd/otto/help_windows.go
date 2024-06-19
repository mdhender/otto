// Copyright (c) 2024 Michael D Henderson. All rights reserved.

//go:build windows

package main

import (
	"fmt"
	"os"
)

func HelpConsole() {
	fmt.Printf("You must run Otto from a command prompt.\n")
	fmt.Printf("\n")
	fmt.Printf("You can open a command prompt by pressing the `Win + R` keys\n")
	fmt.Printf("on your keyboard. This opens the Run dialog. Type `cmd` in the\n")
	fmt.Printf("dialog and pressing `Enter`.\n")
	fmt.Printf("\n")
	fmt.Printf("Once the command prompt is open, you can start Otto by\n")
	fmt.Printf("typing `%s` and pressing `Enter`.\n", os.Args[0])
}
