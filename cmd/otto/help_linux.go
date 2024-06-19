// Copyright (c) 2024 Michael D Henderson. All rights reserved.

//go:build linux

package main

import (
	"fmt"
	"os"
)

func HelpConsole() {
	fmt.Printf("You must run Otto from a terminal window.\n")
	fmt.Printf("\n")
	fmt.Printf("You can open a terminal window by pressing the `Ctrl + Alt + T` keys\n")
	fmt.Printf("on your keyboard.\n")
	fmt.Printf("\n")
	fmt.Printf("Once the terminal window isis open, you can start Otto\n")
	fmt.Printf("by typing `%s` and pressing `Enter`.\n", os.Args[0])
}
