// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package hero

type Page struct {
	Title   string
	NavBar  NavBar
	Content any
}

type NavBar struct {
	PageName string
}
