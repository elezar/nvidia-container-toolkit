package discover

import (
	"strings"
)

// NewCUDACompatHookDiscoverer creates a discoverer for a enable-cuda-compat hook.
// This hook is responsible for setting up CUDA compatibility in the container and depends on the host driver version.
func (f *Factory) NewCUDACompatHookDiscoverer(version string, cudaCompatContainerRoot string) Discover {
	var args []string
	if version != "" && !strings.Contains(version, "*") {
		args = append(args, "--host-driver-version="+version)
	}
	if cudaCompatContainerRoot != "" {
		args = append(args, "--cuda-compat-container-root="+cudaCompatContainerRoot)
	}

	return f.hookCreator.Create("enable-cuda-compat", args...)
}
