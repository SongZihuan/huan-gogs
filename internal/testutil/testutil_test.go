// Copyright 2022 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE and LICENSE.gogs file.

// Copyright 2025 Huan-Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInTest(t *testing.T) {
	assert.True(t, InTest)
}
