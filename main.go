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
	CommandAppend       Command = "APPEND"
	ComandCommitIndex   Command = "COMMIT_INDEX"
	CommandSnapshot     Command = "SNAPSHOT"
	CommandLogLen       Command = "LOG_LEN"
	CommandSnapshotInfo Command = "SNAPSHOT_INFO"
)

type Node struct {
	term        int // 当前 term
	commitIndex int
	logs        map[int][]string // term -> logs
}

func newNode() *Node {
	return &Node{
		logs: make(map[int][]string),
	}
}

func (n *Node) handleCmd(parts []string) string {
	cmd := Command(parts[0])
	switch cmd {
	case CommandAppend:
		if len(parts) < 3 {
			return ""
		}
		term, _ := strconv.Atoi(parts[1])
		if term < n.term {
			return ""
		}
		n.term = term
		conmmand := strings.Join(parts[2:], " ")
		n.logs[term] = append(n.logs[term], conmmand)
	case ComandCommitIndex:
		if len(parts) < 2 {
			return ""
		}
		index, _ := strconv.Atoi(parts[1])
		if index < 0 {
			return ""
		}
		n.commitIndex = index
	case CommandSnapshot:
		n.logs[n.term] = n.logs[n.term][n.commitIndex:]
	case CommandLogLen:
		return strconv.Itoa(len(n.logs[n.term]))
	case CommandSnapshotInfo:
		if n.commitIndex <= 0 {
			return "none"
		}
		return fmt.Sprintf("last_idx=%d last_term=%d", n.commitIndex, n.term)
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
