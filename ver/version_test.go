package ver_test

import (
	"testing"

	"github.com/blang/semver"
	"github.com/chanzuckerberg/go-misc/ver"
	"github.com/stretchr/testify/assert"
)

func TestVersionString(t *testing.T) {
	type args struct {
		version    string
		gitsha     string
		releaseStr string
		dirtyStr   string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"simple", args{"0.1.0", "abcdef", "true", "false"}, "0.1.0", false},
		{"bad release", args{"0.1.0", "abcdef", "junk", "false"}, "", true},
		{"bad dirty", args{"0.1.0", "abcdef", "false", "junk"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ver.VersionString(tt.args.version, tt.args.gitsha, tt.args.releaseStr, tt.args.dirtyStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("VersionString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VersionString() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestVersionCacheKey(t *testing.T) {
// 	type args struct {
// 		version string
// 		gitsha  string
// 		release string
// 		dirty   string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := VersionCacheKey(tt.args.version, tt.args.gitsha, tt.args.release, tt.args.dirty); got != tt.want {
// 				t.Errorf("VersionCacheKey() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestParse(t *testing.T) {
	a := assert.New(t)

	testCases := []struct {
		input   string
		version string
		sha     string
		dirty   bool
	}{
		{"0.1.0", "0.1.0", "", false},
		{"0.1.0-abcdef", "0.1.0", "abcdef", false},
		{"0.1.0-abcdef.dirty", "0.1.0", "abcdef", true},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			v, sha, dirty := ver.ParseVersion(tc.input)
			semVersion, e := semver.Parse(tc.version)
			a.NoError(e)
			a.Equal(semVersion, v)
			a.Equal(tc.sha, sha)
			a.Equal(tc.dirty, dirty)
		})
	}

}
