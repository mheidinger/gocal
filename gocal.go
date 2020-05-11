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

	for it, module := range layerModules[:len(layerModules)-1] {
		importMap := getModuleImports(module)

		checkImports(importMap, layerModules[it+1:], modulePath)
	}
}

func getModulePath() (path string, err error) {
	moduleBytes, err := ioutil.ReadFile(modPath)
	if err != nil {
		return
	}
	path = modfile.ModulePath(moduleBytes)
	return
}

func getLayerModules(configPath string) (modules []string, err error) {
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return
	}

	content := strings.ReplaceAll(string(configBytes), "\r\n", "\n")
	modules = strings.Split(content, "\n")
	return
}

func getModuleImports(moduleName string) (imports map[string][]string) {
	if strings.HasPrefix(moduleName, "/") {
		moduleName = "." + moduleName
	}

	imports = make(map[string][]string)
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
			if files, ok := imports[cleanPath]; ok {
				imports[cleanPath] = append(files, path)
			} else {
				imports[cleanPath] = []string{path}
			}
		}

		return nil
	})
	return
}

func checkImports(importMap map[string][]string, layerModules []string, modulePath string) {
	fullLayerModule := make([]string, len(layerModules))
	for it, module := range layerModules {
		fullLayerModule[it] = modulePath + "/" + module
	}

	for imp, files := range importMap {
		for _, mod := range fullLayerModule {
			if strings.HasPrefix(imp, mod) {
				fmt.Printf("Forbidden import '%s' in files '%v'\n", imp, files)
			}
		}
	}
}
