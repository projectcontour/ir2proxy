package translator

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

			if warnings == nil && len(tc.warnings) > 0 {
				t.Fatalf("Expected translation warnings not present:\n%+v", tc.warnings)
			}

			if warnings != nil {
				errorsDiff := cmp.Diff(warnings, tc.warnings)
				if errorsDiff != "" {
					t.Fatalf("Translation Errors:\n%v\n", errorsDiff)
				}
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
			t.Fatalf("testdata/%s/input.yaml must be present, %s", fileinfo.Name(), err)
		}

		fixture := testFixture{
			input: inputdata,
		}

		output, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s/output.yaml", fileinfo.Name()))
		if err != nil {
			t.Fatalf("testdata/%s/output.yaml must be present, %s", fileinfo.Name(), err)
		}
		fixture.output = output

		errordata, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s/errors.txt", fileinfo.Name()))
		if err != nil {
			t.Fatalf("testdata/%s/errors.txt must be present, %s", fileinfo.Name(), err)
		}

		// strings.Split on a zero-length byte slice returns
		// []string{""}, not []string{}. Clean out any empty strings.
		var cleanerrors []string
		for _, error := range strings.Split(string(errordata), "\n") {
			if error != "" {
				cleanerrors = append(cleanerrors, error)
			}
		}
		fixture.warnings = cleanerrors

		fixtures[fileinfo.Name()] = fixture

	}

	return fixtures

}
