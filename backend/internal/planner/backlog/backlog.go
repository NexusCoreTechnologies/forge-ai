package backlog

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strings"
)

// Task represents a backlog task with completion state.
type Task struct {
    Text string
    Done bool
}

// Sprint represents a sprint with tasks and parent epic name.
type Sprint struct {
    Name  string
    Epic  string
    Tasks []Task
}

// ReadBacklog reads MASTER_BACKLOG.md and returns lines.
func ReadBacklog(path string) ([]string, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("failed to open backlog: %w", err)
    }
    defer f.Close()
    var lines []string
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}

var sprintHeading = regexp.MustCompile(`(?i)^\s*sprint\b`)
var epicHeading = regexp.MustCompile(`(?i)^\s*#`) // simple: lines starting with '#'
var taskLine = regexp.MustCompile(`^\s*[-*+]\s*(?:\[[ xX]\]\s*)?(.*)$`)
var checkbox = regexp.MustCompile(`\[([ xX])\]`)

// ParseBacklog returns parsed sprints in document order.
func ParseBacklog(lines []string) ([]Sprint, error) {
    var sprints []Sprint
    currentEpic := ""
    var currentSprint *Sprint

    for _, raw := range lines {
        l := strings.TrimSpace(raw)
        if l == "" {
            continue
        }
        if epicHeading.MatchString(l) {
            // treat this as an epic heading
            // strip leading hashes and whitespace
            name := strings.TrimSpace(strings.TrimLeft(l, "# "))
            if name != "" {
                currentEpic = name
            }
            continue
        }
        if sprintHeading.MatchString(l) {
            // start a new sprint
            if currentSprint != nil {
                sprints = append(sprints, *currentSprint)
            }
            currentSprint = &Sprint{Name: l, Epic: currentEpic, Tasks: []Task{}}
            continue
        }
        // task lines (list items)
        if taskLine.MatchString(raw) {
            matches := taskLine.FindStringSubmatch(raw)
            if len(matches) >= 2 && currentSprint != nil {
                text := strings.TrimSpace(matches[1])
                done := false
                cb := checkbox.FindStringSubmatch(raw)
                if len(cb) == 2 && (cb[1] == "x" || cb[1] == "X") {
                    done = true
                }
                currentSprint.Tasks = append(currentSprint.Tasks, Task{Text: text, Done: done})
            }
            continue
        }
        // fallback: non-heading non-empty lines under a sprint treated as tasks
        if currentSprint != nil && !sprintHeading.MatchString(l) && !epicHeading.MatchString(l) {
            // avoid capturing lines that look like metadata; treat as task
            text := strings.TrimSpace(l)
            if text != "" {
                currentSprint.Tasks = append(currentSprint.Tasks, Task{Text: text, Done: false})
            }
            continue
        }
        // non-matching lines ignored
    }
    if currentSprint != nil {
        sprints = append(sprints, *currentSprint)
    }
    if len(sprints) == 0 {
        return nil, fmt.Errorf("no sprints found in backlog")
    }

    // Ensure each sprint has at least one task; if none, add default "Implement <Sprint Name>" task.
    for i, s := range sprints {
        if len(s.Tasks) == 0 {
            name := sprintDisplayName(s.Name)
            defaultTask := Task{Text: "Implement " + name, Done: false}
            sprints[i].Tasks = append(sprints[i].Tasks, defaultTask)
        }
    }

    return sprints, nil
}

// sprintDisplayName extracts the human-friendly sprint name after a hyphen, e.g.
// "Sprint 001 - Initial scaffold" -> "Initial scaffold". Falls back to full name.
func sprintDisplayName(raw string) string {
    if idx := strings.Index(raw, "-"); idx != -1 && idx+1 < len(raw) {
        return strings.TrimSpace(raw[idx+1:])
    }
    // remove leading 'Sprint' word if present
    parts := strings.Fields(raw)
    if len(parts) > 1 && strings.ToLower(parts[0]) == "sprint" {
        return strings.TrimSpace(strings.Join(parts[1:], " "))
    }
    return strings.TrimSpace(raw)
}

// PartitionSprints separates completed and pending sprints based on task completion.
func PartitionSprints(sprints []Sprint) (completed []Sprint, pending []Sprint) {
    for _, s := range sprints {
        doneAll := true
        if len(s.Tasks) == 0 {
            doneAll = false
        }
        for _, t := range s.Tasks {
            if !t.Done {
                doneAll = false
                break
            }
        }
        if doneAll {
            completed = append(completed, s)
        } else {
            pending = append(pending, s)
        }
    }
    return completed, pending
}
