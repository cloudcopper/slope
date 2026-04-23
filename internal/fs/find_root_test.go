package fs_test

import (
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/egorse/slope/internal/fs"
	"github.com/egorse/slope/internal/utils/testutil"
)

func TestFindRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*testutil.FsBuilder)
		cwd     string
		search  string
		want    string
		wantErr error
	}{
		{
			name: "found in cwd",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/.slope")
			},
			cwd:    "/project",
			search: ".slope",
			want:   "/project",
		},
		{
			name: "found one level up",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/.slope")
				b.Dir("/project/sub")
			},
			cwd:    "/project/sub",
			search: ".slope",
			want:   "/project",
		},
		{
			name: "found several levels up",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/root/.slope")
				b.Dir("/root/a/b/c")
			},
			cwd:    "/root/a/b/c",
			search: ".slope",
			want:   "/root",
		},
		{
			name: "found as a file not directory",
			setup: func(b *testutil.FsBuilder) {
				b.File("/project/marker.txt", "")
				b.Dir("/project/sub")
			},
			cwd:    "/project/sub",
			search: "marker.txt",
			want:   "/project",
		},
		{
			name: "not found returns ErrRootNotFound",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/sub")
			},
			cwd:     "/project/sub",
			search:  ".slope",
			wantErr: fs.ErrRootNotFound,
		},
		{
			name: "empty filesystem returns ErrRootNotFound",
			setup: func(b *testutil.FsBuilder) {
				// nothing
			},
			cwd:     "/",
			search:  ".slope",
			wantErr: fs.ErrRootNotFound,
		},
		{
			name: "cwd with trailing slash is cleaned",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/.slope")
			},
			cwd:    "/project/",
			search: ".slope",
			want:   "/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fsys := memfs.New()
			builder := testutil.NewFsBuilder(fsys)
			tt.setup(builder)

			got, err := fs.FindRoot(fsys, tt.cwd, tt.search)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
