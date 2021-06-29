package cmd

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
)

func init() {
	getGoVersionFunc = mockGetGoVersion
}

func TestNewGolangVersion(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		gv, err := GetGolangVersion()
		assert.Nil(t, err)
		assert.NotNil(t, gv)
		assert.Equal(t, &golangVersion{
			Version: "1.16.4",
			OS:      "windows",
			Arch:    "amd64",
		}, gv)
	})

	t.Run("Error", func(t *testing.T) {
		getGoVersionFunc = mockGetGoVersionErrored
		gv, err := GetGolangVersion()
		assert.Nil(t, gv)
		assert.NotNil(t, err)
		assert.Equal(t, errors.New("error"), err)
	})
}

func TestGolangVersion_String(t *testing.T) {
	gv := mockGolangVersion()
	assert.Equal(t, "version=1.16.4, os=windows, arch=amd64", gv.String())
}

func TestGolangVersion_JSON(t *testing.T) {
	gv := mockGolangVersion()
	assert.Equal(t, []byte(`{"version":"1.16.4","os":"windows","arch":"amd64"}`), gv.JSON())
}

func TestGolangVersion_FromRaw(t *testing.T) {
	gv := mockGolangVersion()
	assert.Equal(t, "1.16.4", gv.Version)
	assert.Equal(t, "windows", gv.OS)
	assert.Equal(t, "amd64", gv.Arch)
}

func TestGetGoVersion(t *testing.T) {
	actual, err := exec.Command("go", "version").Output()
	if err != nil {
		t.Fatalf("Failed to execute `go version`: %v", err)
	}

	out, err := getGoVersion()
	assert.Nil(t, err)
	assert.Equal(t, string(actual), out)
}

func mockGetGoVersion() (string, error) {
	return "go version go1.16.4 windows/amd64", nil
}

func mockGolangVersion() *golangVersion {
	gv := new(golangVersion)
	out, _ := mockGetGoVersion()
	gv.fromRaw(out)
	return gv
}

func mockGetGoVersionErrored() (string, error) {
	return "", errors.New("error")
}
