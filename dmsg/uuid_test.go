package main

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func init() {
	uuidFileName = ".uuid-test"
}

func TestGetUUID(t *testing.T) {
	t.Cleanup(func() {
		_ = os.Remove(uuidFileName)
	})

	UUID := getUUID()
	assert.IsType(t, uuid.UUID{}, UUID)
	assert.FileExists(t, uuidFileName)

	UUID2 := getUUID()
	assert.Equal(t, UUID, UUID2)
}
