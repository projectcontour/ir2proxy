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

// Package translator translates IngressRoute objects to HTTPProxy ones
package translator

import (
	"errors"
	"fmt"

	irv1beta1 "github.com/projectcontour/contour/apis/contour/v1beta1"
	hpv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressRouteToHTTPProxy translates IngressRoute objects to HTTPProxy ones, emitting warnings
// as it goes.
// There are currently no fatal conditions (that should not produces a HTTPProxy output)
// TODO(youngnick) - change this signature to return HTTPProxy, []string, error if we need that.
func IngressRouteToHTTPProxy(ir *irv1beta1.IngressRoute) (*hpv1.HTTPProxy, []string, error) {

	var warnings []string
	// TODO(youngnick): Investigate if we should skip logically empty IngressRoutes

	if ir.Spec.VirtualHost == nil {
		return nil, warnings, errors.New("unimplemented: Can't translate non-root IngressRoutes yet")
	}

	hpRoutes, hpIncludes, hpWarnings := translateRoutes(ir.Spec.Routes)

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
		Spec: hpv1.HTTPProxySpec{
			VirtualHost: ir.Spec.VirtualHost,
		},
	}

	hp.Spec.Routes = hpRoutes
	hp.Spec.Includes = hpIncludes
	warnings = append(warnings, hpWarnings...)
	return hp, warnings, nil
}

func translateRoute(irRoute irv1beta1.Route) (hpv1.Route, []string) {
	var warnings []string

	hpRoute := hpv1.Route{
		Conditions: []hpv1.Condition{
			hpv1.Condition{
				Prefix: irRoute.Match,
			},
		},
	}

	if irRoute.TimeoutPolicy != nil {
		hpRoute.TimeoutPolicy = &hpv1.TimeoutPolicy{
			Response: irRoute.TimeoutPolicy.Request,
		}
	}
	var seenLBStrategy string
	for _, irService := range irRoute.Services {

		hpService := hpv1.Service{
			Name:   irService.Name,
			Port:   irService.Port,
			Weight: irService.Weight,
		}

		if irService.Strategy != "" {
			if seenLBStrategy == "" {
				seenLBStrategy = irService.Strategy
				// Copy the first strategy we encounter into the HP loadbalancerpolicy
				// and save that we've seen that one.
				hpRoute.LoadBalancerPolicy = &hpv1.LoadBalancerPolicy{
					Strategy: irService.Strategy,
				}
			} else {
				if seenLBStrategy != irService.Strategy {
					warnings = append(warnings, fmt.Sprintf("Strategy %s on Service %s could not be applied, HTTPProxy only supports a single load balancing policy across all services. %s is already applied.", irService.Strategy, irService.Name, seenLBStrategy))
				}
			}

		}
		if irService.HealthCheck != nil {
			hpRoute.HealthCheckPolicy = &hpv1.HTTPHealthCheckPolicy{
				Path:                    irService.HealthCheck.Path,
				Host:                    irService.HealthCheck.Host,
				TimeoutSeconds:          irService.HealthCheck.TimeoutSeconds,
				UnhealthyThresholdCount: irService.HealthCheck.UnhealthyThresholdCount,
				HealthyThresholdCount:   irService.HealthCheck.HealthyThresholdCount,
			}
		}
		hpRoute.Services = append(hpRoute.Services, hpService)
	}

	return hpRoute, warnings
}

func translateInclude(irRoute irv1beta1.Route) *hpv1.Include {
	var hpInclude *hpv1.Include

	if irRoute.Delegate != nil {
		// If there's a delegation, short-circuit.
		hpInclude = &hpv1.Include{
			Conditions: []hpv1.Condition{
				hpv1.Condition{
					Prefix: irRoute.Match,
				},
			},
			Name:      irRoute.Delegate.Name,
			Namespace: irRoute.Delegate.Namespace,
		}
	}

	return hpInclude
}

func translateRoutes(irRoutes []irv1beta1.Route) ([]hpv1.Route, []hpv1.Include, []string) {

	var hpRoutes []hpv1.Route
	var hpIncludes []hpv1.Include
	var warnings []string
	for _, irRoute := range irRoutes {
		hpInclude := translateInclude(irRoute)
		if hpInclude != nil {
			hpIncludes = append(hpIncludes, *hpInclude)
			continue
		}
		hpRoute, routeWarnings := translateRoute(irRoute)
		hpRoutes = append(hpRoutes, hpRoute)
		warnings = append(warnings, routeWarnings...)
	}

	return hpRoutes, hpIncludes, warnings
}
