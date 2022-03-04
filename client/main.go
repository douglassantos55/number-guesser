package main

import "example.com/game/client/client"

func main() {
    c := client.NewClient()
    c.Connect("0.0.0.0:8080")

    c.Loop()
}
