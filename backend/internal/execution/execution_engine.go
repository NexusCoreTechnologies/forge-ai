package execution

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/google/uuid"
    "forgeai/backend/internal/config"
    "forgeai/backend/internal/filesystem"
    "forgeai/backend/internal/logger"
    "forgeai/backend/internal/models"
    "forgeai/backend/internal/prompt"
    "forgeai/backend/internal/providers/factory"
    providersModels "forgeai/backend/internal/providers/models"
    "forgeai/backend/internal/reports"
    "forgeai/backend/internal/workspace"
)

// ExecutionEngine runs the pipeline without AI providers.
type ExecutionEngine struct {
    config           *config.Config
    logger           *logger.JSONLogger
    workspaceManager *workspace.WorkspaceManager
    promptLoader     *prompt.PromptLoader
    reportManager    *reports.ReportManager
}

func NewExecutionEngine(cfg *config.Config, lg *logger.JSONLogger, wm *workspace.WorkspaceManager, pl *prompt.PromptLoader, rm *reports.ReportManager) *ExecutionEngine {
    return &ExecutionEngine{cfg, lg, wm, pl, rm}
}

func (e *ExecutionEngine) Execute(project string, promptName string) (*models.ExecutionReport, error) {
    start := time.Now()
    id := uuid.New().String()
    logs := []string{}
    warnings := []string{}
    errors := []string{}

    e.logger.Info("execution.start", map[string]any{"executionId": id, "project": project, "prompt": promptName})
    logs = append(logs, "execution started")

    if err := e.workspaceManager.OpenWorkspace(e.config.Workspace); err != nil {
        errors = append(errors, err.Error())
        return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errors), err
    }

    if err := e.workspaceManager.ValidateWorkspace(); err != nil {
        errors = append(errors, err.Error())
        return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errors), err
    }
    logs = append(logs, "workspace validated")

    prompt, err := e.promptLoader.ReadPrompt(promptName)
    if err != nil {
        errors = append(errors, err.Error())
        return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errors), err
    }
    logs = append(logs, "prompt loaded")

    if err := e.promptLoader.ValidatePrompt(prompt); err != nil {
        errors = append(errors, err.Error())
        return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errors), err
    }
    logs = append(logs, "prompt validated")

    generatedDir := filepath.Join(e.config.Workspace, "generated")
    if err := filesystem.Create(generatedDir); err != nil {
        errors = append(errors, err.Error())
        return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errors), err
    }

    // Instantiate provider (environment overrides configuration)
    providerName := os.Getenv("FORGEAI_PROVIDER")
    if providerName == "" {
        providerName = "openai"
    }
    prov, err := factory.NewProvider(providerName, &providersModels.Config{APIKeyEnv: "OPENAI_API_KEY"})
    if err != nil {
        errors = append(errors, err.Error())
        return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errors), err
    }

    resp, err := prov.Generate(prompt.Content)
    if err != nil {
        errors = append(errors, err.Error())
        return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errors), err
    }

    outputPath := filepath.Join(generatedDir, fmt.Sprintf("%s.execution.md", prompt.Name))
    outputContent := fmt.Sprintf(`# Execution result for %s

Project: %s

Prompt content:

%s

Response:

%s
`, prompt.Name, project, prompt.Content, resp.Text)
    if err := filesystem.Write(outputPath, []byte(outputContent)); err != nil {
        errors = append(errors, err.Error())
        return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errors), err
    }
    logs = append(logs, "generated execution artifact")

    report := e.finishReport(id, project, promptName, start, "completed", logs, warnings, errors)
    if err := e.reportManager.SaveExecutionReport(report); err != nil {
        errors = append(errors, err.Error())
        report.Status = "failed"
        report.Errors = errors
        e.logger.Error("report save failed", map[string]any{"error": err.Error()})
        return report, err
    }
    logs = append(logs, "report saved")
    e.logger.Info("execution.completed", map[string]any{"executionId": id, "status": report.Status})

    return report, nil
}

func (e *ExecutionEngine) finishReport(id, project, promptName string, start time.Time, status string, logs, warnings, errors []string) *models.ExecutionReport {
    end := time.Now()
    return &models.ExecutionReport{
        ExecutionID:    id,
        Project:        project,
        Prompt:         promptName,
        StartTime:      start,
        EndTime:        end,
        DurationSeconds: end.Sub(start).Seconds(),
        Status:         status,
        Errors:         errors,
        Warnings:       warnings,
        Logs:           logs,
    }
}
