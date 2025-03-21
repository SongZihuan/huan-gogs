// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE and LICENSE.gogs file.

package convert

import (
	"context"
	"fmt"

	"github.com/unknwon/com"

	"github.com/gogs/git-module"
	api "github.com/gogs/go-gogs-client"

	"gogs.io/gogs/internal/database"
)

func ToEmail(email *database.EmailAddress) *api.Email {
	return &api.Email{
		Email:    email.Email,
		Verified: email.IsActivated,
		Primary:  email.IsPrimary,
	}
}

func ToBranch(b *database.Branch, c *git.Commit) *api.Branch {
	return &api.Branch{
		Name:   b.Name,
		Commit: ToCommit(c),
	}
}

type Tag struct {
	Name   string             `json:"name"`
	Commit *api.PayloadCommit `json:"commit"`
}

func ToTag(b *database.Tag, c *git.Commit) *Tag {
	return &Tag{
		Name:   b.Name,
		Commit: ToCommit(c),
	}
}

func ToCommit(c *git.Commit) *api.PayloadCommit {
	authorUsername := ""
	author, err := database.Handle.Users().GetByEmail(context.TODO(), c.Author.Email)
	if err == nil {
		authorUsername = author.Name
	}
	committerUsername := ""
	committer, err := database.Handle.Users().GetByEmail(context.TODO(), c.Committer.Email)
	if err == nil {
		committerUsername = committer.Name
	}
	return &api.PayloadCommit{
		ID:      c.ID.String(),
		Message: c.Message,
		URL:     "Not implemented",
		Author: &api.PayloadUser{
			Name:     c.Author.Name,
			Email:    c.Author.Email,
			UserName: authorUsername,
		},
		Committer: &api.PayloadUser{
			Name:     c.Committer.Name,
			Email:    c.Committer.Email,
			UserName: committerUsername,
		},
		Timestamp: c.Author.When,
	}
}

func ToPublicKey(apiLink string, key *database.PublicKey) *api.PublicKey {
	return &api.PublicKey{
		ID:      key.ID,
		Key:     key.Content,
		URL:     apiLink + com.ToStr(key.ID),
		Title:   key.Name,
		Created: key.Created,
	}
}

func ToHook(repoLink string, w *database.Webhook) *api.Hook {
	config := map[string]string{
		"url":          w.URL,
		"content_type": w.ContentType.Name(),
	}
	if w.HookTaskType == database.SLACK {
		s := w.SlackMeta()
		config["channel"] = s.Channel
		config["username"] = s.Username
		config["icon_url"] = s.IconURL
		config["color"] = s.Color
	}

	return &api.Hook{
		ID:      w.ID,
		Type:    w.HookTaskType.Name(),
		URL:     fmt.Sprintf("%s/settings/hooks/%d", repoLink, w.ID),
		Active:  w.IsActive,
		Config:  config,
		Events:  w.EventsArray(),
		Updated: w.Updated,
		Created: w.Created,
	}
}

func ToDeployKey(apiLink string, key *database.DeployKey) *api.DeployKey {
	return &api.DeployKey{
		ID:       key.ID,
		Key:      key.Content,
		URL:      apiLink + com.ToStr(key.ID),
		Title:    key.Name,
		Created:  key.Created,
		ReadOnly: true, // All deploy keys are read-only.
	}
}

func ToOrganization(org *database.User) *api.Organization {
	return &api.Organization{
		ID:          org.ID,
		AvatarUrl:   org.AvatarURL(),
		UserName:    org.Name,
		FullName:    org.FullName,
		Description: org.Description,
		Website:     org.Website,
		Location:    org.Location,
	}
}

func ToTeam(team *database.Team) *api.Team {
	return &api.Team{
		ID:          team.ID,
		Name:        team.Name,
		Description: team.Description,
		Permission:  team.Authorize.String(),
	}
}
