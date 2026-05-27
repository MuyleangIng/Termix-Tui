package project

import (
	"os"
	"path/filepath"
)

type Stack string

const (
	React      Stack = "React"
	NextJS     Stack = "Next.js"
	Node       Stack = "Node.js"
	Go         Stack = "Go"
	Rust       Stack = "Rust"
	Python     Stack = "Python"
	Kubernetes Stack = "Kubernetes"
	Java       Stack = "Java"
	DotNet     Stack = ".NET"
	Laravel    Stack = "Laravel"
	Vite       Stack = "Vite"
	Astro      Stack = "Astro"
)

func Detect(dir string) []Stack {
	tests := []struct {
		stack Stack
		files []string
	}{
		{Node, []string{"package.json"}},
		{NextJS, []string{"next.config.js", "next.config.mjs", "next.config.ts"}},
		{Go, []string{"go.mod"}},
		{Rust, []string{"Cargo.toml"}},
		{Python, []string{"pyproject.toml", "requirements.txt"}},
		{Kubernetes, []string{"k8s", "kubernetes", "helm"}},
		{Java, []string{"pom.xml", "build.gradle"}},
		{DotNet, []string{"*.csproj", "*.sln"}},
		{Laravel, []string{"artisan"}},
		{Vite, []string{"vite.config.js", "vite.config.ts"}},
		{Astro, []string{"astro.config.mjs", "astro.config.ts"}},
		{React, []string{"src/App.jsx", "src/App.tsx"}},
	}
	var stacks []Stack
	for _, test := range tests {
		if anyExists(dir, test.files) {
			stacks = append(stacks, test.stack)
		}
	}
	return stacks
}

func anyExists(dir string, files []string) bool {
	for _, file := range files {
		matches, _ := filepath.Glob(filepath.Join(dir, file))
		if len(matches) > 0 {
			return true
		}
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			return true
		}
	}
	return false
}
