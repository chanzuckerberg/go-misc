package cmds_test

import (
	"testing"

	"github.com/chanzuckerberg/go-misc/cmds"
	"github.com/stretchr/testify/require"
)

func TestRoot(t *testing.T) {
	r := require.New(t)

	c := cmds.Root("foo", "short", "loooooong")

	r.NotNil(c)
	r.Equal("short", c.Short)
	r.Equal("loooooong", c.Long)
	r.NotNil(c.PersistentPreRunE)
}
