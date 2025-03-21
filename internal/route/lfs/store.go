// Copyright 2022 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE and LICENSE.gogs file.

// Copyright 2025 Huan-Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lfs

import (
	"context"

	"gogs.io/gogs/internal/database"
	"gogs.io/gogs/internal/lfsutil"
)

// Store is the data layer carrier for LFS endpoints. This interface is meant to
// abstract away and limit the exposure of the underlying data layer to the
// handler through a thin-wrapper.
type Store interface {
	// GetAccessTokenBySHA1 returns the access token with given SHA1. It returns
	// database.ErrAccessTokenNotExist when not found.
	GetAccessTokenBySHA1(ctx context.Context, sha1 string) (*database.AccessToken, error)
	// TouchAccessTokenByID updates the updated time of the given access token to
	// the current time.
	TouchAccessTokenByID(ctx context.Context, id int64) error

	// CreateLFSObject creates an LFS object record in database.
	CreateLFSObject(ctx context.Context, repoID int64, oid lfsutil.OID, size int64, storage lfsutil.Storage) error
	// GetLFSObjectByOID returns the LFS object with given OID. It returns
	// database.ErrLFSObjectNotExist when not found.
	GetLFSObjectByOID(ctx context.Context, repoID int64, oid lfsutil.OID) (*database.LFSObject, error)
	// GetLFSObjectsByOIDs returns LFS objects found within "oids". The returned
	// list could have fewer elements if some oids were not found.
	GetLFSObjectsByOIDs(ctx context.Context, repoID int64, oids ...lfsutil.OID) ([]*database.LFSObject, error)

	// AuthorizeRepositoryAccess returns true if the user has as good as desired
	// access mode to the repository.
	AuthorizeRepositoryAccess(ctx context.Context, userID, repoID int64, desired database.AccessMode, opts database.AccessModeOptions) bool

	// GetRepositoryByName returns the repository with given owner and name. It
	// returns database.ErrRepoNotExist when not found.
	GetRepositoryByName(ctx context.Context, ownerID int64, name string) (*database.Repository, error)

	// IsTwoFactorEnabled returns true if the user has enabled 2FA.
	IsTwoFactorEnabled(ctx context.Context, userID int64) bool

	// GetUserByID returns the user with given ID. It returns
	// database.ErrUserNotExist when not found.
	GetUserByID(ctx context.Context, id int64) (*database.User, error)
	// GetUserByUsername returns the user with given username. It returns
	// database.ErrUserNotExist when not found.
	GetUserByUsername(ctx context.Context, username string) (*database.User, error)
	// CreateUser creates a new user and persists to database. It returns
	// database.ErrNameNotAllowed if the given name or pattern of the name is not
	// allowed as a username, or database.ErrUserAlreadyExist when a user with same
	// name already exists, or database.ErrEmailAlreadyUsed if the email has been
	// verified by another user.
	CreateUser(ctx context.Context, username, email, publicEmail string, opts database.CreateUserOptions) (*database.User, error)
	// AuthenticateUser validates username and password via given login source ID.
	// It returns database.ErrUserNotExist when the user was not found.
	//
	// When the "loginSourceID" is negative, it aborts the process and returns
	// database.ErrUserNotExist if the user was not found in the database.
	//
	// When the "loginSourceID" is non-negative, it returns
	// database.ErrLoginSourceMismatch if the user has different login source ID
	// than the "loginSourceID".
	//
	// When the "loginSourceID" is positive, it tries to authenticate via given
	// login source and creates a new user when not yet exists in the database.
	AuthenticateUser(ctx context.Context, login, password string, loginSourceID int64) (*database.User, error)
}

type store struct{}

// NewStore returns a new Store using the global database handle.
func NewStore() Store {
	return &store{}
}

func (*store) GetAccessTokenBySHA1(ctx context.Context, sha1 string) (*database.AccessToken, error) {
	return database.Handle.AccessTokens().GetBySHA1(ctx, sha1)
}

func (*store) TouchAccessTokenByID(ctx context.Context, id int64) error {
	return database.Handle.AccessTokens().Touch(ctx, id)
}

func (*store) CreateLFSObject(ctx context.Context, repoID int64, oid lfsutil.OID, size int64, storage lfsutil.Storage) error {
	return database.Handle.LFS().CreateObject(ctx, repoID, oid, size, storage)
}

func (*store) GetLFSObjectByOID(ctx context.Context, repoID int64, oid lfsutil.OID) (*database.LFSObject, error) {
	return database.Handle.LFS().GetObjectByOID(ctx, repoID, oid)
}

func (*store) GetLFSObjectsByOIDs(ctx context.Context, repoID int64, oids ...lfsutil.OID) ([]*database.LFSObject, error) {
	return database.Handle.LFS().GetObjectsByOIDs(ctx, repoID, oids...)
}

func (*store) AuthorizeRepositoryAccess(ctx context.Context, userID, repoID int64, desired database.AccessMode, opts database.AccessModeOptions) bool {
	return database.Handle.Permissions().Authorize(ctx, userID, repoID, desired, opts)
}

func (*store) GetRepositoryByName(ctx context.Context, ownerID int64, name string) (*database.Repository, error) {
	return database.Handle.Repositories().GetByName(ctx, ownerID, name)
}

func (*store) IsTwoFactorEnabled(ctx context.Context, userID int64) bool {
	return database.Handle.TwoFactors().IsEnabled(ctx, userID)
}

func (*store) GetUserByID(ctx context.Context, id int64) (*database.User, error) {
	return database.Handle.Users().GetByID(ctx, id)
}

func (*store) GetUserByUsername(ctx context.Context, username string) (*database.User, error) {
	return database.Handle.Users().GetByUsername(ctx, username)
}

func (*store) CreateUser(ctx context.Context, username, email, publicEmail string, opts database.CreateUserOptions) (*database.User, error) {
	return database.Handle.Users().Create(ctx, username, email, opts)
}

func (*store) AuthenticateUser(ctx context.Context, login, password string, loginSourceID int64) (*database.User, error) {
	return database.Handle.Users().Authenticate(ctx, login, password, loginSourceID)
}
