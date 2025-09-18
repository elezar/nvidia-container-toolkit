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

package containerd

import (
	"path/filepath"

	"github.com/NVIDIA/nvidia-container-toolkit/pkg/config/engine"
)

type WithTopLevel struct {
	engine.Interface
	topLevelConfig *topLevelConfig
}

type topLevelConfig struct {
	filename string
	config   *Config
}

func (c *WithTopLevel) Save(path string) (int64, error) {
	n, err := c.Interface.Save(path)
	if err != nil {
		return 0, err
	}

	switch {
	case n > 0:
		c.topLevelConfig.ensureImports(path)
	case n == 0:
		c.topLevelConfig.simplify(path)
	}

	return c.topLevelConfig.config.Save(c.topLevelConfig.filename)
}

func (c *WithTopLevel) RemoveRuntime(name string) error {
	if err := c.topLevelConfig.RemoveRuntime(name); err != nil {
		return err
	}
	return c.Interface.RemoveRuntime(name)
}

func (c *topLevelConfig) RemoveRuntime(name string) error {
	return c.config.RemoveRuntime(name)
}

func (c *topLevelConfig) simplify(dropInFilename string) {
	c.removeImports(dropInFilename)
	c.removeVersion()
}

// removeImports removes the imports specified in the file if the only entry
// corresponds to the path for the drop-in-file and the only other field in the
// file is the version field.
func (c *topLevelConfig) removeImports(dropInFilename string) {
	if len(c.config.Keys()) != 2 {
		return
	}
	if c.config.Get("version") == nil || c.config.Get("imports") == nil {
		return
	}

	currentImports, _ := c.config.getStringArrayValue([]string{"imports"})
	if len(currentImports) != 1 {
		return
	}

	requiredImport := filepath.Dir(dropInFilename) + "/*.toml"
	if currentImports[0] != requiredImport {
		return
	}
	c.config.Delete("imports")
}

// removeVersion removes the version if it is the ONLY field in the file.
func (c *topLevelConfig) removeVersion() {
	if len(c.config.Keys()) > 1 {
		return
	}
	if c.config.Get("version") == nil {
		return
	}
	c.config.Delete("version")
}

func (c *topLevelConfig) ensureImports(dropInFilename string) {
	config := c.config.Tree

	currentImports, _ := c.config.getStringArrayValue([]string{"imports"})
	requiredImport := filepath.Dir(dropInFilename) + "/*.toml"
	for _, currentImport := range currentImports {
		// If the requiredImport is already present, then we need not update the config.
		if currentImport == requiredImport {
			return
		}
	}

	currentImports = append(currentImports, requiredImport)

	// If the config is empty we need to set the version too.
	if len(config.Keys()) == 0 {
		config.Set("version", c.config.Version)
	}
	config.Set("imports", currentImports)
}
