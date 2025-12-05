package common

import "github.com/golobby/container/v3"

// InjectFunc is a function type that takes a container.Container and returns an error.
type InjectFunc func(c container.Container) error

// InjectAll applies a series of injection functions to a container.
// It iterates through the provided injection functions and applies each one to the container.
// If any injection function returns an error, the process stops and the error is returned.
//
// Parameters:
//   - c: The container.Container to which the injection functions will be applied.
//   - injectFuncs: A variadic parameter of InjectFunc functions to be applied to the container.
//
// Returns:
//   - error: nil if all injections are successful, otherwise the first encountered error.
func InjectAll(c container.Container, injectFuncs ...InjectFunc) error {
	for _, inject := range injectFuncs {
		if err := inject(c); err != nil {
			return err
		}
	}
	return nil
}
