package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func init() {
	pathVarName = "__PATH__TEST"
	_ = os.Setenv(pathVarName, fmt.Sprint("testdata/apps/dir1:testdata/apps/dir2:"))
}

func TestRenderAppsOutput(t *testing.T) {
	apps := AppList{
		"/usr/local/bin": {"cat", "ls"},
		"/root/bin":      {"sudo", "wget"},
	}

	t.Run("Human readable output", func(t *testing.T) {
		assert.Equal(t, "/usr/local/bin:\n\t- cat\n\t- ls\n/root/bin:\n\t- sudo\n\t- wget\n", apps.String())
	})

	t.Run("Machine readable output (JSON)", func(t *testing.T) {
		assert.Equal(t, []byte(`{"/root/bin":["sudo","wget"],"/usr/local/bin":["cat","ls"]}`), apps.JSON())
	})
}

func TestGetDirEntries(t *testing.T) {
	assert.Equal(t, []string{"cat", "ls"}, getDirEntries("testdata/apps/dir1", ""))
	assert.Equal(t, []string{"cat", "ls"}, getDirEntries("testdata/apps/dir1", "cat,ls"))
	assert.Equal(t, []string{"a", "b", "c"}, getDirEntries("testdata/apps/dir2", ""))
	assert.Equal(t, []string{"a", "b"}, getDirEntries("testdata/apps/dir2", "a,b"))
	assert.Nil(t, getDirEntries("invalidPath___", ""))
	assert.Nil(t, getDirEntries("__invalidPath___", ""))
}

func TestGetAppList(t *testing.T) {
	t.Run("No filter", func(t *testing.T) {
		apps := GetAppList("")
		assert.Equal(t,
			[]byte(`{"testdata/apps/dir1":["cat","ls"],"testdata/apps/dir2":["a","b","c"],"undefined_location":null}`),
			apps.JSON())
	})

	t.Run("Filtered", func(t *testing.T) {
		apps := GetAppList("cat,b")
		assert.Equal(t,
			[]byte(`{"testdata/apps/dir1":["cat"],"testdata/apps/dir2":["b"],"undefined_location":null}`),
			apps.JSON())
	})
}
