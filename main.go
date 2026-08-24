package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Strategy string

const (
	StrategyReadIndex         Strategy = "READ_INDEX"
	StrategyLease             Strategy = "LEASE"
	StrategyFollowerRead      Strategy = "FOLLOWER_READ"
	StrategyLinearizableWrite Strategy = "LINEARIZABLE_WRITE"
)

var strategyMapping = map[string]Strategy{
	"Critical financial query: must see latest committed value": StrategyReadIndex,
	"Cached homepage data, OK to be slightly stale":             StrategyFollowerRead,
	"Bank transfer (debit + credit)":                            StrategyLinearizableWrite,
	"Read-heavy analytics with bounded staleness":               StrategyFollowerRead,
	"Distributed lock query":                                    StrategyReadIndex,
	"Token validation in API gateway":                           StrategyLease,
	"Banking ledger update":                                     StrategyLinearizableWrite,
	"Strict consistency required":                               StrategyReadIndex,
	"Session check on leader lease":                             StrategyLease,
	"Cached metrics dashboard":                                  StrategyFollowerRead,
	"Linearizable read after a prior commit":                    StrategyReadIndex,
	"Mutation on the ledger":                                    StrategyLinearizableWrite,
}

type Node struct {
}

func newNode() *Node {
	return &Node{}
}

func (n *Node) handleLine(line string) Strategy {
	return strategyMapping[line]
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
		result := node.handleLine(line)
		if result != "" {
			fmt.Fprintln(out, result)
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}
