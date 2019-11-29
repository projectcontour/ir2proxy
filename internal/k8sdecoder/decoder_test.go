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
