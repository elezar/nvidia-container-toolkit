/*
# Copyright (c) 2021, NVIDIA CORPORATION.  All rights reserved.
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
*/

package discover

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/NVIDIA/nvidia-container-toolkit/internal/logger"
	"github.com/NVIDIA/nvidia-container-toolkit/pkg/lookup"
)

// mounts is a generic discoverer for Mounts. It is customized by specifying the
// required entities as a list and a Locator that is used to find the target mounts
// based on the entry in the list.
type mounts struct {
	None
	logger   logger.Interface
	lookup   lookup.Locator
	root     string
	required []string
}

var _ Discover = (*mounts)(nil)

// NewMounts creates a discoverer for the required mounts using the specified locator.
func (f *Factory) NewMounts(lookup lookup.Locator, required []string) Discover {
	return WithCache(f.newMounts(lookup, f.driver.Root, required))
}

// newMounts creates a discoverer for the required mounts using the specified locator.
func (f *Factory) newMounts(lookup lookup.Locator, root string, required []string) *mounts {
	return &mounts{
		logger:   f.logger,
		lookup:   lookup,
		root:     filepath.Join("/", root),
		required: required,
	}
}

// locateAll returns all resolved patterns.
// These are treated as mounts.
func (d *mounts) locateAll(patterns ...string) []string {
	var allCandidates []string
	for _, pattern := range patterns {
		candidates, err := d.lookup.Locate(pattern)
		if err != nil {
			d.logger.Warningf("Could not locate %v: %v", pattern, err)
			continue
		}
		if len(candidates) == 0 {
			d.logger.Warningf("Missing %v", pattern)
			continue
		}
		d.logger.Debugf("Located %v as %v", pattern, candidates)
		allCandidates = append(allCandidates, candidates...)
	}
	return allCandidates
}

func (d *mounts) Mounts() ([]Mount, error) {
	if d.lookup == nil {
		return nil, fmt.Errorf("no lookup defined")
	}

	var mounts []Mount
	seen := make(map[string]bool)
	for _, candidate := range d.locateAll(d.required...) {
		if seen[candidate] {
			d.logger.Debugf("Skipping duplicate mount %v", candidate)
			continue
		}
		seen[candidate] = true
		mounts = append(mounts, d.newMount(candidate))
	}

	return mounts, nil
}

func (d *mounts) newMount(path string) Mount {
	containerPath := d.relativeTo(path)
	if containerPath == "" {
		containerPath = path
	}

	d.logger.Infof("Selecting %v as %v", path, containerPath)
	mount := Mount{
		HostPath: path,
		Path:     containerPath,
		Options: []string{
			"ro",
			"nosuid",
			"nodev",
			"rbind",
			"rprivate",
		},
	}
	return mount
}

// relativeTo returns the path relative to the root for the file locator
func (d *mounts) relativeTo(path string) string {
	if d.root == "/" {
		return path
	}

	return strings.TrimPrefix(path, d.root)
}
