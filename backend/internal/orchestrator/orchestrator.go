package orchestrator

import (
    "forgeai/backend/internal/execution"
    "forgeai/backend/internal/models"
)

// Orchestrator coordinates execution engine operations.
type Orchestrator struct {
    engine *execution.ExecutionEngine
}

func NewOrchestrator(engine *execution.ExecutionEngine) *Orchestrator {
    return &Orchestrator{engine: engine}
}

func (o *Orchestrator) Execute(project string, promptName string) (*models.ExecutionReport, error) {
    return o.engine.Execute(project, promptName)
}
