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

package validate

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/projectcontour/ir2proxy/internal/k8sdecoder"
)

func TestCheckIngressRoute(t *testing.T) {
	for name, tc := range buildFixtureSet(t) {
		t.Run(name, func(t *testing.T) {
			ir, _ := k8sdecoder.DecodeIngressRoute(tc.input)
			warnings := CheckIngressRoute(ir)
			diff := cmp.Diff(warnings, tc.want)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

type testFixture struct {
	input []byte
	want  []string
}

func buildFixtureSet(t *testing.T) map[string]testFixture {
	testdataFiles, err := ioutil.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	newFixtures := make(map[string]testFixture)
	for _, fileinfo := range testdataFiles {
		if !fileinfo.IsDir() {
			continue
		}

		inputdata, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s/input.yaml", fileinfo.Name()))
		if err != nil {
			t.Fatal(err)
		}

		errordata, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s/errors.txt", fileinfo.Name()))
		if err != nil {
			t.Fatal(err)
		}

		trimerrordata := strings.TrimSpace(string(errordata))
		var errors []string
		if len(trimerrordata) > 0 {
			errors = strings.Split(trimerrordata, "\n")
		}

		newFixtures[fileinfo.Name()] = testFixture{
			input: inputdata,
			want:  errors,
		}
	}

	return newFixtures

}
