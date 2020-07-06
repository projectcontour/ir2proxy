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

// Package k8sdecoder decodes YAML []bytes into IngressRoute objects
package k8sdecoder

import (
	"fmt"

	irv1beta1 "github.com/projectcontour/contour/apis/contour/v1beta1"
	contourscheme "github.com/projectcontour/contour/apis/generated/clientset/versioned/scheme"
	"k8s.io/client-go/kubernetes/scheme"
)

// DecodeIngressRoute decodes a given byte stream into a IngressRoute or returns an error.
func DecodeIngressRoute(input []byte) (*irv1beta1.IngressRoute, error) {
	contourscheme.AddToScheme(scheme.Scheme)
	decode := scheme.Codecs.UniversalDeserializer().Decode
	ir, groupVersionKind, err := decode(input, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not parse yaml, %s", err)
	}
	switch t := ir.(type) {
	case *irv1beta1.IngressRoute:
		return t, nil
	default:
		return nil, fmt.Errorf("can only parse IngressRoute, a %s was supplied", groupVersionKind)
	}

}
