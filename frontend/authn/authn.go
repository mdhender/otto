// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package authn

type Page struct {
	Title   string
	Content any
}

type Login struct {
	DevMode  bool
	Handle   string
	Password string
}

type SignUp struct {
	InviteLink string
	Handle     string
	Password   string
}
