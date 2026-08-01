package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"

    "forgeai/backend/internal/config"
    "forgeai/backend/internal/execution"
    "forgeai/backend/internal/logger"
    "forgeai/backend/internal/orchestrator"
    "forgeai/backend/internal/prompt"
    "forgeai/backend/internal/reports"
    "forgeai/backend/internal/workspace"
    scanner "forgeai/backend/internal/context/scanner"
    collector "forgeai/backend/internal/context/collector"
    parser "forgeai/backend/internal/context/parser"
    builder "forgeai/backend/internal/context/builder"
    backlog "forgeai/backend/internal/planner/backlog"
    executionplan "forgeai/backend/internal/planner/executionplan"
    _ "forgeai/backend/internal/providers/openai"
    _ "forgeai/backend/internal/providers/mock"
)

func main() {
    if len(os.Args) < 2 {
        log.Fatal("expected subcommand: execute")
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

    // Build context using context packages (scanner, collector, parser, builder)
    ws := cfg.Workspace
    ignored := []string{".git", "node_modules"}
    files, _ := scanner.ScanWorkspace(ws, ignored)
    collected, _ := collector.CollectFiles(files)
    docs := map[string]string{}
    for p, b := range collected {
        // naive parsing based on extension
        parsed, _ := parser.ParseText(b)
        docs[p] = parsed
    }
    ctx := builder.BuildExecutionContext(*project, ws, "", "", "", "", docs, files, nil, nil, nil)
    // save to workspace
    data, _ := json.MarshalIndent(ctx, "", "  ")
    _ = os.WriteFile(filepath.Join(ws, "execution_context.json"), data, 0o644)
    lg.Info("context.built", map[string]any{"files": len(files)})
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
