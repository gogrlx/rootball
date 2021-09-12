package rootball

import (
	"errors"
	"fmt"
	"strings"
)

// Pull in all included RecipieFiles first
// List all recipies across the RecipieFiles
// Starting with the RecipieFile HEAD, build out tree for each Recipe Dependent Graph

var ProtoRecipe Recipe
var RecipeSet []*Recipe

func GenerateTrees(allRecipies []*Recipe) ([]*Recipe, []error) {
	// check for duplicates
	errorList := []error{}
	hasDups, dups := NoDuplicateIDs(allRecipies)
	if hasDups {
		for _, dup := range dups {
			errorList = append(errorList, errors.New(fmt.Sprintf("Recipe identifier is not unique: %s", dup)))
		}
		return []*Recipe{}, errorList
	}
	// check for undefined deps
	allDefined, mising := AllDependenciesDefined(allRecipies)
	if !allDefined {
		for _, dep := range mising {
			errorList = append(errorList, errors.New(fmt.Sprintf("Recipe identifier is required but not defined: %s", dep)))
		}
		return []*Recipe{}, errorList
	}
	// check for cycles
	hasCycle, cycle := HasCycle(allRecipies)
	if hasCycle {
		errorList = append(errorList, fmt.Errorf("%w: %s", ErrDependencyCycleFound, PrintCycle(cycle)))
		return []*Recipe{}, errorList
	}
	// generate and return the roots
	recipeMap := make(map[string]*Recipe)
	for _, recipe := range allRecipies {
		recipeMap[recipe.ID] = recipe
	}
	for _, recipe := range allRecipies {
		for _, dep := range recipe.Dependencies {
			recipe.dependencies = append(recipe.dependencies, recipeMap[dep])
			recipeMap[dep].IsRequisite = true
		}
	}
	return FindRoots(allRecipies), nil
}

// Step 1: render the YAMLs (recipefiles)
// Step 2: recursively gather all recipefiles, adding each to a map[string]bool. Cycles between recipefiles are allowed.
// Step 3: make a list of all states, with dependencies attached, described by *unique* string identifiers
// Step 4: detect non-unique string identifiers, return an error for this
// Step 5: Pass in a list of all possible states, each identifying their dependencies as string IDs
// Step 6: For each of the recipes in the list, check for a dependency cycle using DFS (depth first search)
// Step 7: Build a dependency tree for each of the recipies in the cooked protorecipe
// Step 8: Scan for out-of-tree reicpies that need to be included
// Step 9: Build a dependency tree for each of the out-of-tree dependencies

// Start from step 4
func dfs(allRecipes *map[string]*Recipe, current string, isVisited *map[string]bool, isValidated *map[string]bool) (bool, []string) {
	if (*isVisited)[current] {
		//TODO return the cycle
		return findCycle(allRecipes, current, "", []string{})
	}
	(*isVisited)[current] = true
	for _, id := range (*allRecipes)[current].Dependencies {
		hasCycle, cycle := dfs(allRecipes, id, isVisited, isValidated)
		if hasCycle {
			return true, cycle
		}
	}
	(*isValidated)[current] = true
	(*isVisited)[current] = false
	return false, []string{}
}
func findCycle(allRecipes *map[string]*Recipe, top string, current string, chain []string) (bool, []string) {
	if current == top {
		chain = append(chain, current)
		return true, chain
	}
	if current == "" {
		current = top
	}
	chain = append(chain, current)
	for _, w := range (*allRecipes)[current].Dependencies {
		if w == top {
			chain = append(chain, w)
			return true, chain
		}
		isCycle, chain := findCycle(allRecipes, top, w, chain)
		if isCycle {
			return true, chain
		}
	}
	return false, []string{}
}

func NoDuplicateIDs(allRecipes []*Recipe) (bool, []string) {
	duplicates := []string{}
	recipeMap := make(map[string]*Recipe)
	for _, recipe := range allRecipes {
		if _, ok := recipeMap[recipe.ID]; !ok {
			recipeMap[recipe.ID] = recipe
		} else {
			duplicates = append(duplicates, recipe.ID)
		}
	}
	return len(duplicates) == 0, duplicates

}

func AllDependenciesDefined(allRecipes []*Recipe) (bool, []string) {
	unresolved := []string{}
	recipeMap := make(map[string]*Recipe)
	for _, recipe := range allRecipes {
		recipeMap[recipe.ID] = recipe
	}
	for _, recipe := range allRecipes {
		for _, dep := range recipe.Dependencies {
			if _, ok := recipeMap[dep]; !ok {
				unresolved = append(unresolved, dep)
			}
		}
	}
	return len(unresolved) == 0, unresolved
}

func HasCycle(allRecipes []*Recipe) (bool, []string) {
	isValidated := make(map[string]bool)
	isVisited := make(map[string]bool)
	recipeMap := make(map[string]*Recipe)
	for _, i := range allRecipes {
		isVisited[i.ID] = false
		isValidated[i.ID] = false
		recipeMap[i.ID] = i
	}
	for _, i := range allRecipes {
		if isValidated[i.ID] {
			continue
		}
		hasCycle, cycle := dfs(&recipeMap, i.ID, &isVisited, &isValidated)
		if hasCycle {

			return true, cycle
		}
	}
	return false, []string{}
}

func PrintCycle(cycle []string) string {
	out := ""
	maxLength := 0
	for _, w := range cycle {
		if len(w) > maxLength {
			maxLength = len(w)
		}
	}
	for i := 0; i < len(cycle); i++ {
		switch i {
		case 0:
			out += fmt.Sprintf("> %s%s V\n", cycle[i], strings.Repeat(" ", maxLength-len(cycle[i])))
		case len(cycle) - 1:
			out += fmt.Sprintf("|| %s%s||\n", cycle[i], strings.Repeat(" ", maxLength-len(cycle[i])))
		default:
			out += fmt.Sprintf("^ %s%s <\n", cycle[i], strings.Repeat(" ", maxLength-len(cycle[i])))
		}
	}
	return out
}
func FindRoots(allRecipes []*Recipe) []*Recipe {
	roots := []*Recipe{}
	for _, recipe := range allRecipes {
		if !recipe.IsRequisite {
			roots = append(roots, recipe)
		}
	}
	return roots
}
func PrintTrees(roots []*Recipe) string {
	output := ""
	for _, recipe := range roots {
		output += printNode(recipe, 0, false) + "\n\n"
	}
	return output
}

func printNode(recipe *Recipe, depth int, isLast bool) string {
	nodeline := strings.Repeat("|\t", depth)
	if depth != 0 {
		if isLast {
			nodeline += "└── "
		} else {
			nodeline += "├── "
		}
	}
	nodeline += recipe.ID
	for i, dep := range recipe.dependencies {
		if i == len(recipe.dependencies)-1 {
			printNode(dep, depth+1, true)
		} else {
			printNode(dep, depth+1, false)
		}
	}
	return nodeline
}
