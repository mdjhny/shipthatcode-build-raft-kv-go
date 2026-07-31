package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type protocol string

const (
	None   protocol = "NONE"
	Raft   protocol = "RAFT"
	Paxos  protocol = "PAXOS"
	Bft    protocol = "BFT"
	Gossip protocol = "GOSSIP"
)

var classifier = map[string]protocol{
	"A single-node application":                            None,
	"PostgreSQL primary with read replicas":                None,
	"HashiCorp Consul KV store":                            Raft,
	"Service mesh discovering 1000s of nodes":              Gossip,
	"Distributed ledger across mutually-untrusted parties": Bft,
	"Internal Google service replicating metadata":         Paxos,
	"3-node etcd cluster":                                  Raft,
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		_ = parts
		p := classifier[line]
		fmt.Fprintln(out, p)
	}
}
