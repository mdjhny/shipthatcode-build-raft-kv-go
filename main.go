package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
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
	nodesNew Set
	nodesOld Set
	nodesAll Set
	isJoint  bool
}

type Set map[string]bool

func NewSet(vals ...string) Set {
	m := make(map[string]bool, len(vals))
	for _, val := range vals {
		m[val] = true
	}
	return m
}

func (s Set) Add(v string) {
	s[v] = true
}

func (s Set) Remove(v string) {
	delete(s, v)
}

func (s Set) Clone() Set {
	t := make(Set, len(s))
	for k, v := range s {
		t[k] = v
	}
	return t
}

func (s Set) Union(a Set) Set {
	t := make(Set, len(s)+len(a))
	for k, v := range s {
		t[k] = v
	}
	for k, v := range a {
		t[k] = v
	}
	return t
}

func (s Set) Intersect(a Set) Set {
	t := s.Clone()
	for k := range s {
		if _, ok := a[k]; ok {
			t[k] = true
		}
	}
	return t
}

func (s Set) Clear() {
	s = make(Set, 0)
}

func (s Set) Cardinality() int {
	return len(s)
}

func newNode() *Node {
	return &Node{
		nodesNew: NewSet(),
		nodesOld: NewSet(),
		nodesAll: NewSet(),
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
		n.nodesNew = NewSet(nodes...)
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
		s := NewSet(nodes...)
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
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}
