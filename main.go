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
	CommandInit            Command = "INIT"
	CommandStatus          Command = "STATUS"
	CommandBecomeCandidate Command = "BECOME_CANDIDATE"
	CommandBecomeLeader    Command = "BECOME_LEADER"
	CommandBecomeFollower  Command = "BECOME_FOLLOWER"
	CommandState           Command = "STATE"
	CommandRequestVote     Command = "REQUEST_VOTE"
)

type Node struct {
	name         string
	state        NodeState
	term         int64
	votedFor     string
	lastLogIndex int64
	lastLogTerm  int64
}

func newNode() *Node {
	return &Node{
		state:        NodeStateFollower,
		term:         0,
		votedFor:     "none",
		lastLogIndex: 0,
		lastLogTerm:  0,
	}
}

func (n *Node) handleCmd(parts []string) string {
	cmd := Command(parts[0])
	switch cmd {
	case CommandInit:
		n.name = parts[1]
		n.state = NodeStateFollower
		n.term = 0
		n.votedFor = "none"
	case CommandStatus:
		return fmt.Sprintf("state=%s term=%d voted_for=%s", n.state, n.term, n.votedFor)
	case CommandBecomeCandidate:
		n.state = NodeStateCandidate
		n.term++
		n.votedFor = n.name
	case CommandBecomeLeader:
		if n.state != NodeStateCandidate {
			return ""
		}
		n.state = NodeStateLeader
	case CommandBecomeFollower:
		term, _ := strconv.ParseInt(parts[1], 10, 64)
		if term < n.term {
			return ""
		}
		n.term = term
		n.state = NodeStateFollower
		n.votedFor = "none"
	case CommandState:
		if len(parts) < 5 {
			return ""
		}
		n.term, _ = strconv.ParseInt(parts[1], 10, 64)
		n.votedFor = parts[2]
		n.lastLogIndex, _ = strconv.ParseInt(parts[3], 10, 64)
		n.lastLogTerm, _ = strconv.ParseInt(parts[4], 10, 64)
	case CommandRequestVote:
		if len(parts) < 5 {
			return ""
		}

		candidateID := parts[1]
		candidateTerm, _ := strconv.ParseInt(parts[2], 10, 64)
		if candidateTerm < n.term {
			return "NO"
		}
		if candidateTerm == n.term && n.votedFor != "none" && n.votedFor != candidateID {
			return "NO"
		}
		candidateLastLogIndex, _ := strconv.ParseInt(parts[3], 10, 64)
		candidateLastLogTerm, _ := strconv.ParseInt(parts[4], 10, 64)
		if candidateLastLogTerm < n.lastLogTerm ||
			(candidateLastLogTerm == n.lastLogTerm && candidateLastLogIndex < n.lastLogIndex) {
			return "NO"
		}
		// 投票通过
		n.votedFor = candidateID
		n.term = candidateTerm
		n.lastLogIndex = candidateLastLogIndex
		n.lastLogTerm = candidateLastLogTerm
		n.state = NodeStateFollower
		return "YES"
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
}
