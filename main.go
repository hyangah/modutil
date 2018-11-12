package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	root = &cobra.Command{
		Use:   "modutil",
		Short: "various module utilities",
	}

	test = &cobra.Command{
		Use:   "test",
		Short: "run all tests of modules found in the subdirectories",
		Run:   runTest,
	}
)

var goExe = "go"

func init() {
	goroot := runtime.GOROOT()
	f, err := os.Open(filepath.Join(goroot, "bin", "go"))
	if err == nil {
		goExe = f.Name()
		f.Close()
	}
	root.AddCommand(test)
}

func main() {
	root.Execute()
}

func runTest(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := info.Name()
		if info.IsDir() && len(name) > 1 && name[0] == '.' {
			return filepath.SkipDir
		}

		dir, file := filepath.Split(path)
		if file != "go.mod" {
			return nil
		}

		module, err := readGoMod(path)
		if err != nil {
			return err
		}
		fmt.Println("*", module)

		if dir != "" {
			if err := os.Chdir(dir); err != nil {
				return err
			}
			defer os.Chdir(cwd)
		}
		cmd := exec.Command(goExe, "test", "./...")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	})
	if err != nil {
		panic(err)
	}
}

func readGoMod(gomod string) (string, error) {
	f, err := os.Open(gomod)
	if err != nil {
		return "", err
	}
	r := bufio.NewReader(f)
	for {
		s, err := r.ReadString('\n')
		if err == io.EOF {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "module ") {
			return s, nil
		}
	}
}
