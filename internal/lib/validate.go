package lib

import (
	"fmt"
)

func Validate(filepath string) error {
	exists, err := pathExists(filepath)
	if err != nil {
		return fmt.Errorf("error checking existance of file: %w", err)
	}

	if !exists {
		return fmt.Errorf("path %v does not exist", filepath)
	}

	_, err = NewGodotConfig(filepath)
	if err != nil {
		return err
	}
	return nil
}
