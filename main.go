package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
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
}

type Command string

const (
	CommandInit    Command = "INIT"
	CommandTimeout Command = "TIMEOUT"
	CommandCrash   Command = "CRASH"
	CommandRecover Command = "RECOVER"
	CommandWrite   Command = "WRITE"
	CommandRead    Command = "READ"
	CommandState   Command = "STATE"
)

type KV struct {
	keys []string
	m    map[string]string
}

func (k *KV) String() string {
	if len(k.m) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(k.keys))
	for _, key := range k.keys {
		parts = append(parts, fmt.Sprintf("%s=%v", key, k.m[key]))
	}

	return strings.Join(parts, ",")
}

func (k *KV) Set(key, val string) {
	k.keys = append(k.keys, key)
	k.m[key] = val
}

func (k *KV) Get(key string) string {
	v, ok := k.m[key]
	if !ok {
		return "NIL"
	}
	return v
}

type Node struct {
	nodes  map[string]NodeInfo
	term   int
	leader string
	kv     *KV
}

func newNode() *Node {
	return &Node{}
}

func (n *Node) handleCmd(parts []string) string {
	cmd := Command(parts[0])
	switch cmd {
	case CommandInit:
		num, _ := strconv.Atoi(parts[1])
		n.nodes = make(map[string]NodeInfo, num)
		for i := 0; i < num; i++ {
			n.nodes[strconv.Itoa(i+1)] = NodeInfo{state: StateFollower, healthy: true}
		}
		n.term = 0
		n.kv = &KV{m: map[string]string{}}
	case CommandTimeout:
		node := parts[1]
		n.term++
		n.nodes[node] = NodeInfo{state: StateCandidate, healthy: false}
		if n.isHealthy() {
			n.leader = node
			n.nodes[node] = NodeInfo{state: StateLeader, healthy: true}
		}
	case CommandCrash:
		node := parts[1]
		if n.leader == node {
			n.leader = ""
		}
		// delete(n.nodes, node)
		nn := n.nodes[node]
		nn.healthy = false
		n.nodes[node] = nn
	case CommandRecover:
		node := parts[1]
		n.nodes[node] = NodeInfo{state: StateFollower, healthy: true}
	case CommandWrite:
		if n.leader == "" {
			return "NO_LEADER"
		}
		if !n.isHealthy() {
			return "NO_QUORUM"
		}
		key, val := parts[1], parts[2]
		n.kv.Set(key, val)
		return "OK"
	case CommandRead:
		// if n.leader == "" {
		// 	return "NO_LEADER"
		// }
		// if !n.isHealthy() {
		// 	return "NO_QUORUM"
		// }
		key := parts[1]
		return n.kv.Get(key)
	case CommandState:
		return fmt.Sprintf("term=%d leader=%s kv=%s", n.term, n.leader, n.kv)
	default:
		return ""
	}
	return ""
}

func (n *Node) isHealthy() bool {
	num := 0
	for _, v := range n.nodes {
		if v.healthy {
			num++
		}
	}
	debug("num=%d, quorum=%d", num, len(n.nodes)/2+1)
	return num >= len(n.nodes)/2+1
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
		debug("line=%s, node=%+v, kv=%s", line, node, node.kv)
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
