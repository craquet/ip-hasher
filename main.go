package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"syscall"
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

func pollFile(logFile, outFile string, interval time.Duration) {
	fmt.Println("Opening log file...")

	file, err := os.Open(logFile)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	fd := stat.Sys().(*syscall.Stat_t).Ino

	out, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Failed to open output file:", err)
		return
	}
	defer out.Close()

	// Start at the end of the file
	offset, _ := file.Seek(0, 2)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Println("Polling log file for changes...")

	for {
		<-ticker.C

		newStat, err := os.Stat(logFile)
		if err != nil {
			fmt.Println("Failed to stat log file:", err, ". Assuming log rotation, restarting...")
			return
		}
		newFd := newStat.Sys().(*syscall.Stat_t).Ino

		if fd != newFd {
			// File has rotated
			fmt.Println("Log file has rotated")
			return
		}

		reader := bufio.NewScanner(file)
		for reader.Scan() {
			line := reader.Text()
			if len(line) < 2 {
				continue
			}
			_, err = out.WriteString(processLogLine(line) + "\n")
			if err != nil {
				fmt.Println("Failed to write line to output file:", err)
			}
		}

		if reader.Err() != nil {
			fmt.Println("Error reading log file:", err)
			return
		}

		// In case of log rotation, check if the file size has decreased
		newOffset, _ := file.Seek(0, 2)
		if newOffset < offset {
			fmt.Println("Log file rotated or truncated, reopening...")
			err := os.Truncate(outFile, 0)
			if err != nil {
				fmt.Println("Failed to truncate output file:", err)
			}

			return
		} else {
			offset = newOffset
		}
	}
}

func main() {
	for {
		pollFile(os.ExpandEnv("logs/$FILENAME_IN"), os.ExpandEnv("out/$FILENAME_OUT"), time.Second)
		time.Sleep(time.Second)
	}
}
