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
