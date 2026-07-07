package pathfilter

import (
	"path"
	"strings"
)

var defaultExcludes = []string{
	".git/**",
	"vendor/**",
	"node_modules/**",
	".venv/**",
	"venv/**",
	"dist/**",
	"build/**",
	"coverage/**",
	"**/*.generated.go",
	"**/*_pb2.py",
	"**/*.min.js",
}

// Options controls path-level filtering before file contents are loaded.
type Options struct {
	Include      []string
	Exclude      []string
	IncludeTests bool
}

// ShouldAnalyze returns true when a path should be considered for analysis.
//
// Exclusions take precedence over inclusions. Tests and common generated or
// dependency directories are excluded by default.
func ShouldAnalyze(filePath string, opts Options) bool {
	filePath = normalize(filePath)

	if filePath == "" {
		return false
	}

	if !opts.IncludeTests && isTestFile(filePath) {
		return false
	}

	excludes := make(
		[]string,
		0,
		len(defaultExcludes)+len(opts.Exclude),
	)
	excludes = append(excludes, defaultExcludes...)
	excludes = append(excludes, opts.Exclude...)

	for _, pattern := range excludes {
		if match(pattern, filePath) {
			return false
		}
	}

	if len(opts.Include) == 0 {
		return true
	}

	for _, pattern := range opts.Include {
		if match(pattern, filePath) {
			return true
		}
	}

	return false
}

func isTestFile(filePath string) bool {
	baseName := path.Base(filePath)

	if strings.HasSuffix(baseName, "_test.go") ||
		strings.HasSuffix(baseName, "_test.py") {
		return true
	}

	segments := strings.Split(filePath, "/")
	for _, segment := range segments[:len(segments)-1] {
		if segment == "test" || segment == "tests" {
			return true
		}
	}

	return false
}

func normalize(value string) string {
	value = strings.ReplaceAll(value, "\\", "/")
	value = strings.TrimSpace(value)

	for strings.HasPrefix(value, "./") {
		value = strings.TrimPrefix(value, "./")
	}

	return value
}

func match(pattern, filePath string) bool {
	pattern = normalize(pattern)
	filePath = normalize(filePath)

	if pattern == "" {
		return false
	}

	if pattern == filePath {
		return true
	}

	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")

		return filePath == prefix ||
			strings.HasPrefix(filePath, prefix+"/")
	}

	if strings.HasPrefix(pattern, "**/") {
		suffix := strings.TrimPrefix(pattern, "**/")

		if matched, _ := path.Match(
			suffix,
			path.Base(filePath),
		); matched {
			return true
		}

		return strings.HasSuffix(filePath, "/"+suffix)
	}

	matched, _ := path.Match(pattern, filePath)
	return matched
}
