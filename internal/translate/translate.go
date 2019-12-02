// Copyright © 2019 VMware
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

// Package translate translates IngressRoute objects to HTTPProxy ones
package translate

import (
	irv1beta1 "github.com/projectcontour/contour/apis/contour/v1beta1"
	hpv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressRouteToHTTPProxy translates IngressRoute objects to HTTPProxy ones, emitting warnings
// as it goes.
func IngressRouteToHTTPProxy(ir *irv1beta1.IngressRoute) (*hpv1.HTTPProxy, []string) {

	var warnings []string
	hp := &hpv1.HTTPProxy{
		TypeMeta: v1.TypeMeta{
			Kind:       "HTTPProxy",
			APIVersion: "projectcontour.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			// TODO(youngnick): For some reason, the zero value of the CreationTimestamp field
			// is output as 'null' by the JSON/YAML serializer, even though it's set to 'omitempty'.
			// ¯\_(ツ)_/¯
			// Doesn't stop the objects being applied, but it is weird.
			Name:        ir.ObjectMeta.Name,
			Namespace:   ir.ObjectMeta.Namespace,
			Labels:      ir.ObjectMeta.DeepCopy().GetLabels(),
			Annotations: ir.ObjectMeta.DeepCopy().GetAnnotations(),
		},
	}

	if ir.Spec.VirtualHost == nil && ir.Spec.TCPProxy == nil && len(ir.Spec.Routes) == 0 {
		warnings = append(warnings, "Ingress %s is empty. Not much to convert.", ir.ObjectMeta.Name)
	}
	if ir.Spec.VirtualHost != nil {
		hp.Spec.VirtualHost = ir.Spec.VirtualHost
	}

	return hp, warnings
}
