package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Parse(t *testing.T) {

	testFile := `
# global
# default owner if not declared in a mapping entry	
owner = "foo"

# static repos 
repo_foo > _ 
	
# mappings
# one to one mapping
repo-1 > "svc-one-service"
# one to one mapping with another owner (public)
owner-2/repo-2 > "svc-two-service" 
# repo contains the code for many services
repo-3 > ["svc-three-one-service","svc-three-two-service","svc-three-three-service"]

# ignore these services
_ > "third-party/something/something"
`
	r := strings.NewReader(testFile)
	o, err := UnMarshal("test", r)
	require.Nil(t, err)
	// repr.Println(o, repr.Indent("  "), repr.OmitEmpty(true))
	require.Len(t, o.Entries, 14) // includes comments
}
