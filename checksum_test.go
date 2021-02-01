package deebee_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jacekolszak/deebee"
	"github.com/jacekolszak/deebee/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecksumIntegrityChecker(t *testing.T) {
	t.Run("should return default ChecksumIntegrityChecker", func(t *testing.T) {
		checker := deebee.ChecksumIntegrityChecker()
		assert.NotNil(t, checker)
	})

	t.Run("should use custom checksum algorithm", func(t *testing.T) {
		expectedSum := []byte{1, 2, 3, 4}
		algorithm := &fixedAlgorithm{sum: expectedSum}
		dir := fake.ExistingDir()
		db, err := deebee.Open(dir, deebee.ChecksumIntegrityChecker(deebee.Algorithm(algorithm)))
		require.NoError(t, err)
		// when
		writeData(t, db, "state", []byte("data"))
		// then
		files := filterFilesWithExtension(dir.FakeDir("state").Files(), "fixed")
		require.NotEmpty(t, files)
		assert.Equal(t, expectedSum, files[0].Data())
	})
}

func filterFilesWithExtension(files []*fake.File, extension string) []*fake.File {
	var filtered []*fake.File
	for _, file := range files {
		if strings.HasSuffix(file.Name(), "."+extension) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

type fixedAlgorithm struct {
	sum []byte
}

func (c fixedAlgorithm) FileExtension() string {
	return "fixed"
}

func (c fixedAlgorithm) NewSum() deebee.Sum {
	return &fixedSum{sum: c.sum}
}

type fixedSum struct {
	sum []byte
}

func (c *fixedSum) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (c *fixedSum) Marshal() []byte {
	return c.sum
}

func TestHashSum_Marshal(t *testing.T) {
	t.Run("should marshal sum", func(t *testing.T) {
		tests := map[string]struct {
			algorithm   deebee.ChecksumAlgorithm
			expectedSum string
		}{
			"fnv128a": {
				algorithm:   deebee.Fnv128a,
				expectedSum: "695b598c64757277b806e9704d5d6a5d",
			},
			"fixed": {
				algorithm:   &fixedAlgorithm{sum: []byte{1, 2, 3, 4}},
				expectedSum: "01020304",
			},
		}
		for name, test := range tests {

			t.Run(name, func(t *testing.T) {
				sum := test.algorithm.NewSum()
				_, err := sum.Write([]byte("data"))
				require.NoError(t, err)
				// when
				bytes := sum.Marshal()
				// then
				assert.Equal(t, test.expectedSum, fmt.Sprintf("%x", bytes))
			})
		}
	})

	t.Run("should marshal sum after two writes", func(t *testing.T) {
		tests := map[string]struct {
			algorithm   deebee.ChecksumAlgorithm
			expectedSum string
		}{
			"fnv128a": {
				algorithm:   deebee.Fnv128a,
				expectedSum: "695b598c64757277b806e9704d5d6a5d",
			},
		}
		for name, test := range tests {

			t.Run(name, func(t *testing.T) {
				sum := test.algorithm.NewSum()
				_, err := sum.Write([]byte("da"))
				require.NoError(t, err)
				_, err = sum.Write([]byte("ta"))
				require.NoError(t, err)
				// when
				bytes := sum.Marshal()
				// then
				assert.Equal(t, test.expectedSum, fmt.Sprintf("%x", bytes))
			})
		}
	})
}