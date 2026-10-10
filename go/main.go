// Command qwer is the original Go implementation of qwer.
//
// Deprecated: superseded by the Zig implementation in the repository root
// (build.zig, src/main.zig). Kept for reference only.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/pflag"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type Command struct {
	Name     string
	Code     string
	Level    int
	Children map[string]*Command
}

func parseMarkdown(input []byte) *Command {
	md := goldmark.New()
	reader := text.NewReader(input)
	doc := md.Parser().Parse(reader)

	root := &Command{
		Name:     "root",
		Children: make(map[string]*Command),
	}

	var commandStack []*Command
	var lastHeading *Command

	walkAST(doc, input, root, &commandStack, &lastHeading)
	return root
}

func walkAST(n ast.Node, source []byte, root *Command, commandStack *[]*Command, lastHeading **Command) {
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch v := child.(type) {
		case *ast.Heading:
			content := extractText(v, source)

			// Skip h1 headings (level 1) as they are just section titles.
			// Reset lastHeading so a code block under an h1 doesn't attach
			// to the previous command.
			if v.Level == 1 {
				*lastHeading = nil
				continue
			}

			cmd := &Command{
				Name:     strings.ToLower(strings.ReplaceAll(content, " ", "-")),
				Level:    v.Level,
				Children: make(map[string]*Command),
			}

			// Pop commands from stack that are same or higher level
			for len(*commandStack) > 0 && (*commandStack)[len(*commandStack)-1].Level >= v.Level {
				*commandStack = (*commandStack)[:len(*commandStack)-1]
			}

			// Add to parent
			if len(*commandStack) == 0 {
				root.Children[cmd.Name] = cmd
			} else {
				parent := (*commandStack)[len(*commandStack)-1]
				parent.Children[cmd.Name] = cmd
			}

			*commandStack = append(*commandStack, cmd)
			*lastHeading = cmd

		case *ast.FencedCodeBlock, *ast.CodeBlock:
			if *lastHeading != nil {
				lines := child.Lines()
				var code bytes.Buffer
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					code.Write(line.Value(source))
				}
				(*lastHeading).Code = strings.TrimSpace(code.String())
			}
		}

		walkAST(child, source, root, commandStack, lastHeading)
	}
}

func extractText(n ast.Node, source []byte) string {
	var buf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == ast.KindText {
			segment := c.(*ast.Text).Segment
			buf.Write(segment.Value(source))
		} else {
			buf.WriteString(extractText(c, source))
		}
	}
	return buf.String()
}

func findCommand(root *Command, args []string) *Command {
	current := root
	for _, arg := range args {
		if cmd, ok := current.Children[arg]; ok {
			current = cmd
		} else {
			// Try fuzzy match
			var childNames []string
			for name := range current.Children {
				childNames = append(childNames, name)
			}
			matches := fuzzy.Find(arg, childNames)
			if len(matches) == 1 {
				current = current.Children[matches[0]]
			} else {
				// 0 matches or 2+ matches: don't run anything
				return nil
			}
		}
	}
	return current
}

