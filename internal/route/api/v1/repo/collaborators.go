// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE and LICENSE.gogs file.

package repo

import (
	"net/http"

	api "github.com/gogs/go-gogs-client"

	"gogs.io/gogs/internal/context"
	"gogs.io/gogs/internal/database"
)

func ListCollaborators(c *context.APIContext) {
	collaborators, err := c.Repo.Repository.GetCollaborators()
	if err != nil {
		c.Error(err, "get collaborators")
		return
	}

	apiCollaborators := make([]*api.Collaborator, len(collaborators))
	for i := range collaborators {
		apiCollaborators[i] = collaborators[i].APIFormat()
	}
	c.JSONSuccess(&apiCollaborators)
}

func AddCollaborator(c *context.APIContext, form api.AddCollaboratorOption) {
	collaborator, err := database.Handle.Users().GetByUsername(c.Req.Context(), c.Params(":collaborator"))
	if err != nil {
		if database.IsErrUserNotExist(err) {
			c.Status(http.StatusUnprocessableEntity)
		} else {
			c.Error(err, "get user by name")
		}
		return
	}

	if err := c.Repo.Repository.AddCollaborator(collaborator); err != nil {
		c.Error(err, "add collaborator")
		return
	}

	if form.Permission != nil {
		if err := c.Repo.Repository.ChangeCollaborationAccessMode(collaborator.ID, database.ParseAccessMode(*form.Permission)); err != nil {
			c.Error(err, "change collaboration access mode")
			return
		}
	}

	c.NoContent()
}

func IsCollaborator(c *context.APIContext) {
	collaborator, err := database.Handle.Users().GetByUsername(c.Req.Context(), c.Params(":collaborator"))
	if err != nil {
		if database.IsErrUserNotExist(err) {
			c.Status(http.StatusUnprocessableEntity)
		} else {
			c.Error(err, "get user by name")
		}
		return
	}

	if !c.Repo.Repository.IsCollaborator(collaborator.ID) {
		c.NotFound()
	} else {
		c.NoContent()
	}
}

func DeleteCollaborator(c *context.APIContext) {
	collaborator, err := database.Handle.Users().GetByUsername(c.Req.Context(), c.Params(":collaborator"))
	if err != nil {
		if database.IsErrUserNotExist(err) {
			c.Status(http.StatusUnprocessableEntity)
		} else {
			c.Error(err, "get user by name")
		}
		return
	}

	if err := c.Repo.Repository.DeleteCollaboration(collaborator.ID); err != nil {
		c.Error(err, "delete collaboration")
		return
	}

	c.NoContent()
}
