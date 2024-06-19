// Copyright (c) 2024 Michael D Henderson. All rights reserved.

//go:build darwin

package main

import (
	"fmt"
	"os"
)

func HelpConsole() {
	fmt.Printf("You must run Otto from a terminal window.\n")
	fmt.Printf("\n")
	fmt.Printf("You can open a terminal window by pressing the `Cmd + Space` keys\n")
	fmt.Printf("on your keyboard. This opens the Spotlight Search. Type `Terminal in\n")
	fmt.Printf("the dialog and press `Enter`.\n")
	fmt.Printf("\n")
	fmt.Printf("Once the terminal window is open, you can start Otto by\n")
	fmt.Printf("typing `%s` and pressing `Enter`.\n", os.Args[0])
}
