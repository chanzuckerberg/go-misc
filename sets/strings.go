package sets

type StringSet struct {
	strings map[string]struct{}
}

func NewStringSet() *StringSet {
	return &StringSet{}
}

func (ss *StringSet) Add(vals ...string) *StringSet {
	ss.ensure()

	for _, val := range vals {
		ss.strings[val] = struct{}{}
	}

	return ss
}

func (ss *StringSet) Remove(vals ...string) *StringSet {
	ss.ensure()

	for _, val := range vals {
		delete(ss.strings, val)
	}
	return ss
}

func (ss *StringSet) ContainsElement(val string) bool {
	ss.ensure()

	_, ok := ss.strings[val]
	return ok
}

func (ss *StringSet) Contains(other *StringSet) bool {
	ss.ensure()

	for _, val := range other.List() {
		ok := ss.ContainsElement(val)
		if !ok {
			return false
		}
	}
	return true
}

func (ss *StringSet) Subtract(other *StringSet) *StringSet {
	ss.ensure()

	difference := &StringSet{}

	for _, ours := range ss.List() {
		if !other.ContainsElement(ours) {
			difference.Add(ours)
		}
	}

	return difference
}

func (ss *StringSet) List() []string {
	ss.ensure()

	ret := []string{}
	for val := range ss.strings {
		ret = append(ret, val)
	}
	return ret
}

// Equals iff both sets contain the same items
func (ss *StringSet) Equals(other *StringSet) bool {
	return ss.Contains(other) && other.Contains(ss)
}

func (ss *StringSet) ensure() {
	if ss.strings == nil {
		ss.strings = map[string]struct{}{}
	}
}
