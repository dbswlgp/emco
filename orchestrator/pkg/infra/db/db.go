// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package db

// DbInfo holds the DB collection and attributes info
type DbInfo struct {
	StoreName string // name of the db collection to use for client documents
	TagMeta   string // attribute key name for the json data of a client document
	TagState  string // attribute key name for the resource state
}
