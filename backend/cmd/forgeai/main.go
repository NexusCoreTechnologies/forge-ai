package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"forgeai/backend/internal/config"
	builder "forgeai/backend/internal/context/builder"
	collector "forgeai/backend/internal/context/collector"
	parser "forgeai/backend/internal/context/parser"
	scanner "forgeai/backend/internal/context/scanner"
	"forgeai/backend/internal/execution"
	"forgeai/backend/internal/logger"
	"forgeai/backend/internal/orchestrator"
	backlog "forgeai/backend/internal/planner/backlog"
	executionplan "forgeai/backend/internal/planner/executionplan"
	"forgeai/backend/internal/prompt"
	_ "forgeai/backend/internal/providers/mock"
	_ "forgeai/backend/internal/providers/openai"
	"forgeai/backend/internal/reports"
	"forgeai/backend/internal/workspace"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("expected subcommand: execute, build-context, plan-sprint, or run")
	}
	switch os.Args[1] {
	case "execute":
		executeCmd := flag.NewFlagSet("execute", flag.ExitOnError)
		project := executeCmd.String("project", "Forge AI", "Project name")
		promptName := executeCmd.String("prompt", "00-bootstrap.md", "Prompt file name")
		configPath := executeCmd.String("config", "config/config.yaml", "Path to config file")
		executeCmd.Parse(os.Args[2:])

		cfg, err := config.Load(*configPath)
		if err != nil {
			log.Fatal(err)
		}
		lg := logger.NewJSONLogger(os.Stdout)
		if err := runExecution(cfg, lg, *project, *promptName); err != nil {
			os.Exit(1)
		}
	case "build-context":
		buildContextCmd()
	case "plan-sprint":
		planSprintCmd()
	case "run":
		runCmd()
	default:
		log.Fatalf("unknown subcommand: %s", os.Args[1])
	}
}

// planSprintCmd generates a simple execution plan from MASTER_BACKLOG.md
func planSprintCmd() {
	planCmd := flag.NewFlagSet("plan-sprint", flag.ExitOnError)
	configPath := planCmd.String("config", "config/config.yaml", "Path to config file")
	planCmd.Parse(os.Args[2:])

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	lg := logger.NewJSONLogger(os.Stdout)
	backlogPath := filepath.Join(cfg.Workspace, "MASTER_BACKLOG.md")
	lines, err := backlog.ReadBacklog(backlogPath)
	if err != nil {
		lg.Error("backlog.read.failed", map[string]any{"error": err.Error()})
		fmt.Fprintln(os.Stderr, "error: could not read MASTER_BACKLOG.md - ensure the file exists in the workspace")
		os.Exit(1)
	}
	sprints, err := backlog.ParseBacklog(lines)
	if err != nil {
		lg.Error("backlog.parse.failed", map[string]any{"error": err.Error()})
		fmt.Fprintln(os.Stderr, "error: could not parse backlog:", err.Error())
		os.Exit(1)
	}
	completed, pending := backlog.PartitionSprints(sprints)
	if len(pending) == 0 {
		lg.Info("planner.nothing_to_do", map[string]any{"completed": len(completed)})
		fmt.Fprintln(os.Stderr, "no pending sprints found in MASTER_BACKLOG.md")
		os.Exit(0)
	}
	selectedSprint := pending[0]
	// Collect task texts for execution plan
	tasks := []string{}
	for _, t := range selectedSprint.Tasks {
		tasks = append(tasks, t.Text)
	}
	plan := &executionplan.Plan{SelectedSprint: selectedSprint.Name, Tasks: tasks}
	if err := executionplan.WritePlan(filepath.Join(cfg.Workspace, "execution_plan.json"), plan); err != nil {
		lg.Error("plan.write.failed", map[string]any{"error": err.Error()})
		fmt.Fprintln(os.Stderr, "error: failed to write execution_plan.json:", err.Error())
		os.Exit(1)
	}
	lg.Info("plan.generated", map[string]any{"sprint": selectedSprint.Name, "tasks": len(tasks)})
}

