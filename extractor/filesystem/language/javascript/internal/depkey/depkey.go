// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package depkey resolves a dependency declared by one lockfile entry to the
// key of the installed instance, following npm's nested resolution rules: a
// name is looked up in the depending package's own install directory first,
// then in each enclosing one, and finally at the top level.
package depkey

import "strings"

// Resolver walks the install tree described by a set of lockfile keys.
type Resolver struct {
	// Segment is the directory that separates an enclosing key from a nested
	// install. bun.lock nests keys directly ("a/b"), so it is empty;
	// package-lock.json v2+ nests through node_modules
	// ("node_modules/a/node_modules/b"), so it is "node_modules/".
	Segment string
	// Keys is the set of lockfile keys to resolve against.
	Keys map[string]bool
}

// Resolve returns the key of the instance of depName that the package at
// pkgKey resolves to, walking outwards from pkgKey to the top level. It
// reports false if no enclosing scope installs depName.
func (r Resolver) Resolve(pkgKey, depName string) (string, bool) {
	prefix := pkgKey
	for {
		candidate := r.child(prefix, depName)
		if candidate != pkgKey && r.Keys[candidate] {
			return candidate, true
		}
		if prefix == "" {
			return "", false
		}
		prefix = r.parent(prefix)
	}
}

// child returns the key an install of name would have directly under prefix.
func (r Resolver) child(prefix, name string) string {
	if prefix == "" {
		return r.Segment + name
	}
	return prefix + "/" + r.Segment + name
}

// parent returns the key of the install directory enclosing key, or "" for a
// top-level key.
func (r Resolver) parent(key string) string {
	if r.Segment == "" {
		return stripLastInstallName(key)
	}
	i := strings.LastIndex(key, "/"+r.Segment)
	if i == -1 {
		return ""
	}
	return key[:i]
}

// stripLastInstallName removes the last install-name segment from a
// "/"-joined key. A scoped name ("@scope/name") is a single segment.
func stripLastInstallName(key string) string {
	i := strings.LastIndex(key, "/")
	if i == -1 {
		return ""
	}
	head := key[:i]
	j := strings.LastIndex(head, "/")
	if strings.HasPrefix(head[j+1:], "@") {
		if j == -1 {
			return ""
		}
		return head[:j]
	}
	return head
}
