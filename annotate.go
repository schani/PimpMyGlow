package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type command struct {
	originalLine string
	endLine      string
	lineNo       int
	fields       []string
	subCommands  []command
}

func (c command) line() string {
	if c.originalLine != "" {
		return c.originalLine
	}
	return strings.Join(c.fields, ",")
}

func (c *command) setFields(fields []string) {
	c.fields = fields
	c.originalLine = ""
}

type program []command

func errorExit(lineNo int, format string, args ...interface{}) {
	args = append([]interface{}{lineNo + 1}, args...)
	fmt.Fprintf(os.Stderr, "Error in line %d: "+format+"\n", args...)
	os.Exit(1)
}

func parseNumber(f string, lineNo int) int {
	duration, err := strconv.Atoi(f)
	if err != nil {
		errorExit(lineNo, "Cannot parse number `%s`", f)
	}
	return duration
}

func parseCount(f string, lineNo int) int {
	duration := parseNumber(f, lineNo)
	if duration == 0 {
		errorExit(lineNo, "Count can't be zero")
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

func isBlockCommand(c string) bool {
	return c == "L" || c == "CLUBS" || c == "FILL"
}

func (c *command) hasSubCommands() bool {
	return isBlockCommand(c.fields[0])
}

func parseCommand(lines []string, startLineNo int, fields []string) (c command, lineNo int) {
	lineNo = startLineNo
	lineVerbatim := lines[lineNo]
	c = command{originalLine: lineVerbatim, lineNo: lineNo, fields: fields}
	if fields[0] == "E" {
		panic("cannot parse command E")
	}
	if isBlockCommand(fields[0]) {
		subCommands, newLineNo := parseLines(lines, lineNo+1)
		if newLineNo >= len(lines) {
			errorExit(lineNo, "Unterminated loop")
		}
		c.subCommands = subCommands
		c.endLine = lines[newLineNo]
		lineNo = newLineNo
	}
	lineNo++

	return c, lineNo
}

func commandsDuration(cs []command) int {
	duration := 0
	for _, sc := range cs {
		duration += sc.duration()
	}
	return duration
}

func (c *command) duration() int {
	switch c.fields[0] {
	case "D":
		return parseCount(c.fields[1], c.lineNo)
	case "RAMP":
		return parseCount(c.fields[4], c.lineNo)
	case "L":
		count := parseCount(c.fields[1], c.lineNo)
		duration := commandsDuration(c.subCommands)
		return duration * count
	case "FILL":
		return parseCount(c.fields[1], c.lineNo)
	case "TIME":
		errorExit(c.lineNo, "TIME not supported here")
		return -1
	default:
		if c.hasSubCommands() {
			panic(fmt.Sprintf("unexpected sub-commands in %s in line %d", c.fields[0], c.lineNo))
		}
		return 0
	}
}

func (c *command) print(w io.Writer) {
	fmt.Fprintln(w, c.line())
	if c.hasSubCommands() {
		for _, sc := range c.subCommands {
			sc.print(w)
		}
		fmt.Println(c.endLine)
	}
}

func (p program) print(w io.Writer) {
	for _, c := range p {
		c.print(w)
	}
}

func (p program) annotateTimes(w io.Writer) {
	time := 0
	for _, c := range p {
		c.print(w)
		d := c.duration()
		if d > 0 {
			time += c.duration()
			fmt.Fprintf(w, "    ; time %d\n", time)
		}
	}
}

func (p program) specializeForClub(club int) program {
	var newCommands []command
	for _, c := range p {
		switch c.fields[0] {
		case "CLUBS":
			found := false
			for _, f := range c.fields[1:len(c.fields)] {
				if parseCount(f, c.lineNo) == club {
					found = true
					break
				}
			}
			if found {
				subCommands := program(c.subCommands).specializeForClub(club)
				for _, sc := range subCommands {
					newCommands = append(newCommands, sc)
				}
			}
		default:
			newC := c
			if c.hasSubCommands() {
				newC.subCommands = program(c.subCommands).specializeForClub(club)
			}
			newCommands = append(newCommands, newC)
		}
	}
	return newCommands
}

func (p program) resolveTime() program {
	var newCommands []command
	time := 0
	for _, c := range p {
		switch c.fields[0] {
		case "TIME":
			target := parseCount(c.fields[1], c.lineNo)
			if target < time {
				errorExit(c.lineNo, "Cannot go back in time - it's already %d", time)
			}
			if target == time {
				continue
			}
			fields := []string{"D", fmt.Sprintf("%d", target-time)}
			newCommands = append(newCommands, command{fields: fields, lineNo: c.lineNo})
			time = target
		default:
			newCommands = append(newCommands, c)
			time += c.duration()
		}
	}
	return newCommands
}

type color struct {
	r, g, b int
}

var colorRegexp *regexp.Regexp

func (c color) fields() []string {
	return []string{fmt.Sprintf("%d", c.r), fmt.Sprintf("%d", c.g), fmt.Sprintf("%d", c.b)}
}

func lookupColor(colors map[string]color, description string) (bool, color, string) {
	if colorRegexp == nil {
		colorRegexp = regexp.MustCompile("^([^%]+)(\\s+(\\d+)%)?$")
	}
	matches := colorRegexp.FindStringSubmatch(description)
	name := strings.ToLower(matches[1])
	c, ok := colors[name]
	return ok, c, matches[3]
}

func resolveColor(colors map[string]color, description string, lineNo int) []string {
	ok, c, pString := lookupColor(colors, description)
	if !ok {
		errorExit(lineNo, "Color `%s` not defined", description)
	}
	if pString != "" {
		p := float64(parseNumber(pString, lineNo)) / 100.0
		c.r = int(float64(c.r) * p)
		c.g = int(float64(c.g) * p)
		c.b = int(float64(c.b) * p)
	}
	return c.fields()
}

func resolveColorInCommands(cs []command, colors map[string]color, allowDefine bool) []command {
	var newCommands []command
	for _, c := range cs {
		switch c.fields[0] {
		case "COLOR":
			if !allowDefine {
				errorExit(c.lineNo, "Can't define colors here")
			}
		case "C":
			newC := c
			if len(c.fields) == 2 {
				clr := resolveColor(colors, c.fields[1], c.lineNo)
				newC.setFields([]string{"C", clr[0], clr[1], clr[2]})
			}
			newCommands = append(newCommands, newC)
		case "RAMP":
			newC := c
			if len(c.fields) == 3 {
				clr := resolveColor(colors, c.fields[1], c.lineNo)
				newC.setFields([]string{"RAMP", clr[0], clr[1], clr[2], c.fields[2]})
			}
			newCommands = append(newCommands, newC)
		default:
			newC := c
			if c.hasSubCommands() {
				newC.subCommands = resolveColorInCommands(c.subCommands, colors, false)
			}
			newCommands = append(newCommands, newC)
		}
	}
	return newCommands
}

func gatherColorsInCommands(cs []command, colors map[string]color) {
	for _, c := range cs {
		switch c.fields[0] {
		case "COLOR":
			name := strings.ToLower(c.fields[1])
			_, ok := colors[name]
			if ok {
				errorExit(c.lineNo, "Color `%s` redefined", name)
			}
			var colorFields []string
			if len(c.fields) == 3 {
				colorFields = resolveColor(colors, c.fields[2], c.lineNo)
			} else {
				colorFields = c.fields[2:5]
			}
			colors[name] = color{
				r: parseNumber(colorFields[0], c.lineNo),
				g: parseNumber(colorFields[1], c.lineNo),
				b: parseNumber(colorFields[2], c.lineNo),
			}
		default:
			if c.hasSubCommands() {
				gatherColorsInCommands(c.subCommands, colors)
			}
		}
	}
}

func (p program) gatherColors() map[string]color {
	colors := make(map[string]color)
	gatherColorsInCommands(p, colors)
	return colors
}

func (p program) resolveColor() program {
	colors := p.gatherColors()
	return resolveColorInCommands(p, colors, true)
}

type sub struct {
	name     string
	commands []command
}

func (p program) gatherSubs() map[string]sub {
	subs := make(map[string]sub)

	i := 0
	for i < len(p) {
		if p[i].fields[0] != "DEFSUB" {
			i++
			continue
		}
		j := i + 1
		for j < len(p) {
			if p[j].fields[0] == "ENDSUB" {
				break
			}
			j++
		}
		if j == len(p) {
			errorExit(p[i].lineNo, "DEFSUB without ENDSUB")
		}

		name := strings.ToLower(p[i].fields[1])
		sub := sub{name: name, commands: p[i+1 : j]}
		subs[name] = sub

		i = j + 1
	}

	return subs
}

type label struct {
	name   string
	fields []string
	start  int
	end    int
}

func cannotInterpret(expr ast.Expr, lineNo int) {
	errorExit(lineNo, "Cannot interpret %T expression `%v`", expr, expr)
}

func lookupLabel(labels map[string]label, name string, lineNo int) label {
	label, ok := labels[name]
	if !ok {
		errorExit(lineNo, "Unknown label `%s`", name)
	}
	return label
}

func interpretExpr(expr ast.Expr, labels map[string]label, definitions map[string]int, lineNo int) int {
	switch expr := expr.(type) {
	case *ast.BasicLit:
		if expr.Kind == token.INT {
			return parseNumber(expr.Value, lineNo)
		}
	case *ast.Ident:
		v, ok := definitions[expr.Name]
		if ok {
			return v
		}
		return lookupLabel(labels, expr.Name, lineNo).start
	case *ast.UnaryExpr:
		if expr.Op == token.SUB {
			ident, ok := expr.X.(*ast.Ident)
			if !ok {
				cannotInterpret(expr, lineNo)
			}
			return lookupLabel(labels, ident.Name, lineNo).end
		} else if expr.Op == token.AND {
			ident, ok := expr.X.(*ast.Ident)
			if !ok {
				cannotInterpret(expr, lineNo)
			}
			label := lookupLabel(labels, ident.Name, lineNo)
			return label.end - label.start
		}
	case *ast.BinaryExpr:
		if expr.Op == token.QUO {
			left := interpretExpr(expr.X, labels, definitions, lineNo)
			right := interpretExpr(expr.Y, labels, definitions, lineNo)
			return left / right
		}
	}
	cannotInterpret(expr, lineNo)
	return -1
}

func (p program) resolveExprs(labels map[string]label, definitions map[string]int) program {
	var newCommands []command
	for _, c := range p {
		newC := c
		if c.fields[0] == "D" || c.fields[0] == "TIME" || c.fields[0] == "RAMP" || c.fields[0] == "L" || c.fields[0] == "FILL" {
			exprField := len(c.fields) - 1
			newC.setFields(make([]string, len(c.fields)))
			copy(newC.fields, c.fields)

			expr, err := parser.ParseExpr(c.fields[exprField])
			if err != nil {
				errorExit(c.lineNo, "Parse error: %s", err.Error())
			}
			result := interpretExpr(expr, labels, definitions, c.lineNo)

			newC.fields[exprField] = strconv.FormatInt(int64(result), 10)
		}
		if c.hasSubCommands() {
			newC.subCommands = program(c.subCommands).resolveExprs(labels, definitions)
		}
		newCommands = append(newCommands, newC)
	}
	return newCommands
}

func (c command) fill(duration int) []command {
	//fmt.Fprintf(os.Stderr, "filling command to %d\n", duration)
	//c.print(os.Stderr)

	if c.duration() < duration {
		panic("can't fill a command that's too short")
	}

	newC := c
	newC.setFields(make([]string, len(c.fields)))
	copy(newC.fields, c.fields)

	if c.fields[0] == "D" || c.fields[0] == "RAMP" {
		// FIXME: we shouldn't just shorten a ramp, it might
		// produce a weird effect
		newC.fields[len(c.fields)-1] = strconv.FormatInt(int64(duration), 10)
		return []command{newC}
	}

	if c.fields[0] != "L" {
		errorExit(c.lineNo, "Illegal command `%s` within `FILL`", c.fields[0])
	}

	var newCommands []command

	loopDuration := commandsDuration(c.subCommands)
	numIterations := duration / loopDuration
	//fmt.Fprintf(os.Stderr, "loop is %d, doing %d iterations\n", loopDuration, numIterations)
	if numIterations > 0 {
		newC.fields[1] = strconv.FormatInt(int64(numIterations), 10)
		newCommands = append(newCommands, newC)
	}

	left := duration - loopDuration*numIterations
	if left > 0 {
		newCommands = append(newCommands, fillCommands(c.subCommands, left)...)
	}

	if commandsDuration(newCommands) != duration {
		panic("we can't do commands fill math")
	}

	return newCommands
}

func fillCommands(commands []command, duration int) []command {
	var newCommands []command
	time := 0
	for _, sc := range commands {
		left := duration - time
		if left <= 0 {
			break
		}

		scDuration := sc.duration()
		if scDuration <= left {
			newCommands = append(newCommands, sc)
			time += scDuration
			continue
		}

		filledCommands := sc.fill(left)
		newCommands = append(newCommands, filledCommands...)
		time += left
		break
	}
	left := duration - time
	if left < 0 {
		panic("I can't do fill math")
	}
	if left > 0 {
		newCommands = append(newCommands, command{fields: []string{"D", strconv.FormatInt(int64(left), 10)}})
	}
	if commandsDuration(newCommands) != duration {
		//fmt.Fprintf(os.Stderr, "duration %d should be %d\n", commandsDuration(newCommands), duration)
		//program(newCommands).print(os.Stderr)
		panic("We filled incorrectly")
	}
	return newCommands
}

func (p program) resolveFill() program {
	var newCommands []command
	for _, c := range p {
		if !c.hasSubCommands() {
			newCommands = append(newCommands, c)
			continue
		}

		subCommands := program(c.subCommands).resolveFill()
		if c.fields[0] != "FILL" {
			newC := c
			newC.subCommands = subCommands
			newCommands = append(newCommands, newC)
			continue
		}

		duration := parseCount(c.fields[1], c.lineNo)
		newCommands = append(newCommands, fillCommands(subCommands, duration)...)
	}
	return newCommands
}

func parseProgram(r io.Reader) program {
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	commands, lineNo := parseLines(lines, 0)
	if lineNo < len(lines) {
		errorExit(lineNo, "E without L")
	}

	return commands
}

// XMLLabel must be exported to work with encoding/xml.
type XMLLabel struct {
	Title string  `xml:"title,attr"`
	Start float64 `xml:"t,attr"`
	End   float64 `xml:"t1,attr"`
}

// XMLProject must be exported to work with encoding/xml.
type XMLProject struct {
	Labels []XMLLabel `xml:"labeltrack>label"`
}

func readLabels(reader io.Reader) ([]label, error) {
	var project XMLProject
	if err := xml.NewDecoder(reader).Decode(&project); err != nil {
		return nil, err
	}

	var labels []label
	for _, l := range project.Labels {
		fields := strings.Split(l.Title, ":")
		for i, f := range fields {
			fields[i] = strings.TrimSpace(f)
		}

		labels = append(labels, label{name: l.Title, fields: fields, start: int(l.Start * 100), end: int(l.End * 100)})
	}
	return labels, nil
}

func mapFromLabels(labels []label) map[string]label {
	labelsMap := make(map[string]label)
	for _, l := range labels {
		_, ok := labelsMap[l.name]
		if ok {
			fmt.Fprintf(os.Stderr, "Error: Label %s defined more than once\n", l.name)
			os.Exit(1)
		}
		labelsMap[l.name] = l
	}
	return labelsMap
}

type timeline []label

func (ls timeline) Len() int {
	return len(ls)
}

func (ls timeline) Less(i, j int) bool {
	return ls[i].start < ls[j].start
}

func (ls timeline) Swap(i, j int) {
	tmp := ls[i]
	ls[i] = ls[j]
	ls[j] = tmp
}

func (l label) clubs() ([]string, []string) {
	fields := l.fields
	matches, err := regexp.MatchString("^[cC]\\s*\\d+(,\\d+)*$", fields[0])
	if err != nil {
		panic("Messed up regular expression")
	}
	var clubs []string
	if matches {
		clubs = strings.Split(fields[0][1:len(fields[0])], ",")
		fields = fields[1:len(fields)]
	}
	return clubs, fields
}

func (ls timeline) program(colors map[string]color, subs map[string]sub) program {
	var commands []command
	for name, c := range colors {
		commands = append(commands, command{fields: append([]string{"COLOR", name}, c.fields()...)})
	}
	commands = append(commands, command{fields: []string{"C", "0", "0", "0"}})
	for _, l := range ls {
		duration := l.end - l.start

		var labelCommands []command

		labelCommands = append(labelCommands, command{fields: []string{"TIME", strconv.FormatInt(int64(l.start), 10)}})

		clubs, fields := l.clubs()

		if len(fields) == 1 {
			name := strings.ToLower(fields[0])

			ok, _, _ := lookupColor(colors, name)
			if ok {
				colorCommand := command{fields: []string{"C", name}}
				labelCommands = append(labelCommands, colorCommand)
			} else {
				sub, ok := subs[name]
				if !ok {
					errorExit(-1, "`%s` is not a color or a sub", name)
				}

				definitions := map[string]int{"duration": duration}
				subCommands := program(sub.commands).resolveExprs(make(map[string]label), definitions)

				fillCommand := command{
					fields:      []string{"FILL", strconv.FormatInt(int64(duration), 10)},
					endLine:     "E",
					subCommands: subCommands}

				labelCommands = append(labelCommands, fillCommand)
			}
		} else if len(fields) > 2 && strings.ToLower(fields[0]) == "ramp" {
			colorFields := fields[1:len(fields)]
			for _, c := range colorFields {
				ok, _, _ := lookupColor(colors, c)
				if !ok {
					errorExit(-1, "Unknown color `%s`", c)
				}
			}

			colorCommand := command{fields: []string{"C", colorFields[0]}}
			labelCommands = append(labelCommands, colorCommand)

			rampFields := colorFields[1:len(colorFields)]
			timeSoFar := 0
			for i, c := range rampFields {
				timeTarget := (i + 1) * duration / len(rampFields)
				time := timeTarget - timeSoFar

				rampCommand := command{fields: []string{"RAMP", c, strconv.FormatInt(int64(time), 10)}}
				labelCommands = append(labelCommands, rampCommand)

				timeSoFar += time
			}

			if timeSoFar != duration {
				panic("I can't do math")
			}
		} else {
			errorExit(-1, "Incorrect label `%s`", l.name)
		}

		labelCommands = append(labelCommands, command{fields: []string{"TIME", strconv.FormatInt(int64(l.end), 10)}})
		labelCommands = append(labelCommands, command{fields: []string{"C", "0", "0", "0"}})

		if len(clubs) > 0 {
			clubCommand := command{fields: append([]string{"CLUBS"}, clubs...), endLine: "E"}
			clubCommand.subCommands = labelCommands

			labelCommands = []command{clubCommand}
		}

		commands = append(commands, labelCommands...)
	}

	commands = append(commands, command{fields: []string{"END"}})
	return commands
}

func (ls timeline) checkConsistency() {
	allActive := 0
	var clubsActive []int

	for _, l := range ls {
		clubs, _ := l.clubs()
		var clubsString string
		if len(clubs) == 0 {
			clubsString = "all clubs"
		} else {
			clubsString = "clubs " + strings.Join(clubs, ", ")
		}

		if l.start < allActive {
			errorExit(-1, "Label collision for %s at time %d", clubsString, l.start)
		}

		if len(clubs) == 0 {
			for i, active := range clubsActive {
				if l.start < active {
					errorExit(-1, "Label collision for club %d at time %d", i, l.start)
				}
			}
			allActive = l.end
		}
		for _, c := range clubs {
			i := parseNumber(c, -1)
			for i >= len(clubsActive) {
				clubsActive = append(clubsActive, 0)
			}

			if l.start < clubsActive[i] {
				errorExit(-1, "Label collision for club %d at time %d", i, l.start)
			}

			clubsActive[i] = l.end
		}
	}
}

func main() {
	var err error

	audacityFlag := flag.String("audacity", "", "Audacity file path")
	clubFlag := flag.Int("club", 0, "Club to specialize for")
	inputFlag := flag.String("input", "-", "Input file")
	outputFlag := flag.String("output", "-", "Output file")
	timelineFlag := flag.Bool("timeline", false, "Produce program from timeline")

	flag.Parse()

	if *clubFlag < 0 {
		fmt.Fprintf(os.Stderr, "Error: Club can't be negative\n")
		os.Exit(1)
	}

	var labels []label
	if *audacityFlag != "" {
		file, err := os.Open(*audacityFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Can't open audacity file `%s`: %s\n", *audacityFlag, err.Error())
			os.Exit(1)
		}
		defer file.Close()

		labels, err = readLabels(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading Audacity file `%s`: %s\n", *audacityFlag, err.Error())
			os.Exit(1)
		}
	}

	inFile := os.Stdin
	if *inputFlag != "-" {
		inFile, err = os.Open(*inputFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file `%s`: %s\n", *inputFlag, err.Error())
			os.Exit(1)
		}
		defer inFile.Close()
	}
	inputProgram := parseProgram(inFile)

	var labelsMap map[string]label

	if *timelineFlag {
		sort.Sort(timeline(labels))
		timeline(labels).checkConsistency()
		colors := inputProgram.gatherColors()
		subs := inputProgram.gatherSubs()
		inputProgram = timeline(labels).program(colors, subs)
	} else {
		labelsMap = mapFromLabels(labels)
	}

	specialized := inputProgram
	if *clubFlag != 0 {
		specialized = inputProgram.specializeForClub(*clubFlag)
	}
	colored := specialized.resolveColor()
	delabeled := colored.resolveExprs(labelsMap, make(map[string]int))
	filled := delabeled.resolveFill()
	finalProgram := filled.resolveTime()

	outFile := os.Stdout
	if *outputFlag != "-" {
		outFile, err = os.Create(*outputFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening output file `%s`: %s\n", *outputFlag, err.Error())
			os.Exit(1)
		}
		defer outFile.Close()
	}

	finalProgram.annotateTimes(outFile)
}
