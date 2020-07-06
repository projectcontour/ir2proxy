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

package k8sdecoder

import (
	"fmt"
	"testing"
)

func TestDecodeIngressRoute(t *testing.T) {

	tests := map[string]struct {
		input []byte
		want  error
	}{
		"Minimal valid IngressRoute": {
			input: []byte(`
---
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata: 
  name: basic
spec: {}
`),
			want: nil,
		},
		"Not an IngressRoute": {
			input: []byte(`
---
apiVersion: v1
kind: Service
metadata:
  name: s1
spec:
  selector:
    app: kuard
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 8080
`),
			want: fmt.Errorf("Can only parse IngressRoute, a /v1, Kind=Service was supplied"),
		},
		"Invalid YAML": {
			input: []byte(`
---
# A comments-only YAML file
# won't parse.			
`),
			want: fmt.Errorf("Could not parse YAML, something something"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeIngressRoute(tc.input)
			if (err != nil) != (tc.want != nil) {
				t.Fatalf("want: %v, got: %v", tc.want, err)
			}
		})
	}
}
