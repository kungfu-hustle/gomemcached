package main

import (
	"flag"
	"log"
	//"time"

	"github.com/couchbase/gomemcached/client"
)

var prot = flag.String("prot", "tcp", "Layer 3 protocol (tcp, tcp4, tcp6)")
var dest = flag.String("dest", "localhost:12000", "Host:port to connect to")
var u = flag.String("user", "", "SASL plain username")
var p = flag.String("pass", "", "SASL plain password")

func main() {
	flag.Parse()
	log.Printf("Connecting to %s/%s", *prot, *dest)

	client, err := memcached.Connect(*prot, *dest)
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	if *u != "" {
		resp, err := client.Auth(*u, *p)
		if err != nil {
			log.Fatalf("auth error: %v", err)
		}
		log.Printf("Auth response = %v", resp)
	}

	// get failover logs for some vbuckets
	vbuckets := []uint16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	failovermap, err := client.UprGetFailoverLog(vbuckets)
	if err != nil {
		log.Fatalf("Failed to get failover log %v", err)
	}

	for vb, flog := range failovermap {
		log.Printf("Failover log for vb %d is %v", vb, flog)
	}

	uf, err := client.NewUprFeed()
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	err = uf.UprOpen("example", 0, 400)
	if err != nil {
		log.Fatalf("Error in UPR Open: %v", err)
	}

	//time.Sleep(10 * time.Second)

	for i := 0; i < 64; i++ {
		if err := uf.UprRequestStream(uint16(i), 0, 0, 0, 0xFFFFFFFFFFFFFFFF, 0, 0); err != nil {
			log.Fatalf("Request stream for vb %d Failed %v", i, err)
		}
	}

	if err := uf.UprRequestStream(uint16(100), 0, 0, 0, 0, 0, 0); err != nil {
		log.Fatalf("Request stream for vb 100 Failed %v", err)
	}

	err = uf.StartFeed()
	if err != nil {
		log.Fatalf("Error starting upr feed: %v", err)
	}
	for op := range uf.C {
		if op.String() == "SnapshotMarker" {
			log.Printf("Received Snapshot marker for Vbucket %d. Start Sequence %d End Sequence %d", op.VBucket, op.SnapstartSeq, op.SnapendSeq)
		} else if op.String() == "Mutation" {
			log.Printf("Received %s Key %s, Sequence %d, Cas %d\n", op.String(), op.Key, op.SeqNo, op.Cas)
			if len(op.Value) > 0 && len(op.Value) < 500 {
				log.Printf("\tValue: %s", op.Value)
			}
		}

		if op.Status != 0 {
			log.Printf("Got an Error for vbucket %d, Error %s", op.VBucket, op.Error.Error())
		}
	}
	log.Printf("Upr feed closed; err = %v.", uf.Error)
}

func failoverLog(vb uint16, flog memcached.FailoverLog, err error) {
	log.Printf("Failover log for vbucket %d, %v", vb, flog)
}
