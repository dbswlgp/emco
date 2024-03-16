// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package istiosubc

type HTTPRoute struct {
	Name    string `yaml:"name,omitempty"`
	Match   []HTTPMatchRequest `yaml:"match,omitempty"`
	Route   []HTTPRouteDestination   `yaml:"route,omitempty"`
}

type StringMatchPrefix struct {
	Prefix string `yaml:"prefix,omitempty"`
}
type MatchPort struct {
	Port uint32 `yaml:"port,omitempty"`
}

type HTTPMatchRequest struct {
	Port uint32 `yaml:"port,omitempty"`
}

type HTTPRouteDestination struct {
	Dest Destination `yaml:"destination,omitempty"`
}

type Destination struct {
	Host string `yaml:"host"`
	Subset string `yaml:"subset,omitempty"`
	Port   PortSelector `yaml:"port,omitempty"`
}

type PortSelector struct {
	Number uint32 `yaml:"number,omitempty"`
}

func createPortSelector(number uint32)(PortSelector) {
	var portSelector = PortSelector {
		Number: number,
	}
	return portSelector
}

func createDestination(host string, port PortSelector)(Destination) {
	var dest = Destination {
		Host: host,
		Port: port,
	}
	return dest
}

func createHTTPMatchPrefix(value string)(StringMatchPrefix) {
	var mprefix = StringMatchPrefix {
		Prefix: value,
	}
	return mprefix
}

func createHTTPRouteDestination(dest Destination)(HTTPRouteDestination) {
	var httpRoutedest = HTTPRouteDestination {
		Dest:   dest,
	}
	return httpRoutedest
}

func createHTTPRoute(name string, route []HTTPRouteDestination, match []HTTPMatchRequest)(HTTPRoute) {
	var httpRoute = HTTPRoute {
		Name: name,
		Match: match,
		Route: route,
	}

	return httpRoute
}
