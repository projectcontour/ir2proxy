// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func TestTranslateIngressRoute(t *testing.T) {
	for name, tc := range buildFixtureSet(t) {
		t.Run(name, func(t *testing.T) {

			ir, err := k8sdecoder.DecodeIngressRoute(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			hp, warnings, err := IngressRouteToHTTPProxy(ir)
			if err != nil {
				t.Fatal(err)
			}
			outputYAML, err := yaml.Marshal(hp)
			if err != nil {
				t.Fatal(err)
			}
			outputYAML = append([]byte("---\n"), outputYAML...)
			// The Kubernetes standard header field `currentTimestamp` serializes weirdly,
			// so filter it out.
			// See https://github.com/projectcontour/ir2proxy/issues/8 for more explanation here.
			outputYAML = bytes.ReplaceAll(outputYAML, []byte("  creationTimestamp: null\n"), []byte(""))

			translateDiff := cmp.Diff(bytes.TrimSpace(outputYAML), bytes.TrimSpace(tc.output))
			if translateDiff != "" {
				// We got a translation error
				if tc.warnings != nil {
					// We're supposed to get some errors with this translation error
					errorsDiff := cmp.Diff(warnings, tc.warnings)
					if errorsDiff != "" {
						// Translation failed, and the error wasn't what we expected.
						t.Fatalf("\nTranslation failure:\n%v\nTranslation Warnings Mismatch:\n%v", translateDiff, errorsDiff)
					}
				}
				// Translation failed, we're not supposed to get any errors, log any errors we got in case.
				t.Fatalf("\nUnexpected translation failure:\n%v, Warnings:\n%v", translateDiff, warnings)

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

func TestTranslateIngressRouteErrors(t *testing.T) {

	tests := map[string]struct {
		input []byte
		want  []string
	}{
		"nonroot IR, multiple invalid paths": {
			input: []byte(`
---
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata:
  name: nonroot-invalid-badpaths
  namespace: default
spec:
  routes:
    - match: foo
      services:
        - name: s1
          port: 80
    - match: bar
      services:
        - name: s2
          port: 80
`),
			want: []string{"invalid IngressRoute: match clauses must share a common prefix"},
		},
		"nonroot IR, multiple nonmatching prefixes": {
			input: []byte(`
---
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata:
  name: nonroot-invalid-badpaths
  namespace: default
spec:
  routes:
    - match: /foo
      services:
        - name: s1
          port: 80
    - match: /bar
      services:
        - name: s2
          port: 80
`),
			want: []string{"invalid IngressRoute: match clauses must share a common prefix"},
		},
		"tcpproxy IR, not root": {
			input: []byte(`
---
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata:
  name: nonroot-tcpproxy
  namespace: default
spec:
  tcpproxy:
     services:
     - name: s1
       port: 80
`),
			want: []string{"invalid IngressRoute: tcpproxy must be in a root IngressRoute"},
		},
		"tcpproxy IR, delegate and services set": {
			input: []byte(`
---
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata:
  name: nonroot-tcpproxy
  namespace: default
spec:
  virtualhost:
    fqdn: "tcpproxy-test.domain.com"
    tls:
      secretName: "secret"
  tcpproxy:
    delegate:
      name: bad
      namespace: default
    services:
      - name: s1
        port: 80
`),
			want: []string{"invalid IngressRoute: Delegate and Services can not both be set"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ir, err := k8sdecoder.DecodeIngressRoute(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			_, _, err = IngressRouteToHTTPProxy(ir)
			if err != nil {
				// Can't translate the IngressRoute at all
				// errors.txt should have the error message.
				errorDiff := cmp.Diff([]string{err.Error()}, tc.want)
				if errorDiff != "" {
					t.Fatalf("Expected translation error not encountered:\n%v", errorDiff)
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

func TestLongestCommonPathPrefix(t *testing.T) {

	tests := map[string]struct {
		input []string
		want  string
	}{
		"nil": {
			input: nil,
			want:  "",
		},
		"single empty string": {
			input: []string{""},
			want:  "",
		},
		"single slash": {
			input: []string{"/"},
			want:  "",
		},
		"single entry": {
			input: []string{"/foo"},
			want:  "/foo",
		},
		"two different no common prefix": {
			input: []string{"/foo", "/bar"},
			want:  "",
		},
		"two full match": {
			input: []string{"/foo", "/foo"},
			want:  "/foo",
		},
		"two with shared prefix": {
			input: []string{"/foo/bar", "/foo/baz"},
			want:  "/foo",
		},
		"three full match": {
			input: []string{"/foo", "/foo", "/foo"},
			want:  "/foo",
		},
		"three, two components shared": {
			input: []string{"/foo/bar/baz", "/foo/bar/quux", "/foo/bar/bar"},
			want:  "/foo/bar",
		},
		"three, all different": {
			input: []string{"/foo", "/bar", "/baz"},
			want:  "",
		},
		"first is longest": {
			input: []string{"/long/path/first", "/long"},
			want:  "/long",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := longestCommonPathPrefix(tc.input)
			if got != tc.want {
				t.Fatalf("expected: '%v', got '%v'", tc.want, got)
			}
		})
	}
}
