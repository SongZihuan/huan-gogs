// Copyright 2018 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE and LICENSE.gogs file.

package context

import (
	"gopkg.in/macaron.v1"

	"gogs.io/gogs/internal/database"
)

// ParamsUser is the wrapper type of the target user defined by URL parameter, namely ':username'.
type ParamsUser struct {
	*database.User
}

// InjectParamsUser returns a handler that retrieves target user based on URL parameter ':username',
// and injects it as *ParamsUser.
func InjectParamsUser() macaron.Handler {
	return func(c *Context) {
		user, err := database.Handle.Users().GetByUsername(c.Req.Context(), c.Params(":username"))
		if err != nil {
			c.NotFoundOrError(err, "get user by name")
			return
		}
		c.Map(&ParamsUser{user})
	}
}
