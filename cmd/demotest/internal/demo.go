/*
* Copyright (c) 2024 The GoPlus Authors (goplus.org). All rights reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */
package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func RunGenPkgDemo(demoRoot string) {
	// we have to feed the pc to llcppg
	os.Setenv("PKG_CONFIG_PATH", demoRoot)

	demosPath := filepath.Join(demoRoot, "_demo")

	fmt.Printf("testing demos in %s\n", demosPath)
	// run the demo
	var demos []os.DirEntry
	demos, err := os.ReadDir(demosPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read demo directory: %v", err))
	}
	for _, demo := range demos {
		if demo.IsDir() {
			fmt.Printf("Running demo: %s\n", demo.Name())
			if demoErr := runCommand(filepath.Join(demosPath, demo.Name()), "llgo", "run", "."); demoErr != nil {
				panic(fmt.Sprintf("failed to run demo: %s: %v", demo.Name(), demoErr))
			}
		}
	}
}

func RunAllGenPkgDemos(baseDir string) {
	fmt.Printf("Starting generated package tests in directory: %s\n", baseDir)

	RunGenPkgDemo(baseDir)
}

func runCommand(dir, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