func executeCommand(code string, extraArgs []string) error {
	// Add default shebang if not present
	if !strings.HasPrefix(code, "#!") {
		code = "#!/usr/bin/env bash\n" + code
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "qwer-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write code to temp file
	if _, err := tmpFile.WriteString(code); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make temp file executable: %w", err)
	}

	fmt.Printf("Running: %s", tmpFile.Name())
	if len(extraArgs) > 0 {
		fmt.Printf(" %s", strings.Join(extraArgs, " "))
	}
	fmt.Println()

	// Run the script with extra arguments
	args := append([]string{tmpFile.Name()}, extraArgs...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func findAllCommandFiles() ([]string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to get current directory: %w", err)
	}

	var qwerFiles []string
	dir := currentDir

	// Find all qwer.md files from current directory up to the git root
	// (first directory containing .git), or the filesystem root if not
	// inside a git repository.
	for {
		// Check for QWER.md first, then qwer.md
		qwerPath := filepath.Join(dir, "QWER.md")
		if _, err := os.Stat(qwerPath); err == nil {
			qwerFiles = append(qwerFiles, qwerPath)
		} else {
			qwerLowerPath := filepath.Join(dir, "qwer.md")
			if _, err := os.Stat(qwerLowerPath); err == nil {
				qwerFiles = append(qwerFiles, qwerLowerPath)
			}
		}

		// Stop at the git root
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	if len(qwerFiles) == 0 {
		return nil, fmt.Errorf("no command file found (QWER.md or qwer.md) in current directory or parent directories")
	}

	// Reverse the slice so parent files come first
	for i, j := 0, len(qwerFiles)-1; i < j; i, j = i+1, j-1 {
		qwerFiles[i], qwerFiles[j] = qwerFiles[j], qwerFiles[i]
	}

	return qwerFiles, nil
}

func mergeCommands(parent, child *Command) {
	// Merge child commands into parent, child commands override parent commands
	for name, childCmd := range child.Children {
		if parentCmd, exists := parent.Children[name]; exists {
			// If command exists in parent, merge recursively
			// Child's code overrides parent's code if present
			if childCmd.Code != "" {
				parentCmd.Code = childCmd.Code
			}
			mergeCommands(parentCmd, childCmd)
		} else {
			// Add new command from child
			parent.Children[name] = childCmd
		}
	}
}

func listCommands(cmd *Command, prefix string) {
	for name, child := range cmd.Children {
		fullPath := prefix
		if fullPath != "" {
			fullPath += " "
		}
		fullPath += name

		if child.Code != "" {
			fmt.Printf("%s\n", fullPath)
		}

		if len(child.Children) > 0 {
			listCommands(child, fullPath)
		}
	}
}

func main() {
	// Define flags
	listFlag := pflag.Bool("list", false, "List all available commands")
	fileFlag := pflag.StringP("file", "f", "", "Markdown file to read commands from")
	helpFlag := pflag.BoolP("help", "h", false, "Show help")

	// Split args at "--": everything after it is passed to the script
	// verbatim and never parsed as flags or command names.
	argv := os.Args[1:]
	var extraArgs []string
	for i, arg := range argv {
		if arg == "--" {
			extraArgs = argv[i+1:]
			argv = argv[:i]
			break
		}
	}
	pflag.CommandLine.Parse(argv)

	// Show help
	if *helpFlag {
		fmt.Println(`qwer - markdown-defined command runner (alternative to make, just, xc)

qwer reads commands from a markdown file (QWER.md or qwer.md) and runs them.
Each "## heading" defines a command; its fenced code block is the script.
A "### heading" nested under it defines a subcommand. Scripts run with
"#!/usr/bin/env bash" by default, or any interpreter via a shebang line.

Usage:
  qwer                                   List all available commands
  qwer [command] [subcommand...]         Run a command
  qwer [command] [args...]               Extra args are passed to the script ("$@")
  qwer --list                            List all available commands
  qwer --file FILE [command]             Use a specific markdown file

Command file:
  QWER.md is preferred over qwer.md. Files are searched from the current
  directory up to the git repository root (or the filesystem root when not
  in a git repo); commands in nested files override parent ones.

  Example QWER.md:
    ## build
    ` + "```" + `
    go build ./...
    ` + "```" + `

    ## deploy
    ### prod
    ` + "```" + `
    #!/usr/bin/env python3
    print("deploying prod")
    ` + "```" + `

  This defines commands "build" and "deploy prod".

Command names are the heading text lowercased with spaces replaced by "-".
Fuzzy matching is supported: "qwer bld" runs "build" if it is the only match.

Flags:`)
		pflag.PrintDefaults()
		fmt.Println(`
Examples:
  qwer build
  qwer deploy prod
  qwer test -- --verbose
  qwer --file docs.md build`)
		os.Exit(0)
	}

	// Determine which files to read
	var qwerFiles []string
	var err error

	if *fileFlag != "" {
		// If a specific file is provided, use only that file
		qwerFiles = []string{*fileFlag}
	} else {
		// Search for QWER.md or qwer.md recursively up through parent git directories
		qwerFiles, err = findAllCommandFiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Parse and merge all qwer.md files
	root := &Command{
		Name:     "root",
		Children: make(map[string]*Command),
	}

	for _, markdownFile := range qwerFiles {
		// Read markdown file
		content, err := os.ReadFile(markdownFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", markdownFile, err)
			os.Exit(1)
		}

		// Parse the markdown file
		fileCommands := parseMarkdown(content)

		// Merge commands (later files override earlier ones)
		mergeCommands(root, fileCommands)
	}

	// List commands if requested
	if *listFlag {
		listCommands(root, "")
		os.Exit(0)
	}

	// Non-flag arguments before "--" form the command path
	cmdArgs := pflag.Args()

	if len(cmdArgs) == 0 {
		listCommands(root, "")
		os.Exit(0)
	}

	// Try to find the longest matching command path
	var cmd *Command
	var commandPathLength int

	for i := len(cmdArgs); i > 0; i-- {
		testPath := cmdArgs[:i]
		if foundCmd := findCommand(root, testPath); foundCmd != nil {
			cmd = foundCmd
			commandPathLength = i
			break
		}
	}

	if cmd == nil {
		fmt.Fprintf(os.Stderr, "Command not found: %s\n", strings.Join(cmdArgs, " "))
		listCommands(root, "")
		os.Exit(1)
	}

	// Any args beyond the command path become extra args
	if commandPathLength < len(cmdArgs) {
		remainingArgs := cmdArgs[commandPathLength:]
		extraArgs = append(remainingArgs, extraArgs...)
	}

	if cmd.Code == "" {
		fmt.Fprintf(os.Stderr, "No code block found for command: %s\n", strings.Join(cmdArgs[:commandPathLength], " "))
		if len(cmd.Children) > 0 {
			listCommands(cmd, strings.Join(cmdArgs[:commandPathLength], " "))
		}
		os.Exit(1)
	}

	if err := executeCommand(cmd.Code, extraArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Command failed: %v\n", err)
		os.Exit(1)
	}
}
