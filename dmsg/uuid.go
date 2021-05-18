package main

import (
	"github.com/google/uuid"
	"os"
)

var uuidFileName = ".uuid"

func getUUID() uuid.UUID {
	// Trying to read existing uuid.
	b, err := os.ReadFile(uuidFileName)
	if err == nil {
		// Read existing UUID.
		UUID, err := uuid.FromBytes(b)
		if err != nil {
			panic(err)
		}

		return UUID
	}

	// Create new file to save UUID.
	f, err := os.Create(uuidFileName)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	// Generate new random UUID.
	UUID := uuid.New()
	b, err = UUID.MarshalBinary()
	if err != nil {
		panic(err)
	}

	// Write newly generated UUID to the file.
	if _, err = f.Write(b); err != nil {
		panic(err)
	}

	return UUID
}
