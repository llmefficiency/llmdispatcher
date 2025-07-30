package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("🧪 LLM Dispatcher Test Suite")
	fmt.Println("=============================")
	fmt.Println("")

	// Load environment variables from .env file
	loadEnvFile()

	// Run tests
	fmt.Println("Running tests with coverage...")

	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "-covermode=atomic", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ All tests passed!")

	// Generate coverage report
	fmt.Println("\nGenerating coverage report...")

	cmd = exec.Command("go", "tool", "cover", "-func=coverage.out")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Warning: Could not generate coverage report: %v\n", err)
	} else {
		fmt.Println(string(output))

		// Show total coverage
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "total:") {
				fmt.Printf("\n📊 Coverage Summary: %s\n", strings.TrimSpace(line))
				break
			}
		}
	}

	// Generate HTML report if requested
	if len(os.Args) > 1 && os.Args[1] == "--html" {
		fmt.Println("Generating HTML coverage report...")
		cmd = exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Could not generate HTML report: %v\n", err)
		} else {
			fmt.Println("✅ HTML coverage report saved to coverage.html")
		}
	}

	// Clean up
	os.Remove("coverage.out")

	fmt.Println("\n✅ Test suite completed successfully!")
}

func loadEnvFile() {
	envFile := ".env"

	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		fmt.Println("⚠️  .env file not found. Using system environment variables.")
		fmt.Println("To use API keys for testing, create a .env file with your API keys:")
		fmt.Println("  cp cmd/example/env.example .env")
		fmt.Println("  # Edit .env and add your API keys")
		return
	}

	fmt.Println("📁 Loading environment variables from .env file...")

	file, err := os.Open(envFile)
	if err != nil {
		fmt.Printf("❌ Error opening .env file: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Remove quotes if present
				if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
					(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
					value = value[1 : len(value)-1]
				}

				os.Setenv(key, value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("❌ Error reading .env file: %v\n", err)
		return
	}

	fmt.Println("✅ Loaded environment variables from .env file")

	// Verify key environment variables
	checkEnvVar("OPENAI_API_KEY", "OpenAI")
	checkEnvVar("ANTHROPIC_API_KEY", "Anthropic")
	checkEnvVar("GOOGLE_API_KEY", "Google")
	checkEnvVar("AZURE_OPENAI_API_KEY", "Azure OpenAI")
	checkEnvVar("COHERE_API_KEY", "Cohere")
}

func checkEnvVar(key, name string) {
	if value := os.Getenv(key); value != "" {
		fmt.Printf("✅ %s API key loaded\n", name)
	} else {
		fmt.Printf("⚠️  %s API key not found in .env file\n", name)
	}
}
