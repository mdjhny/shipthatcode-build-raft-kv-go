package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
)

type NodeState string

const (
	NodeStateFollower  NodeState = "follower"
	NodeStateCandidate NodeState = "candidate"
	NodeStateLeader    NodeState = "leader"
)

type Command string

const (
	CommandInit         Command = "INIT"
	CommandAdd          Command = "ADD"
	CommandRemove       Command = "REMOVE"
	CommandCommitOldNew Command = "COMMIT_OLD_NEW"
	CommandCommitNew    Command = "COMMIT_NEW"
	CommandMajority     Command = "MAJORITY"
)

type Node struct {
	nodesNew mapset.Set[string]
	nodesOld mapset.Set[string]
	nodesAll mapset.Set[string]
	isJoint  bool
}

func newNode() *Node {
	return &Node{
		nodesNew: mapset.NewThreadUnsafeSet[string](),
		nodesOld: mapset.NewThreadUnsafeSet[string](),
		nodesAll: mapset.NewThreadUnsafeSet[string](),
		isJoint:  false,
	}
}

func (n *Node) handleCmd(parts []string) string {
	cmd := Command(parts[0])
	switch cmd {
	case CommandInit:
		if len(parts) < 2 {
			return ""
		}
		nodes := strings.Split(parts[1], ",")
		if len(nodes) == 0 {
			return ""
		}
		n.nodesNew = mapset.NewThreadUnsafeSet(nodes...)
	case CommandAdd:
		if len(parts) < 2 {
			return ""
		}
		node := parts[1]
		n.nodesNew.Add(node)
		n.nodesOld = n.nodesNew.Clone()
		n.isJoint = true
	case CommandRemove:
		if len(parts) < 2 {
			return ""
		}
		node := parts[1]
		n.nodesNew.Remove(node)
		n.nodesOld.Remove(node)
		n.isJoint = true
	case CommandCommitOldNew:
		n.nodesAll = n.nodesNew.Union(n.nodesOld)
	case CommandCommitNew:
		n.nodesNew = n.nodesAll.Clone()
		n.nodesOld.Clear()
		n.nodesAll.Clear()
		n.isJoint = false
	case CommandMajority:
		if len(parts) < 2 {
			return ""
		}
		nodes := strings.Split(parts[1], ",")
		s := mapset.NewThreadUnsafeSet(nodes...)
		newMajority := s.Intersect(n.nodesNew).Cardinality() > n.nodesNew.Cardinality()/2
		oldMajority := s.Intersect(n.nodesOld).Cardinality() > n.nodesOld.Cardinality()/2
		majority := "NO"
		if n.isJoint {
			if newMajority && oldMajority {
				majority = "YES"
			}
		} else {
			if newMajority {
				majority = "YES"
			}
		}
		return majority
	default:
		return fmt.Sprintf("Unknown command: %s", cmd)
	}
	return ""
}

func removeNode(nodes []string, needle string) []string {
	n := 0
	for _, node := range nodes {
		if node != needle {
			nodes[n] = node
			n++
		}
	}
	return nodes[:n]
}

var debuggerOnce sync.Once
var debugger *os.File

func init() {
	debuggerOnce.Do(func() {
		f, err := os.Create("debug.log")
		if err != nil {
			panic(err)
		}
		debugger = f
	})
}

func debug(format string, a ...interface{}) {
	fmt.Fprintf(debugger, format, a...)
	fmt.Fprintln(debugger)
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
		debug("line: %s, node=%+v, result=%s", line, *node, result)
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}
