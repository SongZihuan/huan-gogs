// Copyright 2023 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE and LICENSE.gogs file.

package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNotice_BeforeCreate(t *testing.T) {
	now := time.Now()
	db := &gorm.DB{
		Config: &gorm.Config{
			SkipDefaultTransaction: true,
			NowFunc: func() time.Time {
				return now
			},
		},
	}

	t.Run("CreatedUnix has been set", func(t *testing.T) {
		notice := &Notice{
			CreatedUnix: 1,
		}
		_ = notice.BeforeCreate(db)
		assert.Equal(t, int64(1), notice.CreatedUnix)
	})

	t.Run("CreatedUnix has not been set", func(t *testing.T) {
		notice := &Notice{}
		_ = notice.BeforeCreate(db)
		assert.Equal(t, db.NowFunc().Unix(), notice.CreatedUnix)
	})
}

func TestNotice_AfterFind(t *testing.T) {
	now := time.Now()
	db := &gorm.DB{
		Config: &gorm.Config{
			SkipDefaultTransaction: true,
			NowFunc: func() time.Time {
				return now
			},
		},
	}

	notice := &Notice{
		CreatedUnix: now.Unix(),
	}
	_ = notice.AfterFind(db)
	assert.Equal(t, notice.CreatedUnix, notice.Created.Unix())
}

func TestNotices(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	s := &NoticesStore{
		db: newTestDB(t, "NoticesStore"),
	}

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, s *NoticesStore)
	}{
		{"Create", noticesCreate},
		{"DeleteByIDs", noticesDeleteByIDs},
		{"DeleteAll", noticesDeleteAll},
		{"List", noticesList},
		{"Count", noticesCount},
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

func noticesCreate(t *testing.T, ctx context.Context, s *NoticesStore) {
	err := s.Create(ctx, NoticeTypeRepository, "test")
	require.NoError(t, err)

	count := s.Count(ctx)
	assert.Equal(t, int64(1), count)
}

func noticesDeleteByIDs(t *testing.T, ctx context.Context, s *NoticesStore) {
	err := s.Create(ctx, NoticeTypeRepository, "test")
	require.NoError(t, err)

	notices, err := s.List(ctx, 1, 10)
	require.NoError(t, err)
	ids := make([]int64, 0, len(notices))
	for _, notice := range notices {
		ids = append(ids, notice.ID)
	}

	// Non-existing IDs should be ignored
	ids = append(ids, 404)
	err = s.DeleteByIDs(ctx, ids...)
	require.NoError(t, err)

	count := s.Count(ctx)
	assert.Equal(t, int64(0), count)
}

func noticesDeleteAll(t *testing.T, ctx context.Context, s *NoticesStore) {
	err := s.Create(ctx, NoticeTypeRepository, "test")
	require.NoError(t, err)

	err = s.DeleteAll(ctx)
	require.NoError(t, err)

	count := s.Count(ctx)
	assert.Equal(t, int64(0), count)
}

func noticesList(t *testing.T, ctx context.Context, s *NoticesStore) {
	err := s.Create(ctx, NoticeTypeRepository, "test 1")
	require.NoError(t, err)
	err = s.Create(ctx, NoticeTypeRepository, "test 2")
	require.NoError(t, err)

	got1, err := s.List(ctx, 1, 1)
	require.NoError(t, err)
	require.Len(t, got1, 1)

	got2, err := s.List(ctx, 2, 1)
	require.NoError(t, err)
	require.Len(t, got2, 1)
	assert.True(t, got1[0].ID > got2[0].ID)

	got, err := s.List(ctx, 1, 3)
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func noticesCount(t *testing.T, ctx context.Context, s *NoticesStore) {
	count := s.Count(ctx)
	assert.Equal(t, int64(0), count)

	err := s.Create(ctx, NoticeTypeRepository, "test")
	require.NoError(t, err)

	count = s.Count(ctx)
	assert.Equal(t, int64(1), count)
}
