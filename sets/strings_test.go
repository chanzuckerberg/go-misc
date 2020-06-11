package sets

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringSetAdd(t *testing.T) {
	r := require.New(t)
	stringSet := &StringSet{}

	stringSet.Add("foo")
	r.True(stringSet.ContainsElement("foo"))

	stringSet.Add("foo", "bar", "baz")
	r.True(stringSet.ContainsElement("foo"))
	r.True(stringSet.ContainsElement("bar"))
	r.True(stringSet.ContainsElement("baz"))
}

func TestStringSetRemove(t *testing.T) {
	r := require.New(t)
	stringSet := &StringSet{}
	// remove not there
	stringSet.Add("missing")
	r.True(stringSet.ContainsElement("missing"))
	stringSet.Remove("missing")
	r.False(stringSet.ContainsElement("missing"))
}

func TestStringSetList(t *testing.T) {
	r := require.New(t)
	stringSet := &StringSet{}
	stringSet.Add("foo", "bar", "baz")

	ls := stringSet.List()

	sort.Strings(ls)
	r.Equal(ls, []string{"bar", "baz", "foo"})
}

func TestSTringSetEquals(t *testing.T) {
	r := require.New(t)
	this := &StringSet{}
	that := &StringSet{}

	this.Add("a", "b", "c")
	that.Add("a", "b", "c")

	r.True(this.Equals(that))

	this.Add("d")
	r.False(this.Equals(that))

	that.Add("d")
	r.True(this.Equals(that))

	that.Add("e")
	r.False(this.Equals(that))

	this.Add("e")
	r.True(this.Equals(that))
}
