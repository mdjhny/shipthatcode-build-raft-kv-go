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
	CommandInit Command = "INIT"
	// CommandStatus          Command = "STATUS"
	// CommandBecomeCandidate Command = "BECOME_CANDIDATE"
	// CommandBecomeLeader    Command = "BECOME_LEADER"
	// CommandBecomeFollower  Command = "BECOME_FOLLOWER"
	// CommandState           Command = "STATE"
	// CommandRequestVote     Command = "REQUEST_VOTE"
	// CommandNodes           Command = "NODES"
	// CommandTimeout         Command = "TIMEOUT"
	// CommandVote            Command = "VOTE"
	// CommandResult          Command = "RESULT"
	// CommandClient      Command = "CLIENT"
	// CommandAck         Command = "ACK"
	// CommandLog         Command = "LOG"
	// CommandCommitIndex Command = "COMMIT_INDEX"
	CommandLeaderLog   Command = "LEADER_LOG"
	CommandFollowerLog Command = "FOLLOWER_LOG"
	ComandReconcile    Command = "RECONCILE"
)

type LogEntry struct {
	index   int
	term    int
	command string
}

type Node struct {
	name         string
	state        NodeState
	term         int
	numFollowers int

	logs string
}

func newNode() *Node {
	return &Node{
		state: NodeStateFollower,
	}
}

func (n *Node) handleCmd(parts []string) string {
	cmd := Command(parts[0])
	switch cmd {
	case CommandInit:
		numFollowers, err := strconv.Atoi(parts[1])
		if err != nil || numFollowers < 0 {
			return ""
		}
		n.name = "0"
		n.term = 1
		n.state = NodeStateLeader
		n.numFollowers = numFollowers
	case CommandLeaderLog:
		if len(parts) < 2 {
			return ""
		}
		n.logs = parts[1]
	case CommandFollowerLog:
		// do nothing
	case ComandReconcile:
		return n.logs
	default:
		return fmt.Sprintf("Unknown command: %s", cmd)
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
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}
