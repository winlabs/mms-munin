package main

import (
	"mongodb.com/munin-agent/components"
)

func main() {
	agent := &components.Agent{}
	agent.Run()
}