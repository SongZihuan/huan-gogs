// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE and LICENSE.gogs file.

package database

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogs/git-module"
	"github.com/pkg/errors"
)

// CommitToPushCommit transforms a git.Commit to PushCommit type.
func CommitToPushCommit(commit *git.Commit) *PushCommit {
	return &PushCommit{
		Sha1:           commit.ID.String(),
		Message:        commit.Message,
		AuthorEmail:    commit.Author.Email,
		AuthorName:     commit.Author.Name,
		CommitterEmail: commit.Committer.Email,
		CommitterName:  commit.Committer.Name,
		Timestamp:      commit.Committer.When,
	}
}

func CommitsToPushCommits(commits []*git.Commit) *PushCommits {
	if len(commits) == 0 {
		return &PushCommits{}
	}

	pcs := make([]*PushCommit, len(commits))
	for i := range commits {
		pcs[i] = CommitToPushCommit(commits[i])
	}
	return &PushCommits{len(pcs), pcs, "", nil}
}

type PushUpdateOptions struct {
	OldCommitID  string
	NewCommitID  string
	FullRefspec  string
	PusherID     int64
	PusherName   string
	RepoUserName string
	RepoName     string
}

// PushUpdate must be called for any push actions in order to generate necessary
// push action history feeds.
func PushUpdate(opts PushUpdateOptions) (err error) {
	ctx := context.TODO()

	isNewRef := strings.HasPrefix(opts.OldCommitID, git.EmptyID)
	isDelRef := strings.HasPrefix(opts.NewCommitID, git.EmptyID)
	if isNewRef && isDelRef {
		return fmt.Errorf("both old and new revisions are %q", git.EmptyID)
	}

	repoPath := RepoPath(opts.RepoUserName, opts.RepoName)

	gitUpdate := exec.Command("git", "update-server-info")
	gitUpdate.Dir = repoPath
	if err = gitUpdate.Run(); err != nil {
		return fmt.Errorf("run 'git update-server-info': %v", err)
	}

	gitRepo, err := git.Open(repoPath)
	if err != nil {
		return fmt.Errorf("open repository: %v", err)
	}

	owner, err := Handle.Users().GetByUsername(ctx, opts.RepoUserName)
	if err != nil {
		return fmt.Errorf("GetUserByName: %v", err)
	}

	repo, err := GetRepositoryByName(owner.ID, opts.RepoName)
	if err != nil {
		return fmt.Errorf("GetRepositoryByName: %v", err)
	}

	if err = repo.UpdateSize(); err != nil {
		return fmt.Errorf("UpdateSize: %v", err)
	}

	// Push tags
	if strings.HasPrefix(opts.FullRefspec, git.RefsTags) {
		err := Handle.Actions().PushTag(ctx,
			PushTagOptions{
				Owner:       owner,
				Repo:        repo,
				PusherName:  opts.PusherName,
				RefFullName: opts.FullRefspec,
				NewCommitID: opts.NewCommitID,
			},
		)
		if err != nil {
			return errors.Wrap(err, "create action for push tag")
		}
		return nil
	}

	var commits []*git.Commit
	// Skip read parent commits when delete branch
	if !isDelRef {
		// Push new branch
		newCommit, err := gitRepo.CatFileCommit(opts.NewCommitID)
		if err != nil {
			return fmt.Errorf("GetCommit [commit_id: %s]: %v", opts.NewCommitID, err)
		}

		if isNewRef {
			commits, err = newCommit.Ancestors(git.LogOptions{MaxCount: 9})
			if err != nil {
				return fmt.Errorf("CommitsBeforeLimit [commit_id: %s]: %v", newCommit.ID, err)
			}
			commits = append([]*git.Commit{newCommit}, commits...)
		} else {
			commits, err = newCommit.CommitsAfter(opts.OldCommitID)
			if err != nil {
				return fmt.Errorf("CommitsBeforeUntil [commit_id: %s]: %v", opts.OldCommitID, err)
			}
		}
	}

	err = Handle.Actions().CommitRepo(ctx,
		CommitRepoOptions{
			Owner:       owner,
			Repo:        repo,
			PusherName:  opts.PusherName,
			RefFullName: opts.FullRefspec,
			OldCommitID: opts.OldCommitID,
			NewCommitID: opts.NewCommitID,
			Commits:     CommitsToPushCommits(commits),
		},
	)
	if err != nil {
		return errors.Wrap(err, "create action for commit push")
	}
	return nil
}
