package rootball

import "errors"

type RecipeFile struct {
	Recipes    []*Recipe
	Includes   []string
	includes   []*RecipeFile
	IsIncluded bool
	ID         string
}

type Recipe struct {
	Dependencies []string
	dependencies []*Recipe
	dependents   []*Recipe
	IsRequisite  bool
	ID           string
}

var ErrDependencyCycleFound error

func init() {
	ErrDependencyCycleFound = errors.New("Found a dependency cycle!")
}
