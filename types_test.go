// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package worksheets

import (
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Zuite) TestTypeAssignableTo() {
	cases := []struct {
		left, right Type
	}{
		{&tUndefinedType{}, &tTextType{}},
		{&tUndefinedType{}, &tBoolType{}},
		{&tUndefinedType{}, &tNumberType{0}},
		{&tUndefinedType{}, &tNumberType{1}},

		{&tTextType{}, &tTextType{}},

		{&tBoolType{}, &tBoolType{}},

		{&tNumberType{0}, &tNumberType{0}},
		{&tNumberType{1}, &tNumberType{1}},
	}
	for _, ex := range cases {
		require.True(s.T(), ex.left.AssignableTo(ex.right), "%s should be assignable to %s", ex.left, ex.right)
	}
}

func (s *Zuite) TestTypeNotAssignableTo() {
	cases := []struct {
		left, right Type
	}{
		{&tTextType{}, &tUndefinedType{}},
		{&tBoolType{}, &tUndefinedType{}},
		{&tNumberType{0}, &tUndefinedType{}},
		{&tNumberType{1}, &tUndefinedType{}},

		{&tBoolType{}, &tTextType{}},
		{&tNumberType{9}, &tTextType{}},

		{&tTextType{}, &tBoolType{}},
		{&tNumberType{9}, &tBoolType{}},

		{&tTextType{}, &tNumberType{1}},
		{&tNumberType{2}, &tNumberType{1}},
	}
	for _, ex := range cases {
		assert.False(s.T(), ex.left.AssignableTo(ex.right), "%s should not be assignable to %s", ex.left, ex.right)
	}
}

func (s *Zuite) TestTypeString() {
	cases := map[Type]string{
		&tUndefinedType{}:           "undefined",
		&tTextType{}:                "text",
		&tBoolType{}:                "bool",
		&tNumberType{1}:             "number[1]",
		&SliceType{&tBoolType{}}:    "[]bool",
		&Definition{name: "simple"}: "simple",
	}
	for typ, expected := range cases {
		assert.Equal(s.T(), expected, typ.String(), expected)
	}
}

func (s *Zuite) TestWorksheetDefinition_Fields() {
	defs, err := NewDefinitions(strings.NewReader(`worksheet simple {1:name text}`))
	require.NoError(s.T(), err)

	ws := defs.MustNewWorksheet("simple")

	fields := ws.Type().(*Definition).Fields()
	require.Len(s.T(), fields, 3)

	expectedFields := []*Field{
		{
			index: 1,
			name:  "name",
			typ:   &tTextType{},
		},
		{
			index: -2,
			name:  "id",
			typ:   &tTextType{},
		},
		{
			index: -1,
			name:  "version",
			typ:   &tNumberType{},
		},
	}
	for _, field := range expectedFields {
		require.Contains(s.T(), fields, field)
	}
}
