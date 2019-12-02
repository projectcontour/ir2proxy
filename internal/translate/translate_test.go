package translate

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/projectcontour/ir2proxy/internal/k8sdecoder"
)

// testFixture holds test fixtures
// The expected contract is that if there is no errors.txt, then no errors are
// expected, and the translation should succeed.
type testFixture struct {
	input    []byte
	output   []byte
	warnings []string
}

// In this test, we're testing six conditions:
// IngressRoute cannot be translated, there are warnings,
func TestTranslateIngressRoute(t *testing.T) {
	for name, tc := range buildFixtureSet(t) {
		t.Run(name, func(t *testing.T) {

			ir, err := k8sdecoder.DecodeIngressRoute(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			hp, warnings := IngressRouteToHTTPProxy(ir)
			outputYAML, err := yaml.Marshal(hp)
			if err != nil {
				t.Fatal(err)
			}

			// hp being nil means there was a fatal error in translation.
			if hp == nil {
				if tc.warnings != nil {
					errorsDiff := cmp.Diff(warnings, tc.warnings)
					if errorsDiff != "" {
						// The errors weren't what we expected.
						t.Fatalf("Translation Errors Mismatch:\n%v", errorsDiff)
					}
					// httpproxy is nil, but the errors were expected, pass.
					return
				}
				// Show the errors in this case, so you can set up your test case correctly.
				t.Fatalf("HTTProxy is nil, Translation Errors:\n%v", warnings)
			}
			if tc.output != nil {
				translateDiff := cmp.Diff(bytes.TrimSpace(outputYAML), bytes.TrimSpace(tc.output))
				if translateDiff != "" {
					// We got a translation error
					if tc.warnings != nil {
						// We're supposed to get some errors with this translation error
						errorsDiff := cmp.Diff(warnings, tc.warnings)
						if errorsDiff != "" {
							// Translation failed, and the error wasn't what we expected.
							t.Fatalf("\nTranslation failure:\n%v\nTranslation Errors Mismatch:\n%v", translateDiff, errorsDiff)
						}
					}
					// Translation failed, we're not supposed to get any errors, log any errors we got in case.
					t.Fatalf("\nUnexpected translation failure:\n%v, Errors:\n%v", translateDiff, warnings)

				}
			}
			// Last case, if there is supposed to be no errors, and we got some, fail.
			if tc.warnings == nil && len(warnings) > 0 {
				t.Fatalf("Translation Errors:\n%v", warnings)
			}
		})
	}
}

func buildFixtureSet(t *testing.T) map[string]testFixture {
	testdataFiles, err := ioutil.ReadDir("testdata")
	if err != nil {
		panic(err)
	}
	fixtures := make(map[string]testFixture)
	for _, fileinfo := range testdataFiles {
		if !fileinfo.IsDir() {
			continue
		}

		inputdata, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s/input.yaml", fileinfo.Name()))
		if err != nil {
			t.Fatal(err)
		}

		fixture := testFixture{
			input: inputdata,
		}

		output, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s/output.yaml", fileinfo.Name()))
		if err != nil {
			// If we can't open output.yaml, then we need to signal that with a nil pointer.
			fixture.output = nil
		} else {
			fixture.output = output
		}

		errordata, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s/errors.txt", fileinfo.Name()))
		if err != nil {
			// If we can't open errors.txt, we don't have any errors defined.
			// Signaled by nil for warnings.
			fixture.warnings = nil
			fixtures[fileinfo.Name()] = fixture
			continue
		}

		errors := strings.Split(string(errordata), "\n")

		fixtures[fileinfo.Name()] = testFixture{
			input:    inputdata,
			output:   output,
			warnings: errors,
		}
	}

	return fixtures

}
