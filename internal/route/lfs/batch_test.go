// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE and LICENSE.gogs file.

package lfs

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/macaron.v1"

	"gogs.io/gogs/internal/conf"
	"gogs.io/gogs/internal/database"
)

func TestServeBatch(t *testing.T) {
	conf.SetMockServer(t, conf.ServerOpts{
		ExternalURL: "https://gogs.example.com/",
	})

	tests := []struct {
		name          string
		body          string
		mockStore     func() *MockStore
		expStatusCode int
		expBody       string
	}{
		{
			name:          "unrecognized operation",
			body:          `{"operation": "update"}`,
			expStatusCode: http.StatusBadRequest,
			expBody:       `{"message": "Operation not recognized"}` + "\n",
		},
		{
			name: "upload: contains invalid oid",
			body: `{
"operation": "upload",
"objects": [
	{"oid": "bad_oid", "size": 123},
	{"oid": "ef797c8118f02dfb649607dd5d3f8c7623048c9c063d532cc95c5ed7a898a64f", "size": 123}
]}`,
			expStatusCode: http.StatusOK,
			expBody: `{
	"transfer": "basic",
	"objects": [
		{"oid": "bad_oid", "size":123, "actions": {"error": {"code": 422, "message": "Object has invalid oid"}}},
		{
			"oid": "ef797c8118f02dfb649607dd5d3f8c7623048c9c063d532cc95c5ed7a898a64f",
			"size": 123,
			"actions": {
				"upload": {
					"href": "https://gogs.example.com/owner/repo.git/info/lfs/objects/basic/ef797c8118f02dfb649607dd5d3f8c7623048c9c063d532cc95c5ed7a898a64f",
					"header": {"Content-Type": "application/octet-stream"}
				},
				"verify": {
					"href": "https://gogs.example.com/owner/repo.git/info/lfs/objects/basic/verify"
				}
			}
		}
	]
}` + "\n",
		},
		{
			name: "download: contains non-existent oid and mismatched size",
			body: `{
"operation": "download",
"objects": [
	{"oid": "bad_oid", "size": 123},
	{"oid": "ef797c8118f02dfb649607dd5d3f8c7623048c9c063d532cc95c5ed7a898a64f", "size": 123},
	{"oid": "5cac0a318669fadfee734fb340a5f5b70b428ac57a9f4b109cb6e150b2ba7e57", "size": 456}
]}`,
			mockStore: func() *MockStore {
				mockStore := NewMockStore()
				mockStore.GetLFSObjectsByOIDsFunc.SetDefaultReturn(
					[]*database.LFSObject{
						{
							OID:  "ef797c8118f02dfb649607dd5d3f8c7623048c9c063d532cc95c5ed7a898a64f",
							Size: 1234,
						}, {
							OID:  "5cac0a318669fadfee734fb340a5f5b70b428ac57a9f4b109cb6e150b2ba7e57",
							Size: 456,
						},
					},
					nil,
				)
				return mockStore
			},
			expStatusCode: http.StatusOK,
			expBody: `{
	"transfer": "basic",
	"objects": [
		{"oid": "bad_oid", "size": 123, "actions": {"error": {"code": 404, "message": "Object does not exist"}}},
		{
			"oid": "ef797c8118f02dfb649607dd5d3f8c7623048c9c063d532cc95c5ed7a898a64f",
			"size": 123,
			"actions": {"error": {"code": 422, "message": "Object size mismatch"}}
		},
		{
			"oid": "5cac0a318669fadfee734fb340a5f5b70b428ac57a9f4b109cb6e150b2ba7e57",
			"size": 456,
			"actions": {
				"download": {
					"href": "https://gogs.example.com/owner/repo.git/info/lfs/objects/basic/5cac0a318669fadfee734fb340a5f5b70b428ac57a9f4b109cb6e150b2ba7e57"
				}
			}
		}
	]
}` + "\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := NewMockStore()
			if test.mockStore != nil {
				mockStore = test.mockStore()
			}

			m := macaron.New()
			m.Use(func(c *macaron.Context) {
				c.Map(&database.User{Name: "owner"})
				c.Map(&database.Repository{Name: "repo"})
			})
			m.Post("/", serveBatch(mockStore))

			r, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(test.body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			m.ServeHTTP(rr, r)

			resp := rr.Result()
			assert.Equal(t, test.expStatusCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var expBody bytes.Buffer
			err = json.Indent(&expBody, []byte(test.expBody), "", "  ")
			require.NoError(t, err)

			var gotBody bytes.Buffer
			err = json.Indent(&gotBody, body, "", "  ")
			require.NoError(t, err)

			assert.Equal(t, expBody.String(), gotBody.String())
		})
	}
}
