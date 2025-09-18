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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareContents(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
		actual   string
		equal    bool
		wantErr  bool
	}{
		{
			name:     "identical strings",
			expected: `key = "value"`,
			actual:   `key = "value"`,
			equal:    true,
		},
		{
			name: "different whitespace",
			expected: `key = "value"
table = { nested = "data" }`,
			actual: `key="value"
table={nested="data"}`,
			equal: true,
		},
		{
			name: "different key order",
			expected: `a = 1
b = 2
c = 3`,
			actual: `c = 3
a = 1
b = 2`,
			equal: true,
		},
		{
			name: "nested tables with different order",
			expected: `[table1]
key1 = "value1"
key2 = "value2"

[table2]
key3 = "value3"`,
			actual: `[table2]
key3 = "value3"

[table1]
key2 = "value2"
key1 = "value1"`,
			equal: true,
		},
		{
			name:     "inline vs expanded tables",
			expected: `table = { key1 = "value1", key2 = "value2" }`,
			actual: `[table]
key1 = "value1"
key2 = "value2"`,
			equal: true,
		},
		{
			name: "comments are ignored by default",
			expected: `# This is a comment
key = "value" # inline comment`,
			actual: `key = "value"`,
			equal:  true,
		},
		{
			name:     "numeric type differences - integers",
			expected: `int_val = 42`,
			actual:   `int_val = 42`,
			equal:    true,
		},
		{
			name:     "numeric type differences - floats",
			expected: `float_val = 1.0`,
			actual:   `float_val = 1`,
			equal:    true,
		},
		{
			name:     "array comparison - same order",
			expected: `arr = [1, 2, 3]`,
			actual:   `arr = [1, 2, 3]`,
			equal:    true,
		},
		{
			name:     "array comparison - different order",
			expected: `arr = [1, 2, 3]`,
			actual:   `arr = [3, 2, 1]`,
			equal:    false,
		},
		{
			name:     "empty tables",
			expected: `[empty_table]`,
			actual:   `[empty_table]`,
			equal:    true,
		},
		{
			name: "missing key",
			expected: `key1 = "value1"
key2 = "value2"`,
			actual: `key1 = "value1"`,
			equal:  false,
		},
		{
			name:     "extra key",
			expected: `key1 = "value1"`,
			actual: `key1 = "value1"
key2 = "value2"`,
			equal: false,
		},
		{
			name:     "different values",
			expected: `key = "value1"`,
			actual:   `key = "value2"`,
			equal:    false,
		},
		{
			name: "complex nested structure",
			expected: `[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
    enable_cdi = true
    [plugins."io.containerd.grpc.v1.cri".containerd]
      default_runtime_name = "runc"
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.nvidia]
          runtime_type = "io.containerd.runc.v2"
          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.nvidia.options]
            BinaryName = "/usr/bin/nvidia-container-runtime"`,
			actual: `[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
    [plugins."io.containerd.grpc.v1.cri".containerd]
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.nvidia]
          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.nvidia.options]
            BinaryName = "/usr/bin/nvidia-container-runtime"
          runtime_type = "io.containerd.runc.v2"
      default_runtime_name = "runc"
    enable_cdi = true`,
			equal: true,
		},
		{
			name: "invalid expected TOML",
			expected: `key = "value
missing quote`,
			actual:  `key = "value"`,
			wantErr: true,
		},
		{
			name:     "invalid actual TOML",
			expected: `key = "value"`,
			actual: `key = "value
missing quote`,
			wantErr: true,
		},
		{
			name: "multiline strings",
			expected: `text = """
Line 1
Line 2
Line 3"""`,
			actual: `text = """
Line 1
Line 2
Line 3"""`,
			equal: true,
		},
		{
			name: "boolean values",
			expected: `enabled = true
disabled = false`,
			actual: `disabled = false
enabled = true`,
			equal: true,
		},
		{
			name: "array of tables",
			expected: `[[servers]]
name = "alpha"
ip = "10.0.0.1"

[[servers]]
name = "beta"
ip = "10.0.0.2"`,
			actual: `[[servers]]
name = "alpha"
ip = "10.0.0.1"

[[servers]]
name = "beta"
ip = "10.0.0.2"`,
			equal: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			equal, err := CompareContents(tc.expected, tc.actual)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.equal, equal)
			}
		})
	}
}

