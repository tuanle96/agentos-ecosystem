package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tuanle96/agentos-ecosystem/core/agent-engine/orchestrator"
	"github.com/tuanle96/agentos-ecosystem/core/agent-engine/scheduler"
	"github.com/tuanle96/agentos-ecosystem/core/agent-engine/communication"
)

func main() {
	log.Println("Starting AgentOS Agent Engine...")
	
	// Initialize communication layer
	comm := communication.New()
	if err := comm.Start(); err != nil {
		log.Fatalf("Failed to start communication layer: %v", err)
	}
	defer comm.Stop()
	
	// Initialize scheduler
	sched := scheduler.New(comm)
	if err := sched.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	defer sched.Stop()
	
	// Initialize orchestrator
	orch := orchestrator.New(sched, comm)
	if err := orch.Start(); err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}
	defer orch.Stop()
	
	log.Println("AgentOS Agent Engine started successfully")
	
	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	
	log.Println("Shutting down AgentOS Agent Engine...")
}
