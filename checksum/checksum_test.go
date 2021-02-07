package checksum_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/jacekolszak/deebee/checksum"
	"github.com/jacekolszak/deebee/fake"
	"github.com/jacekolszak/deebee/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecksumIntegrityChecker(t *testing.T) {
	t.Run("should return default ChecksumIntegrityChecker", func(t *testing.T) {
		checker := checksum.ChecksumIntegrityChecker()
		assert.NotNil(t, checker)
	})

	t.Run("should return error when option returned error", func(t *testing.T) {
		dir := fake.ExistingDir()
		optionReturningError := func(checker *checksum.ChecksumDataIntegrityChecker) error {
			return errors.New("failed")
		}
		db, err := store.Open(dir, checksum.ChecksumIntegrityChecker(optionReturningError))
		assert.Error(t, err)
		assert.Nil(t, db)
	})

	t.Run("should return error error when checksum algorithm has invalid name", func(t *testing.T) {
		names := []string{"", ".", "-", " ", "A", "Z", ".7z", "a ", " 6"}
		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				algorithm := invalidNameAlgorithm{name: name}
				db, err := store.Open(fake.ExistingDir(), checksum.ChecksumIntegrityChecker(checksum.Algorithm(algorithm)))
				assert.Error(t, err)
				assert.Nil(t, db)
			})
		}
	})

	t.Run("should accept algorithm with valid name", func(t *testing.T) {
		names := []string{"a", "z", "0", "9", "2b", "fnv128a"}
		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				algorithm := invalidNameAlgorithm{name: name}
				db, err := store.Open(fake.ExistingDir(), checksum.ChecksumIntegrityChecker(checksum.Algorithm(algorithm)))
				require.NoError(t, err)
				assert.NotNil(t, db)
			})
		}
	})

	t.Run("should write checksum to a file with an extension having algorithm name", func(t *testing.T) {
		expectedSum := []byte{1, 2, 3, 4}
		algorithm := &fixedAlgorithm{sum: expectedSum}
		dir := fake.ExistingDir()
		db, err := store.Open(dir, checksum.ChecksumIntegrityChecker(checksum.Algorithm(algorithm)))
		require.NoError(t, err)
		// when
		writeData(t, db, []byte("data"))
		// then
		files := filterFilesWithExtension(dir.Files(), "fixed")
		require.NotEmpty(t, files)
		assert.Equal(t, expectedSum, files[0].Data())
	})

	t.Run("should use checksum algorithm", func(t *testing.T) {
		expectedSum := []byte{1, 2, 3, 4}
		algorithm := &fixedAlgorithm{sum: expectedSum}
		dir := fake.ExistingDir()
		db, err := store.Open(dir, checksum.ChecksumIntegrityChecker(checksum.Algorithm(algorithm)))
		require.NoError(t, err)
		expectedData := []byte("data")
		// when
		writeData(t, db, expectedData)
		actualData := readData(t, db)
		// then
		assert.Equal(t, expectedData, actualData)
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

type invalidNameAlgorithm struct {
	name string
}

func (i invalidNameAlgorithm) NewSum() checksum.Sum {
	return nil
}

func (i invalidNameAlgorithm) Name() string {
	return i.name
}

type fixedAlgorithm struct {
	sum []byte
}

func (c fixedAlgorithm) Name() string {
	return "fixed"
}

func (c fixedAlgorithm) NewSum() checksum.Sum {
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
	tests := map[string]struct {
		algorithm   checksum.ChecksumAlgorithm
		expectedSum string
	}{
		"crc32": {
			algorithm:   checksum.CRC32,
			expectedSum: "adf3f363",
		},
		"crc64": {
			algorithm:   checksum.CRC64,
			expectedSum: "3408641350000000",
		},
		"sha512": {
			algorithm:   checksum.SHA512,
			expectedSum: "77c7ce9a5d86bb386d443bb96390faa120633158699c8844c30b13ab0bf92760b7e4416aea397db91b4ac0e5dd56b8ef7e4b066162ab1fdc088319ce6defc876",
		},
		"md5": {
			algorithm:   checksum.MD5,
			expectedSum: "8d777f385d3dfec8815d20f7496026dc",
		},
		"fnv32": {
			algorithm:   checksum.FNV32,
			expectedSum: "74cb23bd",
		},
		"fnv32a": {
			algorithm:   checksum.FNV32a,
			expectedSum: "d872e2a5",
		},
		"fnv64": {
			algorithm:   checksum.FNV64,
			expectedSum: "14dfb87eecce7a1d",
		},
		"fnv64a": {
			algorithm:   checksum.FNV64a,
			expectedSum: "855b556730a34a05",
		},
		"fnv128": {
			algorithm:   checksum.FNV128,
			expectedSum: "66ab729108757277b806e89c746322b5",
		},
		"fnv128a": {
			algorithm:   checksum.FNV128a,
			expectedSum: "695b598c64757277b806e9704d5d6a5d",
		},
		"fixed": {
			algorithm:   &fixedAlgorithm{sum: []byte{1, 2, 3, 4}},
			expectedSum: "01020304",
		},
	}

	t.Run("should marshal sum", func(t *testing.T) {
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

func writeData(t *testing.T, db *store.DB, data []byte) {
	writer, err := db.Writer()
	require.NoError(t, err)
	_, err = writer.Write(data)
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)
}

func readData(t *testing.T, db *store.DB) []byte {
	reader, err := db.Reader()
	require.NoError(t, err)
	actual, err := ioutil.ReadAll(reader)
	require.NoError(t, err)
	err = reader.Close()
	require.NoError(t, err)
	return actual
}