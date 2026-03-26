package main

import (
	"fmt"
	"log"

	"noc-mcp/internal/server"
)

func main() {
	log.Println("Iniciando Agente MCP para NOC...")

	if err := server.SetupAndRun(); err != nil {
		fmt.Printf("Error fatal iniciando el servidor: %v\n", err)
	}
}
