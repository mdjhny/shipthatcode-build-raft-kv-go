package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type NodeState string

const (
	NodeStateFollower  NodeState = "follower"
	NodeStateCandidate NodeState = "candidate"
	NodeStateLeader    NodeState = "leader"
)

type Command string

const (
	// CommandInit   Command = "INIT"
	CommandLog    Command = "LOG"
	CommandLeader Command = "LEADER"
	ComandCommit  Command = "COMMIT"
)

// type LogEntry struct {
// 	index   int
// 	term    int
// 	command string
// }

type Node struct {
	// name         string
	// state NodeState
	// term  int

	leader     string
	leaderTerm int
	terms      map[string][]int
}

func newNode() *Node {
	return &Node{
		// state: NodeStateFollower,
		terms: make(map[string][]int),
	}
}

func (n *Node) handleCmd(parts []string) string {
	cmd := Command(parts[0])
	switch cmd {
	// case CommandInit:
	// numFollowers, err := strconv.Atoi(parts[1])
	// if err != nil || numFollowers < 0 {
	// 	return ""
	// }
	// n.term = 1
	// n.state = NodeStateLeader
	case CommandLog:
		if len(parts) < 3 {
			return ""
		}
		node := parts[1]
		terms := make([]int, 0, len(parts)-2)
		for _, term := range parts[2:] {
			t, _ := strconv.Atoi(term)
			terms = append(terms, t)
		}
		n.terms[node] = terms
	case CommandLeader:
		if len(parts) < 3 {
			return ""
		}
		term, err := strconv.Atoi(parts[2])
		if err != nil {
			return ""
		}
		n.leader = parts[1]
		n.leaderTerm = term
	case ComandCommit:
		if len(parts) < 3 {
			return ""
		}
		index, err := strconv.Atoi(parts[1])
		if err != nil {
			return ""
		}
		term, err := strconv.Atoi(parts[2])
		if err != nil {
			return ""
		}
		// check majority
		num := 0
		for _, terms := range n.terms {
			if len(terms) >= index && terms[index-1] == term {
				num++
			}
		}
		debug("n.terms: %+v, num: %d", n.terms, num)

		if num < (len(n.terms)+1)/2 {
			return "NO"
		}

		// check term
		if term != n.leaderTerm {
			return "NO"
		}
		return "YES"
	default:
		return fmt.Sprintf("Unknown command: %s", cmd)
	}
	return ""
}

var debugLogger = bufio.NewWriter(os.Stderr)

func debug(format string, a ...interface{}) {
	defer debugLogger.Flush()
	fmt.Fprintf(debugLogger, format, a...)
	fmt.Fprintln(debugLogger)
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	node := newNode()
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		result := node.handleCmd(parts)
		if result != "" {
			fmt.Fprintln(out, result)
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}
