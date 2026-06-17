package main

import (
  "fmt"
  "os"
  "time"
  "github.com/nats-io/nats.go"
)

func main() {
  url := os.Getenv("NATS_URL")
  if url == "" { url = "nats://127.0.0.1:4222" }
  nc, err := nats.Connect(url)
  if err != nil { panic(err) }
  defer nc.Close()
  sub, err := nc.SubscribeSync("raw_events")
  if err != nil { panic(err) }
  nc.Flush()
  deadline := time.Now().Add(10 * time.Second)
  count := 0
  for time.Now().Before(deadline) {
    msg, err := sub.NextMsg(2 * time.Second)
    if err != nil { continue }
    count++
    fmt.Printf("MSG %d %d\n", count, len(msg.Data))
    if count >= 3 { break }
  }
  fmt.Printf("TOTAL %d\n", count)
}
