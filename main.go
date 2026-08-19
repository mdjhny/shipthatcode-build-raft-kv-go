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
	CommandNodes           Command = "NODES"
	CommandTimeout         Command = "TIMEOUT"
	CommandVote            Command = "VOTE"
	CommandResult          Command = "RESULT"
)

type Node struct {
	name         string
	state        NodeState
	term         int64
	votedFor     string
	lastLogIndex int64
	lastLogTerm  int64

	peers         []string
	votesReceived map[string]map[string]bool
}

func newNode() *Node {
	return &Node{
		state:        NodeStateFollower,
		term:         0,
		votedFor:     "",
		lastLogIndex: 0,
		lastLogTerm:  0,
		peers:        []string{},
	}
}

var candidateStore string

func (n *Node) handleCmd(parts []string) string {
	cmd := Command(parts[0])
	switch cmd {
	case CommandInit:
		n.name = parts[1]
		n.state = NodeStateFollower
		n.term = 0
		n.votedFor = ""
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
		n.votedFor = ""
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
		if candidateTerm == n.term && n.votedFor != "" && n.votedFor != candidateID {
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
	case CommandNodes:
		numNodes, err := strconv.Atoi(parts[1])
		if err != nil || numNodes <= 0 {
			return ""
		}
		n.peers = make([]string, 0, numNodes)
		for i := 0; i < numNodes; i++ {
			nodeID := strconv.Itoa(i)
			n.peers = append(n.peers, nodeID)
		}
		n.name = "1"
	case CommandTimeout:
		peer := parts[1]
		n.votesReceived = make(map[string]map[string]bool)
		if peer == n.name {
			n.state = NodeStateCandidate
			n.addVote(n.name, peer)
		} else {
			n.state = NodeStateFollower
		}
		candidateStore = peer
	case CommandVote:
		voter := parts[1]
		candidate := parts[2]
		if candidate == candidateStore {
			n.addVote(voter, candidate)
		}
	case CommandResult:
		for candidate, voters := range n.votesReceived {
			if len(voters) > len(n.peers)/2 {
				return fmt.Sprintf("node %s", candidate)
			}
		}
		return "NO_LEADER"
	default:
		return fmt.Sprintf("Unknown command: %s", cmd)
	}
	return ""
}

func (n *Node) addVote(voter, candidate string) {
	if n.votesReceived == nil {
		n.votesReceived = make(map[string]map[string]bool)
	}
	if n.votesReceived[candidate] == nil {
		n.votesReceived[candidate] = make(map[string]bool)
	}
	n.votesReceived[candidate][voter] = true
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
