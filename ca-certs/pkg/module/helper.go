// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ResourceName generates the name for a given resource
func ResourceName(name, kind string) string {
	return strings.ToLower(fmt.Sprintf("%s+%s", name, kind))
}

// GetKeyValue returns the value of a key from a json encoded struct
func GetValue(val interface{}) (string, error) {
	b, e := json.Marshal(val)
	if e != nil {
		return "", e
	}

	type temp struct {
		Value interface{} `json:"value"`
	}
	s := temp{}
	if e = json.Unmarshal(b, &s); e != nil {
		return "", errors.New("failed to parse the key value")
	}

	v, ok := s.Value.(string)
	if !ok {
		return "", errors.New("unexpected key value type")
	}

	return v, nil
}
