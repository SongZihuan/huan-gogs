// Copyright 2023 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE and LICENSE.gogs file.

package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gogs.io/gogs/internal/conf"
)

func TestPublicKeys(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	s := &PublicKeysStore{
		db: newTestDB(t, "PublicKeysStore"),
	}

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, s *PublicKeysStore)
	}{
		{"RewriteAuthorizedKeys", publicKeysRewriteAuthorizedKeys},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := clearTables(t, s.db)
				require.NoError(t, err)
			})
			tc.test(t, ctx, s)
		})
		if t.Failed() {
			break
		}
	}
}

func publicKeysRewriteAuthorizedKeys(t *testing.T, ctx context.Context, s *PublicKeysStore) {
	// TODO: Use PublicKeys.Add to replace SQL hack when the method is available.
	publicKey := &PublicKey{
		OwnerID:     1,
		Name:        "test-key",
		Fingerprint: "12:f8:7e:78:61:b4:bf:e2:de:24:15:96:4e:d4:72:53",
		Content:     "test-key-content",
	}
	err := s.db.Create(publicKey).Error
	require.NoError(t, err)
	tempSSHRootPath := filepath.Join(os.TempDir(), "publicKeysRewriteAuthorizedKeys-tempSSHRootPath")
	conf.SetMockSSH(t, conf.SSHOpts{RootPath: tempSSHRootPath})
	err = s.RewriteAuthorizedKeys()
	require.NoError(t, err)

	authorizedKeys, err := os.ReadFile(authorizedKeysPath())
	require.NoError(t, err)
	assert.Contains(t, string(authorizedKeys), fmt.Sprintf("key-%d", publicKey.ID))
	assert.Contains(t, string(authorizedKeys), publicKey.Content)
}
