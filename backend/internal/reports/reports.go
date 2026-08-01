package reports

import (
    "encoding/json"
    "fmt"
    "path/filepath"

    "forgeai/backend/internal/models"
    "forgeai/backend/internal/filesystem"
)

// ReportManager writes execution reports to disk.
type ReportManager struct {
    workspaceRoot string
}

func NewReportManager(workspaceRoot string) *ReportManager {
    return &ReportManager{workspaceRoot: workspaceRoot}
}

func (r *ReportManager) SaveExecutionReport(report *models.ExecutionReport) error {
    reportDir := filepath.Join(r.workspaceRoot, "reports")
    if err := filesystem.Create(reportDir); err != nil {
        return err
    }
    path := filepath.Join(reportDir, fmt.Sprintf("execution_report_%s.json", report.ExecutionID))
    data, err := json.MarshalIndent(report, "", "  ")
    if err != nil {
        return err
    }
    return filesystem.Write(path, data)
}
