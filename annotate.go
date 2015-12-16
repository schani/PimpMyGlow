package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type command struct {
	line        string
	endLine     string
	lineNo      int
	fields      []string
	subCommands []command
}

type program []command

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

func splitLine(lineVerbatim string) []string {
	line := lineVerbatim
	if strings.Contains(line, ";") {
		line = line[0:strings.Index(line, ";")]
	}
	line = strings.Trim(line, " \t")
	fields := strings.Split(line, ",")
	for i, f := range fields {
		fields[i] = strings.Trim(f, " \t")
	}
	return fields
}

func parseLines(lines []string, startLineNo int) (commands []command, lineNo int) {
	lineNo = startLineNo
	for lineNo < len(lines) {
		lineVerbatim := lines[lineNo]
		fields := splitLine(lineVerbatim)
		if fields[0] == "E" {
			break
		}
		command, newLineNo := parseCommand(lines, lineNo, fields)
		commands = append(commands, command)
		lineNo = newLineNo
	}
	return commands, lineNo
}

func parseCommand(lines []string, startLineNo int, fields []string) (c command, lineNo int) {
	lineNo = startLineNo
	lineVerbatim := lines[lineNo]
	c = command{line: lineVerbatim, lineNo: lineNo, fields: fields}
	switch fields[0] {
	case "L":
		subCommands, newLineNo := parseLines(lines, lineNo+1)
		if newLineNo >= len(lines) {
			fmt.Fprintf(os.Stderr, "Error in program: unterminated loop\n")
			os.Exit(1)
		}
		c.subCommands = subCommands
		c.endLine = lines[newLineNo]
		lineNo = newLineNo + 1
	case "E":
		panic("cannot parse command E")
	default:
		lineNo++
	}

	return c, lineNo
}

func (c *command) hasSubCommands() bool {
	return c.fields[0] == "L"
}

func (c *command) duration() int {
	switch c.fields[0] {
	case "D":
		return parseCount(c.fields[1], c.lineNo)
	case "RAMP":
		return parseCount(c.fields[4], c.lineNo)
	case "L":
		count := parseCount(c.fields[1], c.lineNo)
		duration := 0
		for _, sc := range c.subCommands {
			duration += sc.duration()
		}
		return duration * count
	default:
		if c.hasSubCommands() {
			panic("unexpected sub-commands")
		}
		return 0
	}
}

func (c *command) print() {
	fmt.Println(c.line)
	if c.hasSubCommands() {
		for _, sc := range c.subCommands {
			sc.print()
		}
		fmt.Println(c.endLine)
	}
}

func (p program) annotateTimes() {
	time := 0
	for _, c := range p {
		c.print()
		d := c.duration()
		if d > 0 {
			time += c.duration()
			fmt.Printf("    ; time %d\n", time)
		}
	}
}

func parseProgram(r io.Reader) program {
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	commands, lineNo := parseLines(lines, 0)
	if lineNo < len(lines) {
		fmt.Fprintf(os.Stderr, "Error in line %d: E without L\n", lineNo)
		os.Exit(1)
	}

	return program(commands)
}

func main() {
	program := parseProgram(os.Stdin)
	program.annotateTimes()
}
