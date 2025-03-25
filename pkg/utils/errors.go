package utils

import "fmt"

func TraceError(context string) func(error) error {
	return (func(err error) error {
		return fmt.Errorf("%s: %w", context, err)
	})
}
