package remover

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/exp/slices"
)

const (
	goExtension   = ".go"
	recursivePath = "./..."
)

var (
	currentPaths = []string{".", "." + string(filepath.Separator)}
)

var (
	ErrPathIsNotDir = errors.New("path is not a directory")
)

// SourceDir to validate and fix import
type SourceDir struct {
	dir           string
	isRecursive   bool
	isOutputOnly  bool
	generatedName string
	paths         *[]string
}

func NewSourceDir(path string, isRecursive, isOutputOnly bool, generatedName string) *SourceDir {
	if path == recursivePath {
		isRecursive = true
	}
	return &SourceDir{
		dir:           path,
		paths:         &[]string{},
		isRecursive:   isRecursive,
		isOutputOnly:  isOutputOnly,
		generatedName: generatedName,
	}
}

func (d *SourceDir) DeleteAutoGeneratedFiles() ([]string, error) {
	var ok bool
	d.dir, ok = IsDir(d.dir)
	if !ok {
		return nil, ErrPathIsNotDir
	}

	err := filepath.WalkDir(d.dir, d.walk(d.paths))
	if err != nil {
		return nil, fmt.Errorf("failed to walk dif: %w", err)
	}

	return *d.paths, nil
}

func (d *SourceDir) walk(paths *[]string) fs.WalkDirFunc {
	return func(path string, dirEntry fs.DirEntry, err error) error {
		if !d.isRecursive && dirEntry.IsDir() && filepath.Base(d.dir) != dirEntry.Name() {
			return filepath.SkipDir
		}
		if isGoFile(path) && !dirEntry.IsDir() {
			isAutoGen, err := NewSourceFile(path, d.generatedName).IsAutoGenerated()
			if err != nil {
				return err
			}

			if isAutoGen && !d.isOutputOnly {
				err = os.Remove(path)
				if err != nil {
					return err
				}
			} else if isAutoGen {
				*paths = append(*paths, path)
			}

		}
		return nil
	}
}

func IsDir(path string) (string, bool) {
	if path == recursivePath || slices.Contains(currentPaths, path) {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return path, false
		}
	}

	dir, err := os.Open(path)
	if err != nil {
		return path, false
	}

	dirStat, err := dir.Stat()
	if err != nil {
		return path, false
	}

	return path, dirStat.IsDir()
}

func isGoFile(path string) bool {
	return filepath.Ext(path) == goExtension
}
