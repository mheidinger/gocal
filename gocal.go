package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

const (
	modPath           = "go.mod"
	defaultConfigPath = ".gocal"
)

// Expects current dir to be the go module to be linted
// Use first commandline argument to pass in a different config file name
func main() {
	configPath := defaultConfigPath
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	modulePath, err := getModulePath()
	if err != nil {
		fmt.Printf("Failed to get module path from go mod file: %v\n", err)
		os.Exit(1)
	}

	layerModules, err := getLayerModules(configPath)
	if err != nil {
		fmt.Printf("Failed to get layer modules from config: %v\n", err)
		os.Exit(1)
	} else if len(layerModules) < 2 {
		fmt.Printf("Minimum two layers need to be defined")
		os.Exit(1)
	}

	layerModulePaths := make([]string, len(layerModules))
	for it, module := range layerModules {
		layerModulePaths[it] = modulePath + "/" + module
	}

	// Ignore last layer as it is allowed to import everything
	for it, module := range layerModules[:len(layerModules)-1] {
		// Use all following layers as forbidden imports
		checkModuleImports(module, layerModulePaths[it+1:])
	}
}

// Get path of the module we want to lint
func getModulePath() (path string, err error) {
	moduleBytes, err := ioutil.ReadFile(modPath)
	if err != nil {
		return
	}
	path = modfile.ModulePath(moduleBytes)
	return
}

// Read config file for the modules of the clean architecture layers
func getLayerModules(configPath string) (modules []string, err error) {
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return
	}

	content := strings.ReplaceAll(string(configBytes), "\r\n", "\n")
	modules = strings.Split(content, "\n")
	return
}

// Walk through all imports of the module and its submodules and check against forbidden module paths
func checkModuleImports(moduleName string, forbiddenModulePaths []string) {
	if strings.HasPrefix(moduleName, "/") {
		moduleName = "." + moduleName
	}

	filepath.Walk(moduleName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Failed to traverse to directory: %v\n", err)
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			fmt.Printf("Failed to parse file: %v\n", err)
			return nil
		}

		for _, imp := range file.Imports {
			cleanPath := strings.ReplaceAll(imp.Path.Value, "\"", "")
			for _, forbiddenPath := range forbiddenModulePaths {
				if strings.HasPrefix(cleanPath, forbiddenPath) {
					fmt.Printf("%s: Forbidden import of %s\n", path, forbiddenPath)
				}
			}
		}

		return nil
	})
	return
}
