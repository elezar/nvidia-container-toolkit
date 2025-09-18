/**
# Copyright 2025 NVIDIA CORPORATION
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package toml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RequireEqualTOML asserts that two TOML strings are semantically equal.
// It parses both strings as TOML and compares their structure, ignoring
// formatting differences, comments, and key ordering.
func RequireEqualTOML(t *testing.T, expected, actual string, msgAndArgs ...interface{}) {
	t.Helper()
	equal, err := CompareContents(expected, actual)
	if err != nil {
		require.FailNow(t, fmt.Sprintf("Failed to compare TOML: %v", err), msgAndArgs...)
	}
	if !equal {
		msg := fmt.Sprintf("TOML documents are not semantically equal:\n\nExpected:\n%s\n\nActual:\n%s", expected, actual)
		require.FailNow(t, msg, msgAndArgs...)
	}
}

// RequireEqualTOMLWithOptions asserts that two TOML strings are semantically
// equal according to the provided options.
func RequireEqualTOMLWithOptions(t *testing.T, expected, actual string, opts CompareOptions, msgAndArgs ...interface{}) {
	t.Helper()
	equal, err := CompareContentsWithOptions(expected, actual, opts)
	if err != nil {
		require.FailNow(t, fmt.Sprintf("Failed to compare TOML: %v", err), msgAndArgs...)
	}
	if !equal {
		msg := fmt.Sprintf("TOML documents are not semantically equal:\n\nExpected:\n%s\n\nActual:\n%s", expected, actual)
		require.FailNow(t, msg, msgAndArgs...)
	}
}

// AssertEqualTOML checks that two TOML strings are semantically equal.
// It parses both strings as TOML and compares their structure, ignoring
// formatting differences, comments, and key ordering.
func AssertEqualTOML(t *testing.T, expected, actual string, msgAndArgs ...interface{}) bool {
	t.Helper()
	equal, err := CompareContents(expected, actual)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("Failed to compare TOML: %v", err), msgAndArgs...)
	}
	if !equal {
		msg := fmt.Sprintf("TOML documents are not semantically equal:\n\nExpected:\n%s\n\nActual:\n%s", expected, actual)
		return assert.Fail(t, msg, msgAndArgs...)
	}
	return true
}

// AssertEqualTOMLWithOptions checks that two TOML strings are semantically
// equal according to the provided options.
func AssertEqualTOMLWithOptions(t *testing.T, expected, actual string, opts CompareOptions, msgAndArgs ...interface{}) bool {
	t.Helper()
	equal, err := CompareContentsWithOptions(expected, actual, opts)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("Failed to compare TOML: %v", err), msgAndArgs...)
	}
	if !equal {
		msg := fmt.Sprintf("TOML documents are not semantically equal:\n\nExpected:\n%s\n\nActual:\n%s", expected, actual)
		return assert.Fail(t, msg, msgAndArgs...)
	}
	return true
}

// RequireEqualTOMLFiles asserts that two TOML files are semantically equal.
func RequireEqualTOMLFiles(t *testing.T, expectedPath, actualPath string, msgAndArgs ...interface{}) {
	t.Helper()
	expectedTree, err := LoadFile(expectedPath)
	require.NoError(t, err, "Failed to load expected TOML file")

	actualTree, err := LoadFile(actualPath)
	require.NoError(t, err, "Failed to load actual TOML file")

	RequireEqualTOML(t, expectedTree.String(), actualTree.String(), msgAndArgs...)
}

// AssertEqualTOMLFiles checks that two TOML files are semantically equal.
func AssertEqualTOMLFiles(t *testing.T, expectedPath, actualPath string, msgAndArgs ...interface{}) bool {
	t.Helper()
	expectedTree, err := LoadFile(expectedPath)
	if !assert.NoError(t, err, "Failed to load expected TOML file") {
		return false
	}

	actualTree, err := LoadFile(actualPath)
	if !assert.NoError(t, err, "Failed to load actual TOML file") {
		return false
	}

	return AssertEqualTOML(t, expectedTree.String(), actualTree.String(), msgAndArgs...)
}
