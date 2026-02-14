package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ArchitectureSuite struct {
	suite.Suite
	moduleRoot string
}

func (s *ArchitectureSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().NoError(err)
	s.moduleRoot = wd
}

func (s *ArchitectureSuite) TestForbiddenImports() {
	type rule struct {
		layer   string
		pattern string
		denied  []string
	}

	rules := []rule{
		{
			layer:   "domain",
			pattern: "internal/domain",
			denied: []string{
				"internal/service",
				"internal/infrastructure",
				"internal/transport",
				"internal/config",
				"database/sql",
				"net/http",
			},
		},
		{
			layer:   "service",
			pattern: "internal/service",
			denied: []string{
				"internal/infrastructure",
				"internal/transport",
				"database/sql",
				"net/http",
			},
		},
		{
			layer:   "infrastructure",
			pattern: "internal/infrastructure",
			denied: []string{
				"internal/service",
				"internal/transport",
			},
		},
		{
			layer:   "transport",
			pattern: "internal/transport",
			denied: []string{
				"internal/infrastructure",
			},
		},
	}

	for _, r := range rules {
		s.Run(r.layer, func() {
			layerDir := filepath.Join(s.moduleRoot, r.pattern)
			if _, err := os.Stat(layerDir); os.IsNotExist(err) {
				return
			}

			err := filepath.Walk(layerDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
					return err
				}
				if strings.HasSuffix(path, "_test.go") || strings.Contains(path, "sqlcgen") {
					return nil
				}

				fset := token.NewFileSet()
				f, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
				if parseErr != nil {
					return parseErr
				}

				rel, _ := filepath.Rel(s.moduleRoot, path)
				for _, imp := range f.Imports {
					importPath := strings.Trim(imp.Path.Value, "\"")
					for _, denied := range r.denied {
						if strings.Contains(importPath, denied) {
							s.Failf("forbidden import",
								"%s imports %q (layer %q must not depend on %q)",
								rel, importPath, r.layer, denied)
						}
					}
				}
				return nil
			})
			s.Require().NoError(err)
		})
	}
}

func (s *ArchitectureSuite) TestFileSizeLimit() {
	const maxLines = 500

	err := filepath.Walk(s.moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		if strings.Contains(path, "sqlcgen") || strings.Contains(path, "tests/mocks") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		lines := strings.Count(string(data), "\n") + 1
		if lines > maxLines {
			rel, _ := filepath.Rel(s.moduleRoot, path)
			s.Failf("file too large", "%s has %d lines (max %d)", rel, lines, maxLines)
		}
		return nil
	})
	s.Require().NoError(err)
}

func (s *ArchitectureSuite) TestAllPackagesHaveTests() {
	// Packages covered transitively by handler integration tests.
	skipDirs := map[string]bool{
		"sqlcgen":     true,
		"versions":    true,
		"cmd":         true,
		"config":      true,
		"dto":         true,
		"domain":      true,
		"migrations":  true,
		"persistence": true,
		"service":     true,
	}

	skipPaths := map[string]bool{
		"internal/transport/http": true,
	}

	internalDir := filepath.Join(s.moduleRoot, "internal")
	var packages []string

	err := filepath.Walk(internalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return err
		}
		if skipDirs[filepath.Base(path)] {
			return filepath.SkipDir
		}

		rel, _ := filepath.Rel(s.moduleRoot, path)
		if skipPaths[rel] {
			return nil
		}

		hasGo := false
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
				hasGo = true
				break
			}
		}
		if !hasGo {
			return nil
		}

		hasTest := false
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), "_test.go") {
				hasTest = true
				break
			}
		}

		if !hasTest {
			rel, _ := filepath.Rel(s.moduleRoot, path)
			packages = append(packages, rel)
		}
		return nil
	})
	s.Require().NoError(err)

	if len(packages) > 0 {
		s.Failf("packages without tests", "the following packages have no test files: %s",
			strings.Join(packages, ", "))
	}
}

func TestArchitectureSuite(t *testing.T) {
	suite.Run(t, new(ArchitectureSuite))
}
