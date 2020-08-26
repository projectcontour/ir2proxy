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

// Package translator translates IngressRoute objects to HTTPProxy ones
package translator

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	irv1beta1 "github.com/projectcontour/contour/apis/contour/v1beta1"
	hpv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressRouteToHTTPProxy translates IngressRoute objects to HTTPProxy ones, emitting warnings
// as it goes.
// There are currently no fatal conditions (that should not produces a HTTPProxy output)
// TODO(youngnick) - change this signature to return HTTPProxy, []string, error if we need that.
func IngressRouteToHTTPProxy(ir *irv1beta1.IngressRoute) (*hpv1.HTTPProxy, []string, error) {

	// TODO(youngnick): Investigate if we should skip logically empty IngressRoutes

	var routeLCP string
	var warnings []string

	var tcpproxy *hpv1.TCPProxy

	var includes []hpv1.Include

	if ir.Spec.TCPProxy != nil {
		if ir.Spec.VirtualHost == nil {
			return nil, nil, errors.New("invalid IngressRoute: tcpproxy must be in a root IngressRoute")
		}

		// The compiler won't use the outer tcpproxy correctly if we
		// use := here.
		var tcpincludes []hpv1.Include
		var err error
		var tcpwarnings []string
		tcpproxy, tcpincludes, tcpwarnings, err = translateTCPProxy(ir.Spec.TCPProxy)
		if err != nil {
			return nil, nil, err
		}
		includes = append(includes, tcpincludes...)
		warnings = append(warnings, tcpwarnings...)
	}

	if ir.Spec.VirtualHost == nil {
		routePrefixes := extractPrefixes(ir.Spec.Routes)
		routeLCP = longestCommonPathPrefix(routePrefixes)
		if routeLCP == "" && len(routePrefixes) > 1 {
			// There are no common prefixes here.
			return nil, nil, errors.New("invalid IngressRoute: match clauses must share a common prefix")
		}
		if len(routePrefixes) == 1 && routePrefixes[0] != "/" {
			warnings = append(warnings, fmt.Sprintf("Can't determine include path from single match %s. HTTPProxy prefix conditions should not include the include prefix. Please check this value is correct. See https://projectcontour.io/docs/main/httpproxy/#conditions-and-inclusion", routePrefixes[0]))
			// Reset the largest common prefix back to '/', since we can't replace it.
			routeLCP = ""
		}
		if routeLCP != "" {
			warnings = append(warnings, fmt.Sprintf("The guess for the IngressRoute include path is %s. HTTPProxy prefix conditions should not include the include prefix. Please check this value is correct. See https://projectcontour.io/docs/main/httpproxy/#conditions-and-inclusion", routeLCP))
		}

	}

	routes, routeIncludes, translateWarnings := translateRoutes(ir.Spec.Routes, routeLCP)
	includes = append(includes, routeIncludes...)
	warnings = append(warnings, translateWarnings...)

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
			// This field is filtered out of the marshaled YAML before it's output.
			Name:        ir.ObjectMeta.Name,
			Namespace:   ir.ObjectMeta.Namespace,
			Labels:      ir.ObjectMeta.DeepCopy().GetLabels(),
			Annotations: ir.ObjectMeta.DeepCopy().GetAnnotations(),
		},
		Spec: hpv1.HTTPProxySpec{
			VirtualHost: ir.Spec.VirtualHost,
			Routes:      routes,
			Includes:    includes,
			TCPProxy:    tcpproxy,
		},
	}

	return hp, warnings, nil
}

func translateRoute(irRoute irv1beta1.Route, routeLCP string) (hpv1.Route, []string) {

	var warnings []string

	route := hpv1.Route{
		Conditions: []hpv1.Condition{
			hpv1.Condition{},
		},
	}

	// If we've been passed a largest common prefix for all the routes, trim it
	// off the Match.
	// Note that the empty string for routeLCP here means "no prefix".
	match := irRoute.Match
	if routeLCP != "" {
		match = strings.TrimPrefix(match, routeLCP)
	}
	route.Conditions[0].Prefix = match

	if irRoute.TimeoutPolicy != nil {
		route.TimeoutPolicy = &hpv1.TimeoutPolicy{
			Response: irRoute.TimeoutPolicy.Request,
		}
	}

	if irRoute.PrefixRewrite != "" {
		route.PathRewritePolicy = &hpv1.PathRewritePolicy{
			ReplacePrefix: []hpv1.ReplacePrefix{
				hpv1.ReplacePrefix{
					Replacement: irRoute.PrefixRewrite,
				},
			},
		}
	}

	var seenLBStrategy string
	var seenHealthCheckPolicy *irv1beta1.HealthCheck
	var seenHealthCheckServiceName string
	for _, irService := range irRoute.Services {

		service, healthcheckPolicy, lbpolicy := translateService(irService)

		if lbpolicy != nil {
			if seenLBStrategy == "" {
				// Copy the first strategy we encounter into the HP loadbalancerpolicy
				// and save that we've seen that one.
				seenLBStrategy = irService.Strategy
				route.LoadBalancerPolicy = lbpolicy
			} else {
				if seenLBStrategy != irService.Strategy {
					warnings = append(warnings, fmt.Sprintf("Strategy %s on Service %s could not be applied, HTTPProxy only supports a single load balancing policy across all services. %s is already applied.", irService.Strategy, irService.Name, seenLBStrategy))
				}
			}
		}

		if healthcheckPolicy != nil {
			if seenHealthCheckPolicy == nil {
				// Copy the first strategy we encounter into the HP HealthCheckPolicy
				// and save that we've seen that one.
				seenHealthCheckPolicy = irService.HealthCheck
				seenHealthCheckServiceName = irService.Name
				route.HealthCheckPolicy = healthcheckPolicy
			} else {
				warnings = append(warnings, fmt.Sprintf("A healthcheck on service %s could not be applied, HTTPProxy only supports a single healthcheck across all services. A different healthcheck from service %s is already applied.", irService.Name, seenHealthCheckServiceName))
			}
		}

		route.Services = append(route.Services, service)
	}

	return route, warnings
}

