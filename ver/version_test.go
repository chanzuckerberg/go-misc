package ver_test

import (
	"reflect"
	"testing"

	"github.com/blang/semver"
	"github.com/chanzuckerberg/go-misc/ver"
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
		{"prerelease", args{"0.1.0", "abcdef", "false", "false"}, "0.1.0-pre+abcdef", false},
		{"prerelease dirty", args{"0.1.0", "abcdef", "false", "true"}, "0.1.0-pre+abcdef.dirty", false},
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

func TestParseVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    semver.Version
		want1   string
		want2   bool
		wantErr bool
	}{
		{"_", args{"0.1.0"}, semver.MustParse("0.1.0"), "", false, false},
		{"_", args{"0.1.0-abcdef"}, semver.MustParse("0.1.0"), "abcdef", false, false},
		{"_", args{"0.1.0-abcdef.dirty"}, semver.MustParse("0.1.0"), "abcdef", true, false},
		{"_", args{"a.1.0"}, semver.MustParse("0.0.0"), "", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := ver.ParseVersion(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseVersion() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseVersion() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("ParseVersion() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
