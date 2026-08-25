package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type NodeState string

const (
	StateLeader    NodeState = "leader"
	StateFollower  NodeState = "follower"
	StateCandidate NodeState = "candidate"
)

type NodeInfo struct {
	state   NodeState
	healthy bool
	offline bool
}

type Command string

const (
	CommandPersistTerm Command = "PERSIST_TERM"
	CommandPersistVote Command = "PERSIST_VOTE"
	CommandAppend      Command = "APPEND"
	CommandCrash       Command = "CRASH"
	CommandRecover     Command = "RECOVER"
	CommandStatus      Command = "STATUS"
)

type LogEntry struct {
	term int
	cmd  string
}

type Node struct {
	state    NodeState
	term     int
	votedFor string
	logs     []LogEntry
}

var nodeStore atomic.Value

func (n Node) String() string {
	votedFor := n.votedFor
	if votedFor == "" {
		votedFor = "none"
	}
	return fmt.Sprintf("state=%s term=%d voted_for=%s log_len=%d", n.state, n.term, votedFor, len(n.logs))
}

func newNode() *Node {
	return &Node{state: StateFollower}
}

func (n *Node) handleCmd(parts []string) string {
	cmd := Command(parts[0])
	switch cmd {
	case CommandPersistTerm:
		term, _ := strconv.Atoi(parts[1])
		n.term = term
	case CommandPersistVote:
		votedFor := parts[1]
		n.votedFor = votedFor
	case CommandAppend:
		term, _ := strconv.Atoi(parts[1])
		cmd := parts[2]
		n.logs = append(n.logs, LogEntry{term: term, cmd: cmd})
	case CommandCrash:
		if n.term > 0 {
			nodeStore.Store(*n)
		}
		n.term = 0
		n.votedFor = ""
		n.logs = nil
	case CommandRecover:
		node := nodeStore.Load().(Node)
		n.term = node.term
		n.state = StateFollower
		n.votedFor = node.votedFor
		n.logs = node.logs
	case CommandStatus:
		return n.String()
	default:
		return ""
	}
	return ""
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
		debug("line=%s, node=%+v", line, node)
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}

var (
	debugger     *os.File
	debuggerOnce sync.Once
)

func debug(format string, a ...interface{}) {
	debuggerOnce.Do(func() {
		f, err := os.Create("debug.log")
		if err != nil {
			panic(err)
		}
		debugger = f
	})
	fmt.Fprintf(debugger, format, a...)
	fmt.Fprintln(debugger)
}
