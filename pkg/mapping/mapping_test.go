package mapping

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/mapping/mappingfakes"
	"github.com/mdevilliers/org-scrounger/pkg/mapping/parser"
	"github.com/stretchr/testify/require"
)

func Test_MappedImageIsReturned(t *testing.T) { //nolint:funlen

	reader := strings.NewReader(`
owner = "org-1"

# ignore doesn't map to a repo
_ > "please/ignore"

# static is a repo we want to care about but isn't referenced explicitly
static > _

foo > "bar"
org-2/foo > "image:other-org"
needle > ["image:no", "image:yes", "something_else:maybe", "sonarcloud:foo"]
`)
	rules, err := parser.UnMarshal("foo", reader)
	require.Nil(t, err)

	ctx := context.Background()
	image := &Image{Name: "bar"}

	store := &mappingfakes.FakeRepoGetter{}
	store.GetRepoByNameReturns(gh.RepositorySlim{}, gh.RateLimit{}, nil)

	mapper := New(rules)

	found, err := mapper.Decorate(ctx, store, nil, image)
	require.Nil(t, err)
	require.True(t, found)

	_, org, r := store.GetRepoByNameArgsForCall(0)
	require.Equal(t, "foo", r)
	require.Equal(t, "org-1", org)

	image = &Image{Name: "other-org"}
	found, err = mapper.Decorate(ctx, store, nil, image)
	require.Nil(t, err)
	require.True(t, found)

	_, org, r = store.GetRepoByNameArgsForCall(1)
	require.Equal(t, "foo", r)
	require.Equal(t, "org-2", org)

	image = &Image{Name: "yes"}
	found, err = mapper.Decorate(ctx, store, nil, image)
	require.Nil(t, err)
	require.True(t, found)

	_, _, r = store.GetRepoByNameArgsForCall(2)
	require.Equal(t, "needle", r)

	// lets pretend booyah! exists in github
	image = &Image{Name: "booyah!"}
	found, err = mapper.Decorate(ctx, store, nil, image)
	require.Nil(t, err)
	require.True(t, found)

	_, _, r = store.GetRepoByNameArgsForCall(3)
	require.Equal(t, "booyah!", r)

	// lets pretend booyah! doesn;t exist in github
	store.GetRepoByNameReturns(gh.RepositorySlim{}, gh.RateLimit{}, errors.New("error finding repo, try again"))
	found, err = mapper.Decorate(ctx, store, nil, image)
	require.NotNil(t, err)
	require.False(t, found)

	_, _, r = store.GetRepoByNameArgsForCall(4)
	require.Equal(t, "booyah!", r)
}

func Test_NamespacedItemIsReturned(t *testing.T) {

	reader := strings.NewReader(`
owner = "org-1"

foo > "abc:bar"
needle > ["abc:no", "def:yes", "maybe"]
`)
	rules, err := parser.UnMarshal("foo", reader)
	require.Nil(t, err)

	mapper := New(rules)

	s, v, _ := mapper.resolve("abc", "bar")
	require.Equal(t, ok, s)
	require.Equal(t, "foo", v)

	s, v, _ = mapper.resolve("does-not-exist", "maybe")
	require.Equal(t, ok, s)
	require.Equal(t, "needle", v)

	s, v, _ = mapper.resolve("def", "yes")
	require.Equal(t, ok, s)
	require.Equal(t, "needle", v)
}
