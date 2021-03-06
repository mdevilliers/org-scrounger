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

func Test_MappedRepoIsReturned(t *testing.T) { // nolint: funlen

	reader := strings.NewReader(`
owner = "org-1"

# ignore doesn't map to a repo
_ > "please/ignore"

# static is a repo we want to care about but isn't referenced explicitly
static > _

foo > "bar"
org-2/foo > "image:other-org"
needle > ["image:no", "image:yes", "something_else:maybe"]

`)
	rules, err := parser.UnMarshal("foo", reader)
	require.Nil(t, err)

	store := &mappingfakes.FakeRepoGetter{}
	store.GetRepoByNameReturns(gh.RepositorySlim{}, gh.RateLimit{}, nil)

	mapper, err := New(rules, store)
	require.Nil(t, err)

	found, _, err := mapper.RepositoryFromImage("bar")
	require.Nil(t, err)
	require.True(t, found)

	_, org, r := store.GetRepoByNameArgsForCall(0)
	require.Equal(t, "foo", r)
	require.Equal(t, "org-1", org)

	found, _, err = mapper.RepositoryFromImage("other-org")
	require.Nil(t, err)
	require.True(t, found)

	_, org, r = store.GetRepoByNameArgsForCall(1)
	require.Equal(t, "foo", r)
	require.Equal(t, "org-2", org)

	found, _, err = mapper.RepositoryFromImage("yes")
	require.Nil(t, err)
	require.True(t, found)

	_, _, r = store.GetRepoByNameArgsForCall(2)
	require.Equal(t, "needle", r)

	// lets pretend booyah! exists in github
	found, _, err = mapper.RepositoryFromImage("booyah!")
	require.Nil(t, err)
	require.True(t, found)

	_, _, r = store.GetRepoByNameArgsForCall(3)
	require.Equal(t, "booyah!", r)

	// lets pretend booyah! doesn;t exist in github
	store.GetRepoByNameReturns(gh.RepositorySlim{}, gh.RateLimit{}, errors.New("error finding repo, try again"))

	found, _, err = mapper.RepositoryFromImage("booyah!")
	require.NotNil(t, err)
	require.False(t, found)

	_, _, r = store.GetRepoByNameArgsForCall(4)
	require.Equal(t, "booyah!", r)

}
