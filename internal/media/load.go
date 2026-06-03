package media

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadRecipe parses and validates a media recipe from a YAML file.
func LoadRecipe(path string) (Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Recipe{}, fmt.Errorf("load recipe: %w", err)
	}
	return LoadRecipeFromBytes(data)
}

// LoadRecipeFromBytes parses and validates a media recipe from YAML bytes.
func LoadRecipeFromBytes(data []byte) (Recipe, error) {
	var r Recipe
	if err := yaml.Unmarshal(data, &r); err != nil {
		return Recipe{}, fmt.Errorf("load recipe: parse: %w", err)
	}
	if err := Validate(r); err != nil {
		return Recipe{}, err
	}
	return r, nil
}
