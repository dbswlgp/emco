// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package istiosubc

import (
	"gopkg.in/yaml.v2"
)

const SECURITY_ISTIO_APIVERSION = "security.istio.io/v1beta1"
const NETWORKING_ISTIO_KIND_AUTHORIZATIONPOLICY = "AuthorizationPolicy"

type IstioAuthorizationPolicyResource struct {
	ApiVersion string                  `yaml:"apiVersion"`
	Kind       string                  `yaml:"kind"`
	Metadata   Metadata                `yaml:"metadata"`
	Spec       AuthorizationPolicySpec `yaml:"spec"`
}

type AuthorizationPolicySpec struct {
	Action    string  `yaml:"action"`
	Rules     []Rule  `yaml:"rules"`
}

type Rule struct {
	To  []RuleTo  `yaml:"to"`
}

type RuleTo struct {
	Ope  map[string][]string `yaml:"operation"`
}

func createRule(ruleto []RuleTo) (Rule) {
	var r = Rule {
		To: ruleto,
	}
	return r
}

func createRuleTo(methods, paths, hosts []string) (RuleTo) {

	ope := make(map[string][]string, 3)
	ope["methods"] = methods
	ope["paths"] = paths
	ope["hosts"] = hosts
	var rt = RuleTo {
		Ope: ope,
	}
	return rt
}

func createAuthorizationPolicyResource(meta Metadata, spec AuthorizationPolicySpec) ([]byte, error) {
	var ro = IstioAuthorizationPolicyResource {
		ApiVersion: SECURITY_ISTIO_APIVERSION,
		Kind:       NETWORKING_ISTIO_KIND_AUTHORIZATIONPOLICY,
		Metadata:   meta,
		Spec:       spec,
	}

	y, err := yaml.Marshal(&ro)
	return y, err
}

func createAuthorizationPolicySpec(action string, rule []Rule) (AuthorizationPolicySpec){
	var apspec = AuthorizationPolicySpec {
		Action:   action,
		Rules:    rule,
	}
	return apspec
}

func createAuthPolicy(resname, namespace, action string, methods, paths, hosts []string, )([]byte, error) {
	meta := createGenericMetadata(resname, namespace, "")
	rt := createRuleTo(methods, paths, hosts)
	r := createRule([]RuleTo{rt,})
	spec := createAuthorizationPolicySpec(action, []Rule{r,})
	res, err := createAuthorizationPolicyResource(meta, spec)
	return res, err
}
