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
	"math"

	"github.com/stretchr/testify/require"
	"gopkg.in/mgutz/dat.v2/sqlx-runner"
)

func (s *DbZuite) TestExample() {
	ws := s.store.defs.MustNewWorksheet("simple")
	ws.MustSet("name", NewText("Alice"))

	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)
		return session.Save(ws)
	})

	var wsFromStore *Worksheet
	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)
		var err error
		wsFromStore, err = session.Load(ws.Id())
		return err
	})

	require.Equal(s.T(), `"Alice"`, wsFromStore.MustGet("name").String())
}

func (s *DbZuite) TestSave() {
	ws, err := s.store.defs.NewWorksheet("simple")
	require.NoError(s.T(), err)

	err = ws.Set("name", NewText("Alice"))
	require.NoError(s.T(), err)

	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)
		return session.Save(ws)
	})

	wsRecs, valuesRecs, _ := s.DbState()

	require.Equal(s.T(), []rWorksheet{
		{
			Id:      ws.Id(),
			Version: 1,
			Name:    "simple",
		},
	}, wsRecs)

	require.Equal(s.T(), []rValueForTesting{
		{
			WorksheetId: ws.Id(),
			Index:       IndexId,
			FromVersion: 1,
			ToVersion:   math.MaxInt32,
			Value:       ws.Id(),
		},
		{
			WorksheetId: ws.Id(),
			Index:       IndexVersion,
			FromVersion: 1,
			ToVersion:   math.MaxInt32,
			Value:       `1`,
		},
		{
			WorksheetId: ws.Id(),
			Index:       83,
			FromVersion: 1,
			ToVersion:   math.MaxInt32,
			Value:       `Alice`,
		},
	}, valuesRecs)

	// Upon Save, orig needs to be set to data.
	require.Empty(s.T(), ws.diff())
}

func (s *DbZuite) TestUpdate() {
	ws, err := s.store.defs.NewWorksheet("simple")
	require.NoError(s.T(), err)

	err = ws.Set("name", NewText("Alice"))
	require.NoError(s.T(), err)

	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)
		return session.Save(ws)
	})

	err = ws.Set("name", NewText("Bob"))
	require.NoError(s.T(), err)

	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)
		return session.Update(ws)
	})

	wsRecs, valuesRecs, _ := s.DbState()

	require.Equal(s.T(), []rWorksheet{
		{
			Id:      ws.Id(),
			Version: 2,
			Name:    "simple",
		},
	}, wsRecs)

	require.Equal(s.T(), []rValueForTesting{
		{
			WorksheetId: ws.Id(),
			Index:       IndexId,
			FromVersion: 1,
			ToVersion:   math.MaxInt32,
			Value:       ws.Id(),
		},
		{
			WorksheetId: ws.Id(),
			Index:       IndexVersion,
			FromVersion: 1,
			ToVersion:   1,
			Value:       `1`,
		},
		{
			WorksheetId: ws.Id(),
			Index:       IndexVersion,
			FromVersion: 2,
			ToVersion:   math.MaxInt32,
			Value:       `2`,
		},
		{
			WorksheetId: ws.Id(),
			Index:       83,
			FromVersion: 1,
			ToVersion:   1,
			Value:       `Alice`,
		},
		{
			WorksheetId: ws.Id(),
			Index:       83,
			FromVersion: 2,
			ToVersion:   math.MaxInt32,
			Value:       `Bob`,
		},
	}, valuesRecs)

	// Upon update, version must increase
	require.Equal(s.T(), 2, ws.Version())

	// Upon Update, orig needs to be set to data.
	require.Empty(s.T(), ws.diff())
}

func (s *DbZuite) TestUpdateUndefinedField() {
	ws, err := s.store.defs.NewWorksheet("simple")
	require.NoError(s.T(), err)

	err = ws.Set("name", NewText("Alice"))
	require.NoError(s.T(), err)

	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)
		return session.Save(ws)
	})

	err = ws.Set("age", MustNewValue("73"))
	require.NoError(s.T(), err)

	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)
		return session.Update(ws)
	})
}

func (s *DbZuite) TestProperlyLoadUndefinedField() {
	var wsId string
	s.MustRunTransaction(func(tx *runner.Tx) error {
		ws := defs.MustNewWorksheet("simple")
		wsId = ws.Id()
		ws.MustSet("age", MustNewValue("123456"))

		session := s.store.Open(tx)
		return session.Save(ws)
	})

	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)

		ws, err := session.Load(wsId)
		if err != nil {
			return err
		}
		ws.MustUnset("age")

		return session.Update(ws)
	})

	// Fresh load should show age as being unset.
	var fresh *Worksheet
	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)
		var err error
		fresh, err = session.Load(wsId)
		return err
	})

	require.False(s.T(), fresh.MustIsSet("age"))

	// Lastly, check db state.
	wsRecs, valuesRecs, _ := s.DbState()

	require.Equal(s.T(), []rWorksheet{
		{
			Id:      wsId,
			Version: 2,
			Name:    "simple",
		},
	}, wsRecs)

	require.Equal(s.T(), []rValueForTesting{
		{
			WorksheetId: wsId,
			Index:       IndexId,
			FromVersion: 1,
			ToVersion:   math.MaxInt32,
			Value:       wsId,
		},
		{
			WorksheetId: wsId,
			Index:       IndexVersion,
			FromVersion: 1,
			ToVersion:   1,
			Value:       `1`,
		},
		{
			WorksheetId: wsId,
			Index:       IndexVersion,
			FromVersion: 2,
			ToVersion:   math.MaxInt32,
			Value:       `2`,
		},
		{
			WorksheetId: wsId,
			Index:       91,
			FromVersion: 1,
			ToVersion:   1,
			Value:       `123456`,
		},
		{
			WorksheetId: wsId,
			Index:       91,
			FromVersion: 2,
			ToVersion:   math.MaxInt32,
			IsUndefined: true,
		},
	}, valuesRecs)
}

func (s *DbZuite) TestUpdateOnUpdateDoesNothing() {
	ws := s.store.defs.MustNewWorksheet("simple")
	ws.MustSet("name", alice)

	require.Equal(s.T(), 1, ws.Version())

	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)
		return session.Save(ws)
	})

	require.Equal(s.T(), 1, ws.Version())

	ws.MustSet("name", bob)
	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)
		return session.Update(ws)
	})

	require.Equal(s.T(), 2, ws.Version())

	s.MustRunTransaction(func(tx *runner.Tx) error {
		session := s.store.Open(tx)
		return session.Update(ws)
	})

	require.Equal(s.T(), 2, ws.Version())
}

func (s *DbZuite) MustRunTransaction(fn func(tx *runner.Tx) error) {
	err := RunTransaction(s.db, fn)
	require.NoError(s.T(), err)
}
