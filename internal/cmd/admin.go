// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE and LICENSE.gogs file.

package cmd

import (
	"context"
	"fmt"
	"reflect"
	"runtime"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"gogs.io/gogs/internal/conf"
	"gogs.io/gogs/internal/database"
)

var (
	Admin = cli.Command{
		Name:  "admin",
		Usage: "Perform admin operations on command line",
		Description: `Allow using internal logic of Gogs without hacking into the source code
to make automatic initialization process more smoothly`,
		Subcommands: []cli.Command{
			subcmdCreateUser,
			subcmdDeleteInactivateUsers,
			subcmdDeleteRepositoryArchives,
			subcmdDeleteMissingRepositories,
			subcmdGitGcRepos,
			subcmdRewriteAuthorizedKeys,
			subcmdSyncRepositoryHooks,
			subcmdReinitMissingRepositories,
		},
	}

	subcmdCreateUser = cli.Command{
		Name:   "create-user",
		Usage:  "Create a new user in database",
		Action: runCreateUser,
		Flags: []cli.Flag{
			stringFlag("name", "", "Username"),
			stringFlag("password", "", "User password"),
			stringFlag("email", "", "User email address"),
			boolFlag("admin", "User is an admin"),
			stringFlag("config, c", "", "Custom configuration file path"),
		},
	}

	subcmdDeleteInactivateUsers = cli.Command{
		Name:  "delete-inactive-users",
		Usage: "Delete all inactive accounts",
		Action: adminDashboardOperation(
			func() error { return database.Handle.Users().DeleteInactivated() },
			"All inactivated accounts have been deleted successfully",
		),
		Flags: []cli.Flag{
			stringFlag("config, c", "", "Custom configuration file path"),
		},
	}

	subcmdDeleteRepositoryArchives = cli.Command{
		Name:  "delete-repository-archives",
		Usage: "Delete all repositories archives",
		Action: adminDashboardOperation(
			database.DeleteRepositoryArchives,
			"All repositories archives have been deleted successfully",
		),
		Flags: []cli.Flag{
			stringFlag("config, c", "", "Custom configuration file path"),
		},
	}

	subcmdDeleteMissingRepositories = cli.Command{
		Name:  "delete-missing-repositories",
		Usage: "Delete all repository records that lost Git files",
		Action: adminDashboardOperation(
			database.DeleteMissingRepositories,
			"All repositories archives have been deleted successfully",
		),
		Flags: []cli.Flag{
			stringFlag("config, c", "", "Custom configuration file path"),
		},
	}

	subcmdGitGcRepos = cli.Command{
		Name:  "collect-garbage",
		Usage: "Do garbage collection on repositories",
		Action: adminDashboardOperation(
			database.GitGcRepos,
			"All repositories have done garbage collection successfully",
		),
		Flags: []cli.Flag{
			stringFlag("config, c", "", "Custom configuration file path"),
		},
	}

	subcmdRewriteAuthorizedKeys = cli.Command{
		Name:  "rewrite-authorized-keys",
		Usage: "Rewrite '.ssh/authorized_keys' file (caution: non-Gogs keys will be lost)",
		Action: adminDashboardOperation(
			database.RewriteAuthorizedKeys,
			"All public keys have been rewritten successfully",
		),
		Flags: []cli.Flag{
			stringFlag("config, c", "", "Custom configuration file path"),
		},
	}

	subcmdSyncRepositoryHooks = cli.Command{
		Name:  "resync-hooks",
		Usage: "Resync pre-receive, update and post-receive hooks",
		Action: adminDashboardOperation(
			database.SyncRepositoryHooks,
			"All repositories' pre-receive, update and post-receive hooks have been resynced successfully",
		),
		Flags: []cli.Flag{
			stringFlag("config, c", "", "Custom configuration file path"),
		},
	}

	subcmdReinitMissingRepositories = cli.Command{
		Name:  "reinit-missing-repositories",
		Usage: "Reinitialize all repository records that lost Git files",
		Action: adminDashboardOperation(
			database.ReinitMissingRepositories,
			"All repository records that lost Git files have been reinitialized successfully",
		),
		Flags: []cli.Flag{
			stringFlag("config, c", "", "Custom configuration file path"),
		},
	}
)

func runCreateUser(c *cli.Context) error {
	if !c.IsSet("name") {
		return errors.New("Username is not specified")
	} else if !c.IsSet("password") {
		return errors.New("Password is not specified")
	} else if !c.IsSet("email") {
		return errors.New("Email is not specified")
	}

	err := conf.Init(c.String("config"))
	if err != nil {
		return errors.Wrap(err, "init configuration")
	}
	conf.InitLogging(true)

	if _, err = database.SetEngine(); err != nil {
		return errors.Wrap(err, "set engine")
	}

	user, err := database.Handle.Users().Create(context.Background(), c.String("name"), c.String("email"), database.CreateUserOptions{
		Password:  c.String("password"),
		Activated: true,
		Admin:     c.Bool("admin"),
	})
	if err != nil {
		return errors.Wrap(err, "create user")
	}

	fmt.Printf("New user %q has been successfully created!\n", user.Name)
	return nil
}

func adminDashboardOperation(operation func() error, successMessage string) func(*cli.Context) error {
	return func(c *cli.Context) error {
		err := conf.Init(c.String("config"))
		if err != nil {
			return errors.Wrap(err, "init configuration")
		}
		conf.InitLogging(true)

		if _, err = database.SetEngine(); err != nil {
			return errors.Wrap(err, "set engine")
		}

		if err := operation(); err != nil {
			functionName := runtime.FuncForPC(reflect.ValueOf(operation).Pointer()).Name()
			return fmt.Errorf("%s: %v", functionName, err)
		}

		fmt.Printf("%s\n", successMessage)
		return nil
	}
}
