package main

import (
	"SOMAS2023/internal/server"
)

func main() {
	s := server.GenerateServer() // returns a zero-valued server object
	s.Initialize(100) //sets up the environment (parameter is number of iterations, usually = 100)
	s.Start()
}