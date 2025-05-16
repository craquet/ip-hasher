package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"os"
	"regexp"
	"time"
)

func processLogLine(line string) string {
	// Regular expression to match IPv4 addresses
	ipRegex := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	// https://stackoverflow.com/questions/53497/regular-expression-that-matches-valid-ipv6-addresses
	ip6Regex := regexp.MustCompile(`\b(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\b`)

	hashedLine := ipRegex.ReplaceAllStringFunc(line, func(s string) string {
		hash := sha256.Sum256([]byte(s))
		return fmt.Sprintf("%x", hash)
	})
	hashedLine = ip6Regex.ReplaceAllStringFunc(hashedLine, func(s string) string {
		hash := sha256.Sum256([]byte(s))
		return fmt.Sprintf("%x", hash)
	})

	return hashedLine
}

func tailLogFile(logFile string, outFile string) {
	fmt.Println("Opening log file:", logFile)

	// Open the log file
	file, err := os.Open(logFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Open the log file
	out, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Seek to the end of the file to start tailing
	_, seekErr := file.Seek(0, 2)
	if seekErr != nil {
		return
	}

	// Initialize a file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	// Start watching the directory instead of the file to catch rotations
	err = watcher.Add(".")
	if err != nil {
		panic(err)
	}

	fmt.Println("Tailing log file...")

	// Watch for file changes
	for {
		select {
		case event := <-watcher.Events:
			if event.Name == logFile {
				if event.Op.Has(fsnotify.Write) {
					// Read new lines
					reader := bufio.NewScanner(file)
					readLines := 0
					for reader.Scan() {
						line := reader.Text()

						// Skip lines that are too short, these will usually be empty lines with just \n
						if len(line) < 2 {
							continue
						}

						processedLine := processLogLine(line)
						_, err = out.WriteString(processedLine + "\n")
						if err != nil {
							fmt.Println("Could not write line to output:", err)
						}
						readLines++
					}
					fmt.Println("Read", readLines, "lines from log file")
				}
				if event.Op.Has(fsnotify.Remove) || event.Op.Has(fsnotify.Rename) {
					fmt.Println("Log file rotated, waiting to reopen...")

					_ = file.Close()
					_ = watcher.Close()
					_ = out.Close()

					err = os.Remove(outFile)
					if err != nil {
						fmt.Println("Could not remove output file:", err)
					}

					// Wait for the file to be recreated
					for {
						time.Sleep(1 * time.Second)
						if _, err := os.Stat(logFile); err == nil {
							return
						}
					}
				}
			}
		case err := <-watcher.Errors:
			fmt.Println("File system watch error:", err)
		}
	}
}

func main() {
	for {
		tailLogFile("access.log", "access_hashed.log")
	}
}
