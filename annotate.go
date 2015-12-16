package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func parseCount(f string, lineNo int) int {
	duration, err := strconv.Atoi(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in line %d\n", lineNo)
		os.Exit(1)
	}
	if duration == 0 {
		fmt.Fprintf(os.Stderr, "Error in line %d: count can't be zero", lineNo)
		os.Exit(1)
	}
	return duration
}

func splitLine(lineVerbatim string) (command string, fields []string) {
	line := lineVerbatim
	if strings.Contains(line, ";") {
		line = line[0:strings.Index(line, ";")]
	}
	line = strings.Trim(line, " \t")
	fields = strings.Split(line, ",")
	for i, f := range fields {
		fields[i] = strings.Trim(f, " \t")
	}
	command = strings.Trim(fields[0], " \t")
	return command, fields
}

func parseLines(lines []string, startLineNo int) (duration int, lineNo int) {
	lineNo = startLineNo
	duration = 0
	for lineNo < len(lines) {
		lineVerbatim := lines[lineNo]
		command, _ := splitLine(lineVerbatim)
		if command == "E" {
			fmt.Printf("%s\n", lineVerbatim)
			return duration, lineNo + 1
		}
		subDuration, newLineNo := parseProgram(lines, lineNo)
		duration += subDuration
		lineNo = newLineNo
	}
	fmt.Fprintf(os.Stderr, "Error in program: unterminated loop\n")
	os.Exit(1)
	return -1, -1
}

func parseProgram(lines []string, startLineNo int) (duration int, lineNo int) {
	lineNo = startLineNo
	lineVerbatim := lines[lineNo]
	fmt.Printf("%s\n", lineVerbatim)
	command, fields := splitLine(lineVerbatim)
	duration = 0
	switch command {
	case "D":
		duration = parseCount(fields[1], lineNo)
		lineNo++
	case "RAMP":
		duration = parseCount(fields[4], lineNo)
		lineNo++
	case "C":
		lineNo++
	case "L":
		count := parseCount(fields[1], lineNo)
		duration, lineNo = parseLines(lines, lineNo+1)
		duration *= count
	case "E":
		fmt.Fprintf(os.Stderr, "Error in line %d: E without L\n", lineNo)
		os.Exit(1)
	default:
		lineNo++
	}

	return duration, lineNo
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	time := 0
	lineNo := 0
	for lineNo < len(lines) {
		duration, newLineNo := parseProgram(lines, lineNo)
		lineNo = newLineNo
		if duration == 0 {
			continue
		}
		time += duration
		fmt.Printf("    ; time %d\n", time)
	}
}
