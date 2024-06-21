// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package admin

type Shell struct {
	Title                string
	DesktopSidebarStatic DesktopSidebarStatic
	MobileMenuOffCanvas  MobileMenuOffCanvas
	NavBarSecondary      NavBarSecondary
	SearchHeaderSticky   SearchHeaderSticky

	Content         any
	AccountSettings *AccountSettings
}

type AccountSettings struct{}
type DesktopSidebarStatic struct{}
type MobileMenuOffCanvas struct{}
type NavBarSecondary struct{}
type SearchHeaderSticky struct{}

type Blank struct{}
