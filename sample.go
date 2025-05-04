package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "gopkg.in/zeromq/goczmq.v4"
    "log"
    "os"
    "os/signal"
    "strings"
    "time"
)

const (
    defaultProofHeader = "proofblock"
    defaultHash        = "0x0000000000000000000"
)

type Proof struct {
    Type      string `json:"type"`
    Timestamp string `json:"timestamp,omitempty"`
    Data      string `json:"data"`
}

func main() {
    ep := os.Getenv("PROOF_ENDPOINT")
    if ep == "" {
        ep = "tcp://34.71.52.251:40000"
    }

    sample := goczmq.NewReqChanneler(ep)
    if sample == nil {
        log.Fatalf("Failed to subscribe to endpoint: %s", ep)
    }
    defer sample.Destroy()

    // Graceful shutdown on Ctrl+C
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func() {
        <-c
        log.Println("Interrupt received, shutting down.")
        sample.Destroy()
        os.Exit(0)
    }()

    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter proof data (or press Enter to use default): ")
    input, err := reader.ReadString('\n')
    if err != nil {
        log.Fatalf("Failed to read input: %v", err)
    }
    input = strings.TrimSpace(input)

    time.Sleep(5 * time.Second) // Wait for connection

    var proof Proof
    if input == "" {
        proof = Proof{Type: defaultProofHeader, Timestamp: time.Now().String(), Data: defaultHash}
    } else {
        proof = Proof{Type: defaultProofHeader, Data: input}
    }

    jsonProof, err := json.Marshal(proof)
    if err != nil {
        log.Fatalf("Failed to encode proof as JSON: %v", err)
    }

    select {
    case sample.SendChan <- [][]byte{jsonProof}:
        log.Printf("Proof sent: %s", jsonProof)
    case <-time.After(5 * time.Second):
        log.Println("Timed out while trying to send proof.")
        return
    }

    resp := <-sample.RecvChan
    log.Printf("Response received: %s", resp)
}