func TestCompareContentsWithOptions(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
		actual   string
		opts     CompareOptions
		equal    bool
	}{
		{
			name: "ignore empty tables enabled",
			expected: `[table]
key = "value"`,
			actual: `[table]
key = "value"
[empty_table]`,
			opts: CompareOptions{
				IgnoreComments:    true,
				IgnoreKeyOrder:    true,
				IgnoreEmptyTables: true,
			},
			equal: true,
		},
		{
			name: "ignore empty tables disabled",
			expected: `[table]
key = "value"`,
			actual: `[table]
key = "value"
[empty_table]`,
			opts: CompareOptions{
				IgnoreComments:    true,
				IgnoreKeyOrder:    true,
				IgnoreEmptyTables: false,
			},
			equal: false,
		},
		{
			name:     "custom comparator",
			expected: `version = "1.0.0"`,
			actual:   `version = "1.0.1"`,
			opts: CompareOptions{
				IgnoreComments: true,
				IgnoreKeyOrder: true,
				CustomComparators: map[string]ComparatorFunc{
					"version": func(expected, actual interface{}) bool {
						// Accept any version starting with "1.0"
						e, eOk := expected.(string)
						a, aOk := actual.(string)
						if !eOk || !aOk {
							return false
						}
						return len(e) >= 3 && len(a) >= 3 && e[:3] == a[:3]
					},
				},
			},
			equal: true,
		},
		{
			name:     "float tolerance",
			expected: `pi = 3.14159265358979`,
			actual:   `pi = 3.14159265358978`,
			opts: CompareOptions{
				IgnoreComments: true,
				IgnoreKeyOrder: true,
				FloatTolerance: 1e-10,
			},
			equal: true,
		},
		{
			name:     "float tolerance exceeded",
			expected: `pi = 3.14159`,
			actual:   `pi = 3.14160`,
			opts: CompareOptions{
				IgnoreComments: true,
				IgnoreKeyOrder: true,
				FloatTolerance: 1e-10,
			},
			equal: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			equal, err := CompareContentsWithOptions(tc.expected, tc.actual, tc.opts)
			require.NoError(t, err)
			assert.Equal(t, tc.equal, equal)
		})
	}
}

func TestTestHelpers(t *testing.T) {
	t.Run("RequireEqualTOML success", func(t *testing.T) {
		expected := `key = "value"`
		actual := `key = "value"`
		RequireEqualTOML(t, expected, actual)
	})

	t.Run("AssertEqualTOML success", func(t *testing.T) {
		expected := `key = "value"`
		actual := `key = "value"`
		result := AssertEqualTOML(t, expected, actual)
		assert.True(t, result)
	})

	t.Run("AssertEqualTOML failure", func(t *testing.T) {
		mockT := &testing.T{}
		expected := `key = "value1"`
		actual := `key = "value2"`
		result := AssertEqualTOML(mockT, expected, actual)
		assert.False(t, result)
	})
}

func TestNumericComparison(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
		actual   string
		equal    bool
	}{
		{
			name:     "int to float comparison",
			expected: `num = 42`,
			actual:   `num = 42.0`,
			equal:    true,
		},
		{
			name:     "float to int comparison",
			expected: `num = 1.0`,
			actual:   `num = 1`,
			equal:    true,
		},
		{
			name:     "different integers",
			expected: `num = 42`,
			actual:   `num = 43`,
			equal:    false,
		},
		{
			name:     "hex and decimal",
			expected: `num = 255`,
			actual:   `num = 0xFF`,
			equal:    true,
		},
		{
			name:     "octal and decimal",
			expected: `num = 493`,
			actual:   `num = 0o755`,
			equal:    true,
		},
		{
			name:     "binary and decimal",
			expected: `num = 10`,
			actual:   `num = 0b1010`,
			equal:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			equal, err := CompareContents(tc.expected, tc.actual)
			require.NoError(t, err)
			assert.Equal(t, tc.equal, equal)
		})
	}
}

func TestEmptyValues(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
		actual   string
		opts     CompareOptions
		equal    bool
	}{
		{
			name:     "empty arrays",
			expected: `arr = []`,
			actual:   `arr = []`,
			equal:    true,
		},
		{
			name:     "empty table vs missing with IgnoreEmptyTables",
			expected: `[table]`,
			actual:   ``,
			opts: CompareOptions{
				IgnoreComments:    true,
				IgnoreKeyOrder:    true,
				IgnoreEmptyTables: true,
			},
			equal: true,
		},
		{
			name:     "empty table vs missing without IgnoreEmptyTables",
			expected: `[table]`,
			actual:   ``,
			opts: CompareOptions{
				IgnoreComments:    true,
				IgnoreKeyOrder:    true,
				IgnoreEmptyTables: false,
			},
			equal: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := tc.opts
			if opts.CustomComparators == nil {
				opts = DefaultCompareOptions()
				opts.IgnoreEmptyTables = tc.opts.IgnoreEmptyTables
			}
			equal, err := CompareContentsWithOptions(tc.expected, tc.actual, opts)
			require.NoError(t, err)
			assert.Equal(t, tc.equal, equal)
		})
	}
}
