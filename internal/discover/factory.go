package discover

import (
	"github.com/NVIDIA/nvidia-container-toolkit/internal/logger"
	"github.com/NVIDIA/nvidia-container-toolkit/internal/lookup/root"
)

type Factory struct {
	logger logger.Interface
	driver *root.Driver
	// TODO: replace with driver
	root        string
	devRoot     string
	hookCreator HookCreator
}

type FactoryOption func(*Factory)

func NewFactory(opts ...FactoryOption) *Factory {
	f := &Factory{}
	for _, opt := range opts {
		opt(f)
	}
	if f.logger == nil {
		f.logger = &logger.NullLogger{}
	}
	if f.driver == nil {
		f.driver = root.New()
	}
	if f.root == "" {
		f.root = f.driver.Root
	}
	if f.devRoot == "" {
		f.devRoot = f.root
	}
	return f
}

func WithHookCreator(hookCreator HookCreator) FactoryOption {
	return func(f *Factory) {
		f.hookCreator = hookCreator
	}
}

func WithLogger(logger logger.Interface) FactoryOption {
	return func(f *Factory) {
		f.logger = logger
	}
}

func WithDriver(driver *root.Driver) FactoryOption {
	return func(f *Factory) {
		f.driver = driver
	}
}

func WithRoot(root string) FactoryOption {
	return func(f *Factory) {
		f.root = root
	}
}

func WithDevRoot(devRoot string) FactoryOption {
	return func(f *Factory) {
		f.devRoot = devRoot
	}
}
