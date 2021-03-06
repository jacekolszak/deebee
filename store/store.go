// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package store

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

func Open(dir string, options ...Option) (*Store, error) {
	if dir == "" {
		return nil, errors.New("dir is empty: must be a valid directory path")
	}

	s := &Store{
		dir: dir,
		areChecksumsEqual: func(expected, actual []byte) bool {
			return bytes.Equal(expected, actual) ||
				string(expected) == "ALTERED" || string(expected) == "ALTERED\n" || string(expected) == "ALTERED\r\n"
		},
	}

	for _, apply := range options {
		if apply == nil {
			continue
		}
		if err := apply(s); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	stat, err := os.Lstat(dir)
	switch {
	case os.IsNotExist(err):
		if s.failWhenMissingDir {
			return nil, fmt.Errorf("store directory %s does not exist", dir)
		}
		if mkdirErr := os.MkdirAll(dir, 0775); mkdirErr != nil {
			return nil, fmt.Errorf("mkdir failed for directory %s: %w", dir, mkdirErr)
		}
	case err != nil:
		return nil, fmt.Errorf("lstat failed for directory %s: %w", dir, err)
	case !stat.IsDir():
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	return s, nil
}

type Option func(s *Store) error

var FailWhenMissingDir Option = func(s *Store) error {
	s.failWhenMissingDir = true
	return nil
}

var NoIntegrityCheck Option = func(s *Store) error {
	s.areChecksumsEqual = func(expected, actual []byte) bool {
		return true
	}
	return nil
}

type Store struct {
	failWhenMissingDir bool
	areChecksumsEqual  func(expected, actual []byte) bool
	dir                string
	lastVersionTime    time.Time
	metrics            Metrics
}

func (s *Store) Reader(options ...ReaderOption) (Reader, error) {
	s.metrics.Read.ReaderCalls++

	return s.openReader(options, s.areChecksumsEqual)
}

type ReaderOption func(*ReaderOptions) error

func Time(t time.Time) ReaderOption {
	return func(o *ReaderOptions) error {
		o.chooseVersion = func(versions []Version) (Version, error) {
			for _, version := range versions {
				if version.Time.Equal(t) {
					return version, nil
				}
			}
			return Version{}, NewVersionNotFoundError(fmt.Sprintf("version %s not found", t))
		}
		return nil
	}
}

type Reader interface {
	io.ReadCloser
	Version() Version
}

func (s *Store) Writer(options ...WriterOption) (Writer, error) {
	s.metrics.Write.WriterCalls++

	return s.openWriter(options)
}

type WriterOption func(*WriterOptions) error

type WriterOptions struct {
	time time.Time
	sync func(*os.File) error
}

// WriteTime is not named Time to avoid name conflict with ReaderOption
func WriteTime(t time.Time) WriterOption {
	return func(o *WriterOptions) error {
		o.time = t
		return nil
	}
}

var NoSync WriterOption = func(o *WriterOptions) error {
	o.sync = func(file *os.File) error {
		return nil
	}
	return nil
}

type Writer interface {
	io.Writer
	// Close must be called to make version readable
	Close() error
	Version() Version
	// AbortAndClose aborts writing version. Version will not be available to read.
	AbortAndClose()
}

// Versions return slice sorted by time, oldest first
func (s *Store) Versions() ([]Version, error) {
	return s.versions()
}

type Version struct {
	// Time uniquely identifies version
	Time time.Time
	Size int64
}

func (s *Store) DeleteVersion(t time.Time) error {
	dataFile := s.dataFilename(t)
	checksumFile := checksumFileForDataFile(dataFile)

	for _, file := range []string{dataFile, checksumFile} {
		err := os.Remove(file)
		if os.IsNotExist(err) {
			return NewVersionNotFoundError(fmt.Sprintf("version %s does not exist", t))
		}
		if err != nil {
			return fmt.Errorf("error removing file %s: %w", file, err)
		}
	}
	return nil
}

func (s *Store) Metrics() Metrics {
	return s.metrics
}