// buildContextCmd builds an execution context for a workspace.
func buildContextCmd() {
	buildCmd := flag.NewFlagSet("build-context", flag.ExitOnError)
	project := buildCmd.String("project", "Forge AI", "Project name")
	configPath := buildCmd.String("config", "config/config.yaml", "Path to config file")
	buildCmd.Parse(os.Args[2:])

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	lg := logger.NewJSONLogger(os.Stdout)

	if err := buildExecutionContextFile(cfg, *project, lg); err != nil {
		lg.Error("context.build.failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
}

func runCmd() {
	runCmdFlags := flag.NewFlagSet("run", flag.ExitOnError)
	project := runCmdFlags.String("project", "Forge AI", "Project name")
	promptName := runCmdFlags.String("prompt", "00-bootstrap.md", "Prompt file name")
	configPath := runCmdFlags.String("config", "config/config.yaml", "Path to config file")
	runCmdFlags.Parse(os.Args[2:])

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	lg := logger.NewJSONLogger(os.Stdout)

	wm := workspace.NewWorkspaceManager(cfg, lg)
	if err := wm.OpenWorkspace(cfg.Workspace); err != nil {
		lg.Error("workspace.open.failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	if err := wm.ValidateWorkspace(); err != nil {
		lg.Error("workspace.validate.failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}

	if err := buildExecutionContextFile(cfg, *project, lg); err != nil {
		lg.Error("context.build.failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	if err := generateExecutionPlanFile(cfg, lg); err != nil {
		lg.Error("plan.generate.failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}

	if err := runExecution(cfg, lg, *project, *promptName); err != nil {
		printRunSummary(false, *project, cfg.Workspace, "failed")
		os.Exit(1)
	}

	printRunSummary(true, *project, cfg.Workspace, "success")
}

func buildExecutionContextFile(cfg *config.Config, project string, lg *logger.JSONLogger) error {
	ws := cfg.Workspace
	ignored := []string{".git", "node_modules"}
	files, err := scanner.ScanWorkspace(ws, ignored)
	if err != nil {
		return err
	}
	collected, err := collector.CollectFiles(files)
	if err != nil {
		return err
	}
	docs := map[string]string{}
	for p, b := range collected {
		parsed, _ := parser.ParseText(b)
		docs[p] = parsed
	}
	ctx := builder.BuildExecutionContext(project, ws, "", "", "", "", docs, files, nil, nil, nil)
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(ws, "execution_context.json"), data, 0o644); err != nil {
		return err
	}
	lg.Info("context.built", map[string]any{"files": len(files)})
	return nil
}

func generateExecutionPlanFile(cfg *config.Config, lg *logger.JSONLogger) error {
	backlogPath := filepath.Join(cfg.Workspace, "MASTER_BACKLOG.md")
	lines, err := backlog.ReadBacklog(backlogPath)
	if err != nil {
		return err
	}
	sprints, err := backlog.ParseBacklog(lines)
	if err != nil {
		return err
	}
	completed, pending := backlog.PartitionSprints(sprints)
	if len(pending) == 0 {
		lg.Info("planner.nothing_to_do", map[string]any{"completed": len(completed)})
		return nil
	}
	selectedSprint := pending[0]
	tasks := []string{}
	for _, t := range selectedSprint.Tasks {
		tasks = append(tasks, t.Text)
	}
	plan := &executionplan.Plan{SelectedSprint: selectedSprint.Name, Tasks: tasks}
	if err := executionplan.WritePlan(filepath.Join(cfg.Workspace, "execution_plan.json"), plan); err != nil {
		return err
	}
	lg.Info("plan.generated", map[string]any{"sprint": selectedSprint.Name, "tasks": len(tasks)})
	return nil
}

func printRunSummary(success bool, project, workspace, status string) {
	const (
		reset = "\033[0m"
		cyan  = "\033[36m"
		green = "\033[32m"
		red   = "\033[31m"
		bold  = "\033[1m"
	)
	statusColor := green
	if !success {
		statusColor = red
	}

	fmt.Println("\n========================================")
	fmt.Printf("%sForge AI%s\n", bold+cyan, reset)
	fmt.Printf("Workspace ............ %sOK%s\n", green, reset)
	fmt.Printf("Context .............. %sOK%s\n", green, reset)
	fmt.Printf("Planner .............. %sOK%s\n", green, reset)
	fmt.Printf("Provider ............. %sOpenAI%s\n", cyan, reset)
	fmt.Printf("Model ................ %sGPT%s\n", cyan, reset)
	fmt.Printf("Response ............. %sSaved%s\n", green, reset)
	fmt.Printf("Execution ............ %s%s%s\n", statusColor, status, reset)
	fmt.Println("========================================")
	_ = project
	_ = workspace
}

func runExecution(cfg *config.Config, lg *logger.JSONLogger, project, promptName string) error {
	wm := workspace.NewWorkspaceManager(cfg, lg)
	pl := prompt.NewPromptLoader(cfg)
	rm := reports.NewReportManager(cfg.Workspace)
	engine := execution.NewExecutionEngine(cfg, lg, wm, pl, rm)
	orchestrator := orchestrator.NewOrchestrator(engine)

	report, err := orchestrator.Execute(project, promptName)
	if err != nil {
		lg.Error("execution failed", map[string]any{"error": err.Error()})
	}

	data, marshalErr := json.MarshalIndent(report, "", "  ")
	if marshalErr != nil {
		return marshalErr
	}
	fmt.Println(string(data))
	return err
}
