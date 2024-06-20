// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package authn

type Page struct {
	Title   string
	Content struct {
		DevMode  bool
		Handle   string
		Password string
	}
}
