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
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgutz/dat.v2/sqlx-runner"
)

type DbZuite struct {
	suite.Suite
	db    *runner.DB
	store *DbStore
}

func (s *DbZuite) SetupSuite() {
	// db
	dbUrl := "postgres://ws_user:@localhost/ws_test?sslmode=disable"
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		panic(err)
	}
	s.db = runner.NewDB(db, "postgres")

	// store
	s.store = NewStore(defs)
}

func (s *DbZuite) SetupTest() {
	for table := range tableToEntities {
		_, err := s.db.Exec(fmt.Sprintf("truncate %s", table))
		if err != nil {
			panic(err)
		}
	}
}

func (s *DbZuite) TearDownSuite() {
	err := s.db.DB.Close()
	if err != nil {
		panic(err)
	}
}

func TestRunAllTheDbTests(t *testing.T) {
	suite.Run(t, new(DbZuite))
}

type rValueForTesting struct {
	WorksheetId string
	Index       int
	FromVersion int
	ToVersion   int
	Value       string
	IsUndefined bool
}

type rSliceElementForTesting struct {
	SliceId     string
	Rank        int
	FromVersion int
	ToVersion   int
	Value       string
	IsUndefined bool
}

func (s *DbZuite) DbState() ([]rWorksheet, []rValueForTesting, []rSliceElementForTesting) {
	var (
		err                 error
		wsRecs              []rWorksheet
		dbValuesRecs        []rValue
		dbSliceElementsRecs []rSliceElement
	)

	err = s.db.
		Select("*").
		From("worksheets").
		OrderBy("id").
		QueryStructs(&wsRecs)
	require.NoError(s.T(), err)

	s.db.
		Select("*").
		From("worksheet_values").
		OrderBy("worksheet_id, index, from_version").
		QueryStructs(&dbValuesRecs)
	require.NoError(s.T(), err)

	err = s.db.
		Select("*").
		From("worksheet_slice_elements").
		OrderBy("slice_id, rank, from_version").
		QueryStructs(&dbSliceElementsRecs)
	require.NoError(s.T(), err)

	// rValue to rValueForTesting
	valuesRecs := make([]rValueForTesting, len(dbValuesRecs))
	for i, dbValueRec := range dbValuesRecs {
		valuesRecs[i] = rValueForTesting{
			WorksheetId: dbValueRec.WorksheetId,
			Index:       dbValueRec.Index,
			FromVersion: dbValueRec.FromVersion,
			ToVersion:   dbValueRec.ToVersion,
		}
		if dbValueRec.Value.Valid {
			valuesRecs[i].Value = dbValueRec.Value.String
		} else {
			valuesRecs[i].IsUndefined = true
		}
	}

	// rSliceElement to rSliceElementForTesting
	sliceElementsRecs := make([]rSliceElementForTesting, len(dbSliceElementsRecs))
	for i, dbSliceElementRec := range dbSliceElementsRecs {
		sliceElementsRecs[i] = rSliceElementForTesting{
			SliceId:     dbSliceElementRec.SliceId,
			Rank:        dbSliceElementRec.Rank,
			FromVersion: dbSliceElementRec.FromVersion,
			ToVersion:   dbSliceElementRec.ToVersion,
		}
		if dbSliceElementRec.Value.Valid {
			sliceElementsRecs[i].Value = dbSliceElementRec.Value.String
		} else {
			sliceElementsRecs[i].IsUndefined = true
		}
	}

	return wsRecs, valuesRecs, sliceElementsRecs
}

func p(v string) *string {
	return &v
}
