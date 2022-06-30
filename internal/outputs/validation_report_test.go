// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package outputs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapErrorSliceToStringSlice(t *testing.T) {
	errors := []error{errors.New("error-one"), errors.New("error-two"), errors.New("error-three")}
	expected := []string{"error-one", "error-two", "error-three"}
	result := mapErrorSliceToStringSlice(errors)
	assert.ElementsMatch(t, expected, result, "mapped slice of strings matches")
}
