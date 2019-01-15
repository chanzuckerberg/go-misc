package ver_test

import (
	"testing"

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
