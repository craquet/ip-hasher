package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func main() {
	// Define the log file path
	logFile := "access_log"

	// Open the file
	file, err := os.Open(logFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Regular expression to match IPv4 addresses
	ipRegex := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	fmt.Println("IP Addresses found in the log:")
	for scanner.Scan() {
		line := scanner.Text()
		ip := ipRegex.FindString(line)
		if ip != "" {
			fmt.Println(ip)
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
}
