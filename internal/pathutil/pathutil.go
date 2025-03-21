// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE and LICENSE.gogs file.

package pathutil

import (
	"path"
	"strings"
)

// Clean cleans up given path and returns a relative path that goes straight
// down to prevent path traversal.
//
// 🚨 SECURITY: This function MUST be used for any user input that is used as
// file system path to prevent path traversal.
func Clean(p string) string {
	p = strings.ReplaceAll(p, `\`, "/")
	return strings.Trim(path.Clean("/"+p), "/")
}
