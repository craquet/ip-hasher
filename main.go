package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"time"
)

func processLogFile() {
	// Define the log file path
	logFile := "access.log"
	hashedLogFile := "access_hashed.log"

	// Open the file
	file, err := os.Open(logFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}(file)

	fileHashed, errHashed := os.Create(hashedLogFile)
	if errHashed != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer func(fileHashed *os.File) {
		err := fileHashed.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}(fileHashed)

	// Regular expression to match IPv4 addresses
	ipRegex := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	// https://stackoverflow.com/questions/53497/regular-expression-that-matches-valid-ipv6-addresses
	ip6Regex := regexp.MustCompile(`\b(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\b`)

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	fmt.Println("IP Addresses found in the log:")
	for scanner.Scan() {
		line := scanner.Text()
		hashedLine := ipRegex.ReplaceAllStringFunc(line, func(s string) string {
			hash := sha256.Sum256([]byte(s))
			return fmt.Sprintf("%x", hash)
		})
		hashedLine = ip6Regex.ReplaceAllStringFunc(hashedLine, func(s string) string {
			hash := sha256.Sum256([]byte(s))
			return fmt.Sprintf("%x", hash)
		})

		wrote, err := fileHashed.WriteString(hashedLine + "\n")
		if err != nil {
			fmt.Printf("Error adding line to file: %v\n", err)
			return
		}
		fmt.Printf("Wrote %d bytes to file\n", wrote)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
}

func main() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		processLogFile()
		<-ticker.C
	}
}
