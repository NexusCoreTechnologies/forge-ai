package execution

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"forgeai/backend/internal/config"
	contextmodels "forgeai/backend/internal/context/models"
	"forgeai/backend/internal/context/optimizer"
	"forgeai/backend/internal/filesystem"
	"forgeai/backend/internal/logger"
	"forgeai/backend/internal/models"
	"forgeai/backend/internal/prompt"
	"forgeai/backend/internal/providers/factory"
	providersModels "forgeai/backend/internal/providers/models"
	openaiProvider "forgeai/backend/internal/providers/openai"
	"forgeai/backend/internal/reports"
	"forgeai/backend/internal/workspace"
	"github.com/google/uuid"
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
	errorList := []string{}

	e.logger.Info("execution.start", map[string]any{"executionId": id, "project": project, "prompt": promptName})
	logs = append(logs, "execution started")

	if err := e.workspaceManager.OpenWorkspace(e.config.Workspace); err != nil {
		errorList = append(errorList, err.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), err
	}

	if err := e.workspaceManager.ValidateWorkspace(); err != nil {
		errorList = append(errorList, err.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), err
	}
	logs = append(logs, "workspace validated")

	prompt, err := e.promptLoader.ReadPrompt(promptName)
	if err != nil {
		errorList = append(errorList, err.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), err
	}
	logs = append(logs, "prompt loaded")

	if err := e.promptLoader.ValidatePrompt(prompt); err != nil {
		errorList = append(errorList, err.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), err
	}
	logs = append(logs, "prompt validated")

	generatedDir := filepath.Join(e.config.Workspace, "generated")
	if err := filesystem.Create(generatedDir); err != nil {
		errorList = append(errorList, err.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), err
	}

	// Instantiate provider (environment overrides configuration)
	providerName := os.Getenv("FORGEAI_PROVIDER")
	if providerName == "" {
		providerName = "openai"
	}
	prov, err := factory.NewProvider(providerName, &providersModels.Config{APIKeyEnv: "OPENAI_API_KEY"})
	if err != nil {
		errorList = append(errorList, err.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), err
	}

	// Validate provider configuration (API key presence etc.)
	if verr := prov.ValidateConfiguration(); verr != nil {
		errorList = append(errorList, verr.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), verr
	}

	// Optimize the context before sending it to the provider.
	contextPath := filepath.Join(e.config.Workspace, "execution_context.json")
	planPath := filepath.Join(e.config.Workspace, "execution_plan.json")
	var execContext *contextmodels.ExecutionContext
	if b, err := os.ReadFile(contextPath); err == nil {
		if unmarshalErr := json.Unmarshal(b, &execContext); unmarshalErr != nil {
			e.logger.Info("context.unmarshal.failed", map[string]any{"error": unmarshalErr.Error(), "path": contextPath})
		}
	}

	planContent := "(no execution_plan.json)"
	if b, err := os.ReadFile(planPath); err == nil {
		planContent = string(b)
	}

	optimizedContext, metrics, err := optimizer.Optimize(e.config.Workspace, prompt.Content, planContent, execContext)
	if err != nil {
		errorList = append(errorList, err.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), err
	}
	optData, marshalErr := json.MarshalIndent(optimizedContext, "", "  ")
	if marshalErr != nil {
		errorList = append(errorList, marshalErr.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), marshalErr
	}
	if err := os.WriteFile(filepath.Join(e.config.Workspace, "optimized_context.json"), optData, 0o644); err != nil {
		errorList = append(errorList, err.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), err
	}
	e.logger.Info("context.optimized", map[string]any{"executionId": id, "original_tokens": metrics.OriginalTokens, "optimized_tokens": metrics.OptimizedTokens, "compression_ratio": metrics.CompressionRatio})

	compositePrompt := fmt.Sprintf("Optimized Context:\n%s\n\nPrompt:\n%s", string(optData), prompt.Content)

	provInfo := prov.ProviderInfo()
	e.logger.Info("provider.invoke.start", map[string]any{"provider": provInfo.Name, "model": provInfo.Model, "executionId": id})
	providerStart := time.Now()
	resp, err := prov.Generate(compositePrompt)
	providerDuration := time.Since(providerStart)
	if err != nil {
		result := map[string]any{
			"provider":               provInfo.Name,
			"model":                  provInfo.Model,
			"prompt":                 prompt.Name,
			"tokens":                 0,
			"execution_time_seconds": providerDuration.Seconds(),
			"status":                 "failed",
			"error_code":             "",
			"retryable":              false,
		}
		var quotaErr *openaiProvider.QuotaError
		if errors.As(err, &quotaErr) {
			result["error_code"] = quotaErr.Code
			result["retryable"] = false
		}
		if rb, marshalErr := json.MarshalIndent(result, "", "  "); marshalErr == nil {
			_ = os.WriteFile(filepath.Join(generatedDir, fmt.Sprintf("execution_result_%s.json", id)), rb, 0o644)
		}
		errorList = append(errorList, err.Error())
		e.logger.Error("provider.invoke.failed", map[string]any{"error": err.Error(), "provider": provInfo.Name, "executionId": id})
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), err
	}
	e.logger.Info("provider.invoke.completed", map[string]any{"provider": provInfo.Name, "model": provInfo.Model, "executionId": id, "tokens": resp.Tokens, "latency_ms": providerDuration.Milliseconds()})

	// Save response to generated/response.md
	responsePath := filepath.Join(generatedDir, "response.md")
	if err := filesystem.Write(responsePath, []byte(resp.Text)); err != nil {
		errorList = append(errorList, err.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), err
	}

	// Write execution_result.json into generated
	result := map[string]any{
		"provider":               provInfo.Name,
		"model":                  provInfo.Model,
		"prompt":                 prompt.Name,
		"tokens":                 resp.Tokens,
		"execution_time_seconds": providerDuration.Seconds(),
		"status":                 "completed",
	}
	// cost is provider-specific and may not be available
	_ = result
	resultPath := filepath.Join(generatedDir, fmt.Sprintf("execution_result_%s.json", id))
	if rb, err := json.MarshalIndent(result, "", "  "); err == nil {
		_ = os.WriteFile(resultPath, rb, 0o644)
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
		errorList = append(errorList, err.Error())
		return e.finishReport(id, project, promptName, start, "failed", logs, warnings, errorList), err
	}
	logs = append(logs, "generated execution artifact")

	report := e.finishReport(id, project, promptName, start, "completed", logs, warnings, errorList)
	if err := e.reportManager.SaveExecutionReport(report); err != nil {
		errorList = append(errorList, err.Error())
		report.Status = "failed"
		report.Errors = errorList
		e.logger.Error("report save failed", map[string]any{"error": err.Error()})
		return report, err
	}
	logs = append(logs, "report saved")
	e.logger.Info("execution.completed", map[string]any{"executionId": id, "status": report.Status})

	return report, nil
}

func (e *ExecutionEngine) finishReport(id, project, promptName string, start time.Time, status string, logs, warnings, errorMessages []string) *models.ExecutionReport {
	end := time.Now()
	return &models.ExecutionReport{
		ExecutionID:     id,
		Project:         project,
		Prompt:          promptName,
		StartTime:       start,
		EndTime:         end,
		DurationSeconds: end.Sub(start).Seconds(),
		Status:          status,
		Errors:          errorMessages,
		Warnings:        warnings,
		Logs:            logs,
	}
}
