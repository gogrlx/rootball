package rootball

import (
	"fmt"
	"strings"
)

// Pull in all included RecipieFiles first
// List all recipies across the RecipieFiles
// Starting with the RecipieFile HEAD, build out tree for each Recipe Dependent Graph

var ProtoRecipe Recipe
var RecipeSet []*Recipe

func GenerateTree(recipe *Recipe, isRoot bool, head string) error {
	for _, i := range recipe.dependencies {
		if i.ID == head && !isRoot {
			return ErrDependencyCycleFound
		}
	}
	return nil
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
func dfs(allRecipes *map[string]*Recipe, current string, isVisited *map[string]bool) (bool, []string) {
	if (*isVisited)[current] {
		//TODO return the cycle
		return findCycle(allRecipes, current, "", []string{})
	}
	(*isVisited)[current] = true
	for _, id := range (*allRecipes)[current].Dependencies {
		hasCycle, cycle := dfs(allRecipes, id, isVisited)
		if hasCycle {
			return true, cycle
		}
	}
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

func HasCycle(allRecipes []*Recipe) (bool, []string, error) {
	isVisited := make(map[string]bool)
	recipeMap := make(map[string]*Recipe)
	for _, i := range allRecipes {
		isVisited[i.ID] = false
		recipeMap[i.ID] = i
	}
	for _, i := range allRecipes {
		hasCycle, cycle := dfs(&recipeMap, i.ID, &isVisited)
		if hasCycle {

			return true, cycle, ErrDependencyCycleFound
		}
	}
	return false, []string{}, nil
}

func PrintCycle(cycle []string) {
	maxLength := 0
	for _, w := range cycle {
		if len(w) > maxLength {
			maxLength = len(w)
		}
	}
	for i := 0; i < len(cycle); i++ {
		switch i {
		case 0:
			fmt.Printf("> %s%s V\n", cycle[i], strings.Repeat(" ", maxLength-len(cycle[i])))
		case len(cycle) - 1:
			fmt.Printf("|| %s%s||\n", cycle[i], strings.Repeat(" ", maxLength-len(cycle[i])))
		default:
			fmt.Printf("^ %s%s <\n", cycle[i], strings.Repeat(" ", maxLength-len(cycle[i])))
		}
	}

}

func PrintTrees(roots []*Recipe) {
	for _, recipe := range roots {
		printNode(recipe, 0, false)
	}

}

func printNode(recipe *Recipe, depth int, isLast bool) {
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
		if i == len(recipe.dependencies) {
			printNode(dep, depth+1, true)
		} else {
			printNode(dep, depth+1, false)
		}
	}

}

func BuildTrees(allRecipies []*Recipe) ([]Recipe, error) {

	return []Recipe{}, nil
}
