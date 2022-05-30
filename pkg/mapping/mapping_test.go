package mapping

import (
	"errors"
	"strings"
	"testing"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/mapping/mappingfakes"
	"github.com/mdevilliers/org-scrounger/pkg/mapping/parser"
	"github.com/stretchr/testify/require"
)

func Test_MappedRepoIsReturned(t *testing.T) {

	reader := strings.NewReader(`
owner = "org-1"

# ignore doesn't map to a repo
_ > "please/ignore"

# static is a repo we want to care about
static > _

foo > "bar"
org-2/foo > "other-org"
needle > ["no", "yes", "maybe"]

`)
	rules, err := parser.UnMarshal("foo", reader)
	require.Nil(t, err)

	store := &mappingfakes.FakeRepoGetter{}
	store.GetRepoDetailsReturns(gh.Repository{}, gh.RateLimit{}, nil)

	mapper, err := New(rules, store)
	require.Nil(t, err)

	found, _, err := mapper.RepositoryFromContainer("bar")
	require.Nil(t, err)
	require.True(t, found)

	_, _, r := store.GetRepoDetailsArgsForCall(0)
	require.Equal(t, "foo", r)

	found, _, err = mapper.RepositoryFromContainer("yes")
	require.Nil(t, err)
	require.True(t, found)

	_, _, r = store.GetRepoDetailsArgsForCall(1)
	require.Equal(t, "needle", r)

	// lets pretend booyah! exists in github
	found, _, err = mapper.RepositoryFromContainer("booyah!")
	require.Nil(t, err)
	require.True(t, found)

	_, _, r = store.GetRepoDetailsArgsForCall(2)
	require.Equal(t, "booyah!", r)

	// lets pretend booyah! doesn;t exist in github
	store.GetRepoDetailsReturns(gh.Repository{}, gh.RateLimit{}, errors.New("error finding repo, try again."))

	found, _, err = mapper.RepositoryFromContainer("booyah!")
	require.NotNil(t, err)
	require.False(t, found)

	_, _, r = store.GetRepoDetailsArgsForCall(3)
	require.Equal(t, "booyah!", r)

}
