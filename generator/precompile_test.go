package generator_test

import (
	"regexp"
	"testing"
)

var (
	usedCacheIndexRegexp   = regexp.MustCompile(`CacheIndex: (\d+),`)
	precompiledIndexRegexp = regexp.MustCompile(`\{Index: (\d+), Expr:`)
)

// assertPrecompileCoversCacheIndexes checks that every cache index used by the generated code
// has an entry in the precompile table, otherwise the expression would silently be compiled lazily.
func assertPrecompileCoversCacheIndexes(t *testing.T, generated string) {
	t.Helper()

	precompiled := map[string]struct{}{}
	for _, m := range precompiledIndexRegexp.FindAllStringSubmatch(generated, -1) {
		precompiled[m[1]] = struct{}{}
	}
	for _, m := range usedCacheIndexRegexp.FindAllStringSubmatch(generated, -1) {
		if _, exists := precompiled[m[1]]; !exists {
			t.Errorf("cache index %s is used but is not precompiled", m[1])
		}
	}
}
