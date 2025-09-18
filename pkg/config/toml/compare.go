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
	"math"
	"reflect"
	"sort"

	"github.com/pelletier/go-toml"
)

// CompareOptions allows customization of TOML comparison behavior.
type CompareOptions struct {
	// IgnoreComments ignores all comments in the comparison (default: true)
	IgnoreComments bool

	// IgnoreKeyOrder ignores the order of keys at each level (default: true)
	IgnoreKeyOrder bool

	// IgnoreEmptyTables ignores empty table declarations (default: false)
	IgnoreEmptyTables bool

	// FloatTolerance is the epsilon for floating-point comparisons (default: 1e-10)
	FloatTolerance float64

	// CustomComparators allows registering custom comparison functions for specific paths
	CustomComparators map[string]ComparatorFunc
}

// ComparatorFunc is a custom comparison function for specific TOML paths.
type ComparatorFunc func(expected, actual interface{}) bool

// DefaultCompareOptions returns the default comparison options.
func DefaultCompareOptions() CompareOptions {
	return CompareOptions{
		IgnoreComments:    true,
		IgnoreKeyOrder:    true,
		IgnoreEmptyTables: false,
		FloatTolerance:    1e-10,
		CustomComparators: make(map[string]ComparatorFunc),
	}
}

// CompareContents compares two TOML documents for semantic equality.
// It parses both TOML strings and compares their parsed structures,
// ignoring formatting differences, comments (by default), and key ordering.
func CompareContents(expected, actual string) (bool, error) {
	return CompareContentsWithOptions(expected, actual, DefaultCompareOptions())
}

// CompareContentsWithOptions compares two TOML documents with custom options.
func CompareContentsWithOptions(expected, actual string, opts CompareOptions) (bool, error) {
	// Parse the expected TOML
	expectedTree, err := toml.Load(expected)
	if err != nil {
		return false, fmt.Errorf("failed to parse expected TOML: %w", err)
	}

	// Parse the actual TOML
	actualTree, err := toml.Load(actual)
	if err != nil {
		return false, fmt.Errorf("failed to parse actual TOML: %w", err)
	}

	// Compare the trees
	return compareTrees(expectedTree, actualTree, "", opts), nil
}

// compareTrees recursively compares two TOML trees.
func compareTrees(expected, actual *toml.Tree, path string, opts CompareOptions) bool {
	// Check if there's a custom comparator for this path
	if comparator, exists := opts.CustomComparators[path]; exists {
		return comparator(expected, actual)
	}

	// Get all keys from both trees
	expectedKeys := expected.Keys()
	actualKeys := actual.Keys()

	// If we're not ignoring empty tables, check key count
	if !opts.IgnoreEmptyTables && len(expectedKeys) != len(actualKeys) {
		return false
	}

	// Create maps for easier lookup
	expectedMap := make(map[string]bool)
	for _, key := range expectedKeys {
		expectedMap[key] = true
	}

	actualMap := make(map[string]bool)
	for _, key := range actualKeys {
		actualMap[key] = true
	}

	// Check that all expected keys exist in actual
	for key := range expectedMap {
		if !actualMap[key] {
			if !opts.IgnoreEmptyTables || !isEmptyValue(expected.Get(key)) {
				return false
			}
		}
	}

	// Check that all actual keys exist in expected
	for key := range actualMap {
		if !expectedMap[key] {
			if !opts.IgnoreEmptyTables || !isEmptyValue(actual.Get(key)) {
				return false
			}
		}
	}

	// Compare values for each key
	for _, key := range expectedKeys {
		if !actualMap[key] && opts.IgnoreEmptyTables && isEmptyValue(expected.Get(key)) {
			continue
		}

		currentPath := path
		if currentPath == "" {
			currentPath = key
		} else {
			currentPath = path + "." + key
		}

		if !compareValues(expected.Get(key), actual.Get(key), currentPath, opts) {
			return false
		}
	}

	return true
}

// compareValues compares two values from TOML trees.
func compareValues(expected, actual interface{}, path string, opts CompareOptions) bool {
	// Check if there's a custom comparator for this path
	if comparator, exists := opts.CustomComparators[path]; exists {
		return comparator(expected, actual)
	}

	// Handle nil cases
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil || actual == nil {
		return false
	}

	// Handle TOML tree comparison
	if expectedTree, ok := expected.(*toml.Tree); ok {
		if actualTree, ok := actual.(*toml.Tree); ok {
			return compareTrees(expectedTree, actualTree, path, opts)
		}
		return false
	}

	// Handle slice comparison
	expectedVal := reflect.ValueOf(expected)
	actualVal := reflect.ValueOf(actual)

	if expectedVal.Kind() == reflect.Slice && actualVal.Kind() == reflect.Slice {
		return compareSlices(expectedVal, actualVal, path, opts)
	}

	// Handle float comparison with tolerance
	if isFloat(expected) && isFloat(actual) {
		return compareFloats(toFloat64(expected), toFloat64(actual), opts.FloatTolerance)
	}

	// Handle numeric type conversions
	if isNumeric(expected) && isNumeric(actual) {
		return compareNumeric(expected, actual)
	}

	// Default to deep equal
	return reflect.DeepEqual(expected, actual)
}

// compareSlices compares two slices element by element.
func compareSlices(expected, actual reflect.Value, path string, opts CompareOptions) bool {
	if expected.Len() != actual.Len() {
		return false
	}

	for i := 0; i < expected.Len(); i++ {
		elementPath := fmt.Sprintf("%s[%d]", path, i)
		if !compareValues(expected.Index(i).Interface(), actual.Index(i).Interface(), elementPath, opts) {
			return false
		}
	}

	return true
}

// isEmptyValue checks if a value represents an empty table or nil.
func isEmptyValue(v interface{}) bool {
	if v == nil {
		return true
	}

	if tree, ok := v.(*toml.Tree); ok {
		return len(tree.Keys()) == 0
	}

	// Check for empty slices
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Slice {
		return val.Len() == 0
	}

	return false
}

// isFloat checks if a value is a floating-point number.
func isFloat(v interface{}) bool {
	switch v.(type) {
	case float32, float64:
		return true
	default:
		return false
	}
}

// isNumeric checks if a value is any numeric type.
func isNumeric(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}

// toFloat64 converts a numeric value to float64.
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	default:
		return 0
	}
}

// compareFloats compares two floating-point numbers with a tolerance.
func compareFloats(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

// compareNumeric compares two numeric values, handling type conversions.
func compareNumeric(expected, actual interface{}) bool {
	// Convert both to float64 for comparison
	expectedFloat := toFloat64(expected)
	actualFloat := toFloat64(actual)

	// If both values are integers, use exact comparison
	if isInteger(expected) && isInteger(actual) {
		return expectedFloat == actualFloat
	}

	// Otherwise use float comparison with small tolerance
	return compareFloats(expectedFloat, actualFloat, 1e-10)
}

// isInteger checks if a value is an integer type.
func isInteger(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}

// GetOrderedKeys returns the keys of a TOML tree in sorted order.
// This is useful for consistent iteration when key order matters.
func GetOrderedKeys(tree *toml.Tree) []string {
	keys := tree.Keys()
	sort.Strings(keys)
	return keys
}
