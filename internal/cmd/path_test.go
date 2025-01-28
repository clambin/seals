package cmd

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_makeRelativePath(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		source  string
		getWd   func() (string, error)
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "inside",
			base:    "/mnt/staging",
			source:  "secret.yaml",
			getWd:   func() (string, error) { return "/mnt/staging/secrets", nil },
			want:    "secrets/secret.yaml",
			wantErr: assert.NoError,
		},
		{
			name:    "below",
			base:    "/mnt/staging",
			source:  "secrets/secret.yaml",
			getWd:   func() (string, error) { return "/mnt/staging", nil },
			want:    "secrets/secret.yaml",
			wantErr: assert.NoError,
		},
		{
			name:    "outside",
			base:    "/mnt/staging",
			source:  "secrets/secret.yaml",
			getWd:   func() (string, error) { return "/mnt", nil },
			want:    "../secrets/secret.yaml",
			wantErr: assert.NoError,
		},
		{
			name:    "failure",
			base:    "/mnt/staging",
			source:  "secrets/secret.yaml",
			getWd:   func() (string, error) { return "", os.ErrNotExist },
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getWd = tt.getWd
			path, err := makeRelativePath(tt.base, tt.source)
			assert.Equal(t, tt.want, path)
			tt.wantErr(t, err)
		})
	}
}

func Test_shouldUpdate(t *testing.T) {
	tmpdir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpdir, "file1"), []byte("content"), 0644))
	require.NoError(t, os.Chtimes(filepath.Join(tmpdir, "file1"), time.Now(), time.Now().Add(-time.Hour)))
	require.NoError(t, os.WriteFile(filepath.Join(tmpdir, "file2"), []byte("content"), 0644))

	type args struct {
		source      string
		destination string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "should not update",
			args:    args{source: "file1", destination: "file2"},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name:    "should update",
			args:    args{source: "file2", destination: "file1"},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name:    "should create",
			args:    args{source: "file1", destination: "new-file"},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name:    "missing source file",
			args:    args{source: "missing", destination: "file2"},
			want:    false,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shouldUpdate(filepath.Join(tmpdir, tt.args.source), filepath.Join(tmpdir, tt.args.destination))
			assert.Equal(t, tt.want, got)
			tt.wantErr(t, err)
		})
	}
}

func Test_isWritableDirectory(t *testing.T) {
	tmpdir := t.TempDir()
	require.NoError(t, initFS(tmpdir))

	require.NoError(t, os.WriteFile(filepath.Join(tmpdir, "not-a-directory"), []byte(""), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(tmpdir, "read-only"), 0555))

	assert.NoError(t, isWritableDirectory(filepath.Join(tmpdir, "manifests")))
	assert.Error(t, isWritableDirectory(filepath.Join(tmpdir, "missing")))
	assert.Error(t, isWritableDirectory(filepath.Join(tmpdir, "not-a-directory")))
	assert.Error(t, isWritableDirectory(filepath.Join(tmpdir, "read-only")))
}
