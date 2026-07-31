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
)

type Node struct {
	name     string
	state    NodeState
	term     int64
	votedFor string
}

func newNode() *Node {
	return &Node{
		state:    NodeStateFollower,
		term:     0,
		votedFor: "none",
	}
}

func (n *Node) handleCmd(parts []string) {
	cmd := Command(parts[0])
	switch cmd {
	case CommandInit:
		n.name = parts[1]
		n.state = NodeStateFollower
		n.term = 0
		n.votedFor = "none"
	case CommandStatus:
		fmt.Printf("state=%s term=%d voted_for=%s\n", n.state, n.term, n.votedFor)
	case CommandBecomeCandidate:
		n.state = NodeStateCandidate
		n.term++
		n.votedFor = n.name
	case CommandBecomeLeader:
		if n.state != NodeStateCandidate {
			fmt.Println("Cannot become leader: not in candidate state")
			return
		}
		n.state = NodeStateLeader
	case CommandBecomeFollower:
		term, _ := strconv.ParseInt(parts[1], 10, 64)
		if term < n.term {
			fmt.Println("Cannot become follower: term is less than current term")
			return
		}
		n.term = term
		n.state = NodeStateFollower
		n.votedFor = "none"
	default:
		fmt.Println("Unknown command:", cmd)
	}
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
		node.handleCmd(parts)
	}
}
