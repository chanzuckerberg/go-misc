package sets

type StringSet struct {
	strings map[string]struct{}
}

func (ss *StringSet) Add(vals ...string) {
	ss.ensure()

	for _, val := range vals {
		ss.strings[val] = struct{}{}
	}
}

func (ss *StringSet) Remove(vals ...string) {
	ss.ensure()

	for _, val := range vals {
		delete(ss.strings, val)
	}
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