func translateService(irService irv1beta1.Service) (hpv1.Service, *hpv1.HTTPHealthCheckPolicy, *hpv1.LoadBalancerPolicy) {
	service := hpv1.Service{
		Name:   irService.Name,
		Port:   irService.Port,
		Weight: irService.Weight,
	}

	var healthcheckPolicy *hpv1.HTTPHealthCheckPolicy
	var lbpolicy *hpv1.LoadBalancerPolicy

	if irService.Strategy != "" {
		lbpolicy = &hpv1.LoadBalancerPolicy{
			Strategy: irService.Strategy,
		}
	}

	if irService.HealthCheck != nil {
		healthcheckPolicy = &hpv1.HTTPHealthCheckPolicy{
			Path:                    irService.HealthCheck.Path,
			Host:                    irService.HealthCheck.Host,
			TimeoutSeconds:          irService.HealthCheck.TimeoutSeconds,
			UnhealthyThresholdCount: irService.HealthCheck.UnhealthyThresholdCount,
			HealthyThresholdCount:   irService.HealthCheck.HealthyThresholdCount,
		}
	}

	return service, healthcheckPolicy, lbpolicy
}

func translateInclude(irRoute irv1beta1.Route) *hpv1.Include {

	if irRoute.Delegate == nil {
		return nil
	}

	return &hpv1.Include{
		Conditions: []hpv1.Condition{
			hpv1.Condition{
				Prefix: irRoute.Match,
			},
		},
		Name:      irRoute.Delegate.Name,
		Namespace: irRoute.Delegate.Namespace,
	}
}

func translateRoutes(irRoutes []irv1beta1.Route, routeLCP string) ([]hpv1.Route, []hpv1.Include, []string) {

	var routes []hpv1.Route
	var includes []hpv1.Include
	var warnings []string
	for _, irRoute := range irRoutes {
		hpInclude := translateInclude(irRoute)
		if hpInclude != nil {
			includes = append(includes, *hpInclude)
			continue
		}
		route, translationWarnings := translateRoute(irRoute, routeLCP)
		routes = append(routes, route)
		warnings = append(warnings, translationWarnings...)
	}

	return routes, includes, warnings
}

func translateTCPProxy(irTCPProxy *irv1beta1.TCPProxy) (*hpv1.TCPProxy, []hpv1.Include, []string, error) {

	var includes []hpv1.Include
	var warnings []string

	if irTCPProxy.Delegate != nil {
		if len(irTCPProxy.Services) > 0 {
			return nil, includes, warnings, errors.New("invalid IngressRoute: Delegate and Services can not both be set")
		}
		includes = append(includes, hpv1.Include{
			Name:      irTCPProxy.Delegate.Name,
			Namespace: irTCPProxy.Delegate.Namespace,
		})
		return nil, includes, warnings, nil
	}

	proxy := &hpv1.TCPProxy{}
	for _, irService := range irTCPProxy.Services {

		hpService, healthcheckPolicy, lbpolicy := translateService(irService)

		if healthcheckPolicy != nil {
			warnings = append(warnings, "Healthcheck policy of TCPProxy service has no effect, discarding")
		}

		if lbpolicy != nil {
			proxy.LoadBalancerPolicy = lbpolicy
		}
		proxy.Services = append(proxy.Services, hpService)
	}
	return proxy, includes, warnings, nil
}

func extractPrefixes(routes []irv1beta1.Route) []string {

	var prefixes []string
	for _, route := range routes {
		prefixes = append(prefixes, route.Match)
	}

	return prefixes
}

// longestCommonPathPrefix finds the longest common path prefix by
// splitting a set of strings that give path prefixes on `/` characters,
// then checking which match.
// The empty string means that there is no common path prefix.
func longestCommonPathPrefix(paths []string) string {

	if len(paths) == 0 {
		return ""
	}

	if len(paths) == 1 {
		if paths[0] == "" || paths[0] == "/" {
			return ""
		}
		return paths[0]
	}

	if !sort.StringsAreSorted(paths) {
		sort.Strings(paths)
	}

	// Build a two-dimensional array of paths split by "/"
	// the first element of pathElements will be the shortest path
	pathElements := make([][]string, len(paths))
	for index, path := range paths {
		// Split the first '/' off, to remove the zero-length
		// string that would otherwise be the first element.
		if path[0] == '/' {
			path = path[1:]
		}
		pathElements[index] = strings.Split(path, "/")
	}

	// Next, for each element in the shortest path,
	// check if all the elements in that position in the other
	// paths match. If so, it's common, add it.
	// If not, that's it, break.
	var longestPrefix []string

OuterLoop:
	for index, pathElement := range pathElements[0] {
		for _, pathSlice := range pathElements[1:] {
			if pathSlice[index] != pathElement {
				break OuterLoop
			}
		}
		longestPrefix = append(longestPrefix, pathElement)
	}

	// If there isn't any longest prefix, just return "/"
	if len(longestPrefix) == 0 {
		return ""
	}

	return fmt.Sprintf("/%s", strings.Join(longestPrefix, "/"))

}
