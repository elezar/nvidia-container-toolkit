/**
# Copyright (c) 2022, NVIDIA CORPORATION.  All rights reserved.
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

package discover

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/NVIDIA/nvidia-container-toolkit/internal/lookup"
	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	"github.com/sirupsen/logrus"
)

// NewLDCacheUpdateHook creates a discoverer that updates the ldcache for the specified mounts. A logger can also be specified
func NewLDCacheUpdateHook(logger *logrus.Logger, mounts Discover, cfg *Config) (Discover, error) {
	d := ldconfig{
		logger:                  logger,
		mountsFrom:              mounts,
		lookup:                  lookup.NewExecutableLocator(logger, cfg.Root),
		nvidiaCTKExecutablePath: cfg.NVIDIAContainerToolkitCLIExecutablePath,
	}

	return &d, nil
}

const (
	nvidiaCTKDefaultFilePath = "/usr/bin/nvidia-ctk"
)

type ldconfig struct {
	None
	logger                  *logrus.Logger
	mountsFrom              Discover
	lookup                  lookup.Locator
	nvidiaCTKExecutablePath string
}

// Hooks checks the required mounts for libraries and returns a hook to update the LDcache for the discovered paths.
func (d ldconfig) Hooks() ([]Hook, error) {
	mounts, err := d.mountsFrom.Mounts()
	if err != nil {
		return nil, fmt.Errorf("failed to discover mounts for ldcache update: %v", err)
	}
	h := CreateLDCacheUpdateHook(
		d.logger,
		d.lookup,
		d.nvidiaCTKExecutablePath,
		nvidiaCTKDefaultFilePath,
		getLibraryPaths(mounts),
	)
	return []Hook{h}, nil
}

// CreateLDCacheUpdateHook locates the NVIDIA Container Toolkit CLI and creates a hook for updating the LD Cache
func CreateLDCacheUpdateHook(logger *logrus.Logger, lookup lookup.Locator, execuable string, defaultPath string, libraries []string) Hook {
	hookPath := defaultPath
	targets, err := lookup.Locate(execuable)
	if err != nil {
		logger.Warnf("Failed to locate %v: %v", execuable, err)
	} else if len(targets) == 0 {
		logger.Warnf("%v not found", execuable)
	} else {
		logger.Debugf("Found %v candidates: %v", execuable, targets)
		hookPath = targets[0]
	}
	logger.Debugf("Using NVIDIA Container Toolkit CLI path %v", hookPath)

	args := []string{filepath.Base(hookPath), "hook", "update-ldcache"}
	for _, f := range uniqueFolders(libraries) {
		args = append(args, "--folder", f)
	}
	return Hook{
		Lifecycle: cdi.CreateContainerHook,
		Path:      hookPath,
		Args:      args,
	}
}

// getLibraryPaths extracts the library dirs from the specified mounts
func getLibraryPaths(mounts []Mount) []string {
	var paths []string
	for _, m := range mounts {
		if !isLibName(m.Path) {
			continue
		}
		paths = append(paths, m.Path)
	}
	return paths
}

// isLibName checks if the specified filename is a library (i.e. ends in `.so*`)
func isLibName(filename string) bool {

	base := filepath.Base(filename)

	isLib, err := filepath.Match("lib?*.so*", base)
	if !isLib || err != nil {
		return false
	}

	parts := strings.Split(base, ".so")
	if len(parts) == 1 {
		return true
	}

	return parts[len(parts)-1] == "" || strings.HasPrefix(parts[len(parts)-1], ".")
}

// uniqueFolders returns the unique set of folders for the specified files
func uniqueFolders(libraries []string) []string {
	var paths []string
	checked := make(map[string]bool)

	for _, l := range libraries {
		dir := filepath.Dir(l)
		if dir == "" {
			continue
		}
		if checked[dir] {
			continue
		}
		checked[dir] = true
		paths = append(paths, dir)
	}
	return paths
}
