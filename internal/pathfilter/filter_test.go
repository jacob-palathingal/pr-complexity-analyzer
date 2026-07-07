package pathfilter

import "testing"

func TestShouldAnalyzeExcludesCommonGeneratedAndDependencyPaths(t *testing.T) {
	testCases := []string{
		".git/config",
		"vendor/example.com/library/file.go",
		"node_modules/package/index.js",
		".venv/lib/python/site.py",
		"venv/lib/python/site.py",
		"dist/generated.py",
		"build/output.go",
		"coverage/result.py",
		"api/customer.generated.go",
		"api/customer_pb2.py",
		"web/application.min.js",
	}

	for _, filePath := range testCases {
		t.Run(filePath, func(t *testing.T) {
			if ShouldAnalyze(filePath, Options{}) {
				t.Fatalf(
					"expected %q to be excluded",
					filePath,
				)
			}
		})
	}
}

func TestShouldAnalyzeIncludesOrdinarySourceFiles(t *testing.T) {
	testCases := []string{
		"main.go",
		"service/handler.go",
		"scripts/analyze.py",
		"src/payments/processor.py",
	}

	for _, filePath := range testCases {
		t.Run(filePath, func(t *testing.T) {
			if !ShouldAnalyze(filePath, Options{}) {
				t.Fatalf(
					"expected %q to be included",
					filePath,
				)
			}
		})
	}
}

func TestShouldAnalyzeAppliesIncludeFilters(t *testing.T) {
	opts := Options{
		Include: []string{
			"services/payments/**",
		},
	}

	if !ShouldAnalyze(
		"services/payments/handler.go",
		opts,
	) {
		t.Fatal(
			"expected payments handler to match include filter",
		)
	}

	if ShouldAnalyze(
		"services/users/handler.go",
		opts,
	) {
		t.Fatal(
			"expected users handler to be rejected by include filter",
		)
	}
}

func TestShouldAnalyzeExclusionTakesPrecedenceOverInclusion(t *testing.T) {
	opts := Options{
		Include: []string{
			"services/payments/**",
		},
		Exclude: []string{
			"services/payments/mocks/**",
		},
	}

	if !ShouldAnalyze(
		"services/payments/handler.go",
		opts,
	) {
		t.Fatal(
			"expected payments handler to be included",
		)
	}

	if ShouldAnalyze(
		"services/payments/mocks/client.go",
		opts,
	) {
		t.Fatal(
			"expected exclusion to override inclusion",
		)
	}
}

func TestShouldAnalyzeExcludesTestsByDefault(t *testing.T) {
	testCases := []string{
		"service/handler_test.go",
		"service/handler_test.py",
		"test/handler.py",
		"tests/handler.py",
		"service/test/handler.py",
		"service/tests/handler.py",
	}

	for _, filePath := range testCases {
		t.Run(filePath, func(t *testing.T) {
			if ShouldAnalyze(filePath, Options{}) {
				t.Fatalf(
					"expected test file %q to be excluded",
					filePath,
				)
			}
		})
	}
}

func TestShouldAnalyzeIncludesTestsWhenRequested(t *testing.T) {
	opts := Options{
		IncludeTests: true,
	}

	testCases := []string{
		"service/handler_test.go",
		"service/handler_test.py",
		"test/handler.py",
		"tests/handler.py",
	}

	for _, filePath := range testCases {
		t.Run(filePath, func(t *testing.T) {
			if !ShouldAnalyze(filePath, opts) {
				t.Fatalf(
					"expected test file %q to be included",
					filePath,
				)
			}
		})
	}
}

func TestShouldAnalyzeNormalizesWindowsPaths(t *testing.T) {
	opts := Options{
		Include: []string{
			"services/payments/**",
		},
	}

	if !ShouldAnalyze(
		`services\payments\handler.go`,
		opts,
	) {
		t.Fatal(
			"expected Windows path separators to be normalized",
		)
	}
}

func TestShouldAnalyzeRejectsEmptyPath(t *testing.T) {
	if ShouldAnalyze("   ", Options{}) {
		t.Fatal("expected empty path to be rejected")
	}
}
