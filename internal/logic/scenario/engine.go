package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ──────────────────────────────────────────
// Template rendering
// ──────────────────────────────────────────

var varPattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

// RenderVars replaces {{variable}} placeholders with values from the variables map.
func RenderVars(template string, vars map[string]string) string {
	return varPattern.ReplaceAllStringFunc(template, func(match string) string {
		key := match[2 : len(match)-2]
		if val, ok := vars[key]; ok {
			return val
		}
		return match
	})
}

// RenderVarsInMap recursively replaces variables in a map's string values.
func RenderVarsInMap(params map[string]interface{}, vars map[string]string) map[string]interface{} {
	result := make(map[string]interface{}, len(params))
	for k, v := range params {
		switch val := v.(type) {
		case string:
			result[k] = RenderVars(val, vars)
		case map[string]interface{}:
			result[k] = RenderVarsInMap(val, vars)
		default:
			result[k] = v
		}
	}
	return result
}

// ──────────────────────────────────────────
// Expression evaluator (enhanced)
// ──────────────────────────────────────────

// EvalCondition evaluates a condition expression with variables.
// Supports:
//   - Comparisons: ==, !=, >, <, >=, <=
//   - String functions: contains(a, b), starts_with(a, b), ends_with(a, b)
//   - Logical: expr && expr, expr || expr
//   - Truthy check (non-empty, non-false, non-0)
func EvalCondition(expr string, vars map[string]string) bool {
	rendered := RenderVars(expr, vars)
	return evalExpression(rendered)
}

// evalExpression handles logical operators && and ||.
func evalExpression(expr string) bool {
	expr = strings.TrimSpace(expr)

	// Handle || (lowest precedence)
	if parts := splitLogical(expr, "||"); len(parts) > 1 {
		for _, part := range parts {
			if evalExpression(part) {
				return true
			}
		}
		return false
	}

	// Handle && (higher precedence)
	if parts := splitLogical(expr, "&&"); len(parts) > 1 {
		for _, part := range parts {
			if !evalExpression(part) {
				return false
			}
		}
		return true
	}

	// Handle negation
	if strings.HasPrefix(expr, "!") {
		return !evalExpression(expr[1:])
	}

	// Handle parentheses
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		return evalExpression(expr[1 : len(expr)-1])
	}

	// Handle string functions
	if r, ok := evalStringFunc(expr); ok {
		return r
	}

	// Handle comparisons
	return evalComparison(expr)
}

// splitLogical splits an expression by a logical operator, respecting parentheses.
func splitLogical(expr, op string) []string {
	depth := 0
	var parts []string
	start := 0
	for i := 0; i < len(expr); i++ {
		switch expr[i] {
		case '(':
			depth++
		case ')':
			depth--
		}
		if depth == 0 && i+len(op) <= len(expr) && expr[i:i+len(op)] == op {
			parts = append(parts, expr[start:i])
			start = i + len(op)
			i += len(op) - 1
		}
	}
	parts = append(parts, expr[start:])
	if len(parts) <= 1 {
		return nil // no split happened
	}
	return parts
}

// evalStringFunc evaluates contains(), starts_with(), ends_with().
func evalStringFunc(expr string) (bool, bool) {
	expr = strings.TrimSpace(expr)
	for _, fn := range []string{"contains", "starts_with", "ends_with"} {
		if strings.HasPrefix(expr, fn+"(") && strings.HasSuffix(expr, ")") {
			inner := expr[len(fn)+1 : len(expr)-1]
			args := splitFuncArgs(inner)
			if len(args) != 2 {
				return false, true
			}
			a := strings.Trim(strings.TrimSpace(args[0]), "'\"")
			b := strings.Trim(strings.TrimSpace(args[1]), "'\"")
			switch fn {
			case "contains":
				return strings.Contains(a, b), true
			case "starts_with":
				return strings.HasPrefix(a, b), true
			case "ends_with":
				return strings.HasSuffix(a, b), true
			}
		}
	}
	return false, false
}

// splitFuncArgs splits function arguments by comma, respecting nested parentheses and quotes.
func splitFuncArgs(s string) []string {
	var args []string
	depth := 0
	inQuote := byte(0)
	start := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inQuote != 0 {
			if ch == inQuote {
				inQuote = 0
			}
			continue
		}
		switch ch {
		case '\'', '"':
			inQuote = ch
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				args = append(args, s[start:i])
				start = i + 1
			}
		}
	}
	args = append(args, s[start:])
	return args
}

// evalComparison handles simple comparisons: ==, !=, >, <, >=, <=.
func evalComparison(expr string) bool {
	operators := []string{"!=", ">=", "<=", "==", ">", "<"}
	for _, op := range operators {
		parts := strings.SplitN(expr, op, 2)
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			left = strings.Trim(left, "'\"")
			right = strings.Trim(right, "'\"")

			switch op {
			case "==":
				return left == right
			case "!=":
				return left != right
			case ">", "<", ">=", "<=":
				return compareNumericOrString(left, right, op)
			}
		}
	}

	// If no operator found, treat as truthy check
	expr = strings.TrimSpace(expr)
	return expr != "" && expr != "false" && expr != "0" && expr != "null"
}

// compareNumericOrString tries numeric comparison first, falls back to string comparison.
func compareNumericOrString(left, right, op string) bool {
	lf, errL := strconv.ParseFloat(left, 64)
	rf, errR := strconv.ParseFloat(right, 64)
	if errL == nil && errR == nil {
		switch op {
		case ">":
			return lf > rf
		case "<":
			return lf < rf
		case ">=":
			return lf >= rf
		case "<=":
			return lf <= rf
		}
	}
	// Fallback to lexicographic
	switch op {
	case ">":
		return left > right
	case "<":
		return left < right
	case ">=":
		return left >= right
	case "<=":
		return left <= right
	}
	return false
}

// ──────────────────────────────────────────
// Engine — execute scenario
// ──────────────────────────────────────────

// ActionExecutor is the interface for step executors.
type ActionExecutor interface {
	Execute(ctx context.Context, execCtx *ExecutionContext, step *StepDSL) (output map[string]interface{}, err error)
}

// ActionRegistry maps step types to their executors.
var ActionRegistry = map[string]ActionExecutor{}

// RegisterAction registers an action executor for a step type.
func RegisterAction(stepType string, executor ActionExecutor) {
	ActionRegistry[stepType] = executor
}

// RunScenario is the main entry point to start executing a scenario.
func RunScenario(ctx context.Context, merchantId uint64, scenarioRow interface{ GetId() uint64 }, dsl *ScenarioDSL, triggerData map[string]interface{}) {
	// This is a wrapper for RunScenarioByIds
	// Not used directly — see RunScenarioByIds
}

// RunScenarioByIds starts a scenario execution given IDs.
func RunScenarioByIds(ctx context.Context, merchantId, scenarioId uint64, scenarioJson string, triggerData map[string]interface{}) {
	dsl, err := ParseDSL(scenarioJson)
	if err != nil {
		g.Log().Errorf(ctx, "scenario: failed to parse DSL for scenario %d: %v", scenarioId, err)
		return
	}

	// Initialize variables from trigger data
	vars := make(map[string]string)
	if dsl.Variables != nil {
		for k, v := range dsl.Variables {
			vars[k] = RenderVars(v, flattenTriggerData(triggerData))
		}
	}
	// Also add raw trigger data as variables
	for k, v := range flattenTriggerData(triggerData) {
		if _, exists := vars[k]; !exists {
			vars[k] = v
		}
	}

	// Create execution record
	execId, err := CreateExecution(ctx, merchantId, scenarioId, triggerData, vars)
	if err != nil {
		g.Log().Errorf(ctx, "scenario: failed to create execution for scenario %d: %v", scenarioId, err)
		return
	}

	execCtx := &ExecutionContext{
		ExecutionID: execId,
		MerchantID:  merchantId,
		ScenarioID:  scenarioId,
		Variables:   vars,
		TriggerData: triggerData,
	}

	// Execute steps sequentially
	executeSteps(ctx, execCtx, dsl.Steps, 0)
}

// executeSteps runs steps starting from the given index.
func executeSteps(ctx context.Context, execCtx *ExecutionContext, steps []StepDSL, startIdx int) {
	for i := startIdx; i < len(steps); i++ {
		step := &steps[i]
		startTime := time.Now()

		// Update current step
		_ = UpdateExecutionStatus(ctx, execCtx.ExecutionID, StatusRunning, step.ID, "", execCtx.Variables)

		// Resolve params with variables
		resolvedParams := RenderVarsInMap(step.Params, execCtx.Variables)
		resolvedStep := &StepDSL{
			ID:     step.ID,
			Type:   step.Type,
			Params: resolvedParams,
		}

		// Handle special step types
		switch step.Type {
		case StepCondition:
			i = handleCondition(ctx, execCtx, resolvedStep, steps, i)
			continue

		case StepDelay:
			if handleDelay(ctx, execCtx, resolvedStep, steps, i) {
				// Execution paused — will be resumed by delayed task worker
				return
			}
			continue

		case StepSetVariable:
			handleSetVariable(execCtx, resolvedStep)
			duration := time.Since(startTime).Milliseconds()
			_ = CreateStepLog(ctx, execCtx.ExecutionID, step.ID, step.Type, step.Params, nil, StepStatusSuccess, int(duration), "")
			continue

		case StepLog:
			handleLog(ctx, execCtx, resolvedStep)
			duration := time.Since(startTime).Milliseconds()
			_ = CreateStepLog(ctx, execCtx.ExecutionID, step.ID, step.Type, step.Params, nil, StepStatusSuccess, int(duration), "")
			continue
		}

		// Use registered action executor
		executor, ok := ActionRegistry[step.Type]
		if !ok {
			errMsg := fmt.Sprintf("unknown step type: %s", step.Type)
			_ = CreateStepLog(ctx, execCtx.ExecutionID, step.ID, step.Type, step.Params, nil, StepStatusFailed, 0, errMsg)
			_ = UpdateExecutionStatus(ctx, execCtx.ExecutionID, StatusFailed, step.ID, errMsg, execCtx.Variables)
			g.Log().Errorf(ctx, "scenario exec %d: %s", execCtx.ExecutionID, errMsg)
			return
		}

		output, err := executor.Execute(ctx, execCtx, resolvedStep)
		duration := time.Since(startTime).Milliseconds()

		if err != nil {
			_ = CreateStepLog(ctx, execCtx.ExecutionID, step.ID, step.Type, step.Params, output, StepStatusFailed, int(duration), err.Error())
			_ = UpdateExecutionStatus(ctx, execCtx.ExecutionID, StatusFailed, step.ID, err.Error(), execCtx.Variables)
			g.Log().Errorf(ctx, "scenario exec %d step %s failed: %v", execCtx.ExecutionID, step.ID, err)
			return
		}

		_ = CreateStepLog(ctx, execCtx.ExecutionID, step.ID, step.Type, step.Params, output, StepStatusSuccess, int(duration), "")

		// If the action returned variables, merge them
		if output != nil {
			for k, v := range output {
				if s, ok := v.(string); ok {
					execCtx.Variables[k] = s
				}
			}
		}
	}

	// All steps completed
	_ = UpdateExecutionStatus(ctx, execCtx.ExecutionID, StatusCompleted, "", "", execCtx.Variables)
}

// handleCondition evaluates an if/then/else condition and returns the next step index.
func handleCondition(ctx context.Context, execCtx *ExecutionContext, step *StepDSL, steps []StepDSL, currentIdx int) int {
	ifExpr, _ := step.Params["if"].(string)
	thenStep, _ := step.Params["then"].(string)
	elseStep, _ := step.Params["else"].(string)

	result := EvalCondition(ifExpr, execCtx.Variables)

	var targetStep string
	if result {
		targetStep = thenStep
	} else {
		targetStep = elseStep
	}

	output := map[string]interface{}{"condition_result": result, "target": targetStep}
	_ = CreateStepLog(ctx, execCtx.ExecutionID, step.ID, step.Type, step.Params, output, StepStatusSuccess, 0, "")

	if targetStep == "end" || targetStep == "" {
		// End scenario
		_ = UpdateExecutionStatus(ctx, execCtx.ExecutionID, StatusCompleted, step.ID, "", execCtx.Variables)
		return len(steps) // Will exit the loop
	}

	// Find the target step index
	for j, s := range steps {
		if s.ID == targetStep {
			return j - 1 // -1 because the loop will i++ after continue
		}
	}

	// Step not found — continue to next
	return currentIdx
}

// handleDelay creates a delayed task and pauses execution. Returns true if paused.
func handleDelay(ctx context.Context, execCtx *ExecutionContext, step *StepDSL, steps []StepDSL, currentIdx int) bool {
	durationStr, _ := step.Params["duration"].(string)
	duration := parseDuration(durationStr)

	if duration <= 0 {
		_ = CreateStepLog(ctx, execCtx.ExecutionID, step.ID, step.Type, step.Params, nil, StepStatusSkipped, 0, "invalid duration")
		return false
	}

	executeAt := time.Now().Add(duration).Unix()

	// Find the next step ID
	nextStepId := ""
	if currentIdx+1 < len(steps) {
		nextStepId = steps[currentIdx+1].ID
	}

	if nextStepId == "" {
		// No more steps after delay — just complete
		_ = CreateStepLog(ctx, execCtx.ExecutionID, step.ID, step.Type, step.Params, nil, StepStatusSuccess, 0, "")
		_ = UpdateExecutionStatus(ctx, execCtx.ExecutionID, StatusCompleted, step.ID, "", execCtx.Variables)
		return true
	}

	err := CreateDelayedTask(ctx, execCtx.MerchantID, execCtx.ExecutionID, nextStepId, executeAt)
	if err != nil {
		g.Log().Errorf(ctx, "scenario exec %d: failed to create delayed task: %v", execCtx.ExecutionID, err)
		_ = CreateStepLog(ctx, execCtx.ExecutionID, step.ID, step.Type, step.Params, nil, StepStatusFailed, 0, err.Error())
		_ = UpdateExecutionStatus(ctx, execCtx.ExecutionID, StatusFailed, step.ID, err.Error(), execCtx.Variables)
		return true
	}

	_ = CreateStepLog(ctx, execCtx.ExecutionID, step.ID, step.Type, step.Params,
		map[string]interface{}{"execute_at": executeAt, "next_step": nextStepId}, StepStatusSuccess, 0, "")
	_ = UpdateExecutionStatus(ctx, execCtx.ExecutionID, StatusWaiting, step.ID, "", execCtx.Variables)
	return true
}

// handleSetVariable sets a variable in the execution context.
func handleSetVariable(execCtx *ExecutionContext, step *StepDSL) {
	name, _ := step.Params["name"].(string)
	value, _ := step.Params["value"].(string)
	if name != "" {
		execCtx.Variables[name] = value
	}
}

// handleLog writes a log message.
func handleLog(ctx context.Context, execCtx *ExecutionContext, step *StepDSL) {
	message, _ := step.Params["message"].(string)
	level, _ := step.Params["level"].(string)
	if level == "" {
		level = "info"
	}

	switch level {
	case "error":
		g.Log().Errorf(ctx, "scenario exec %d log: %s", execCtx.ExecutionID, message)
	case "warning":
		g.Log().Warningf(ctx, "scenario exec %d log: %s", execCtx.ExecutionID, message)
	default:
		g.Log().Infof(ctx, "scenario exec %d log: %s", execCtx.ExecutionID, message)
	}
}

// ResumeExecution resumes a paused execution from a given step.
// Called by the delayed task worker.
func ResumeExecution(ctx context.Context, executionId uint64, resumeStepId string) {
	exec, err := GetExecution(ctx, executionId)
	if err != nil || exec == nil {
		g.Log().Errorf(ctx, "scenario: cannot resume execution %d: %v", executionId, err)
		return
	}

	scenario, err := GetScenario(ctx, exec.MerchantId, exec.ScenarioId)
	if err != nil || scenario == nil {
		g.Log().Errorf(ctx, "scenario: cannot find scenario %d for execution %d", exec.ScenarioId, executionId)
		return
	}

	dsl, err := ParseDSL(scenario.ScenarioJson)
	if err != nil {
		g.Log().Errorf(ctx, "scenario: cannot parse DSL for scenario %d: %v", exec.ScenarioId, err)
		return
	}

	// Restore variables
	vars := make(map[string]string)
	if exec.Variables != "" {
		_ = json.Unmarshal([]byte(exec.Variables), &vars)
	}

	// Restore trigger data
	triggerData := make(map[string]interface{})
	if exec.TriggerData != "" {
		_ = json.Unmarshal([]byte(exec.TriggerData), &triggerData)
	}

	execCtx := &ExecutionContext{
		ExecutionID: exec.Id,
		MerchantID:  exec.MerchantId,
		ScenarioID:  exec.ScenarioId,
		Variables:   vars,
		TriggerData: triggerData,
	}

	// Find the step index to resume from
	startIdx := 0
	for i, s := range dsl.Steps {
		if s.ID == resumeStepId {
			startIdx = i
			break
		}
	}

	executeSteps(ctx, execCtx, dsl.Steps, startIdx)
}

// ──────────────────────────────────────────
// Trigger matching
// ──────────────────────────────────────────

// MatchAndRunWebhookScenarios finds and runs scenarios triggered by a webhook event.
func MatchAndRunWebhookScenarios(ctx context.Context, merchantId uint64, event string, data map[string]interface{}) {
	scenarios, err := GetScenariosByTrigger(ctx, merchantId, TriggerWebhookEvent, event)
	if err != nil {
		g.Log().Errorf(ctx, "scenario: error finding scenarios for event %s: %v", event, err)
		return
	}

	triggerData := map[string]interface{}{
		"event": event,
		"data":  data,
	}

	for _, sc := range scenarios {
		go RunScenarioByIds(ctx, merchantId, sc.Id, sc.ScenarioJson, triggerData)
	}
}

// MatchAndRunBotCommand finds and runs scenarios triggered by a bot command.
func MatchAndRunBotCommand(ctx context.Context, merchantId uint64, command string, chatId int64, username string) {
	scenarios, err := GetScenariosByTrigger(ctx, merchantId, TriggerBotCommand, command)
	if err != nil {
		g.Log().Errorf(ctx, "scenario: error finding scenarios for command %s: %v", command, err)
		return
	}

	triggerData := map[string]interface{}{
		"command":  command,
		"chat_id":  chatId,
		"username": username,
	}

	for _, sc := range scenarios {
		go RunScenarioByIds(ctx, merchantId, sc.Id, sc.ScenarioJson, triggerData)
	}
}

// MatchAndRunButtonClick finds and runs scenarios triggered by an inline button callback.
func MatchAndRunButtonClick(ctx context.Context, merchantId uint64, action string, chatId int64, username string) {
	scenarios, err := GetScenariosByTrigger(ctx, merchantId, TriggerButtonClick, action)
	if err != nil {
		g.Log().Errorf(ctx, "scenario: error finding scenarios for button %s: %v", action, err)
		return
	}

	triggerData := map[string]interface{}{
		"action":   action,
		"chat_id":  chatId,
		"username": username,
	}

	for _, sc := range scenarios {
		go RunScenarioByIds(ctx, merchantId, sc.Id, sc.ScenarioJson, triggerData)
	}
}

// ──────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────

// flattenTriggerData converts nested trigger data to a flat key→value map.
func flattenTriggerData(data map[string]interface{}) map[string]string {
	result := make(map[string]string)
	flattenMap("", data, result)
	return result
}

func flattenMap(prefix string, data map[string]interface{}, result map[string]string) {
	for k, v := range data {
		key := k
		if prefix != "" {
			key = prefix + "_" + k
		}
		switch val := v.(type) {
		case string:
			result[key] = val
		case float64:
			result[key] = fmt.Sprintf("%v", val)
		case int:
			result[key] = fmt.Sprintf("%d", val)
		case int64:
			result[key] = fmt.Sprintf("%d", val)
		case bool:
			if val {
				result[key] = "true"
			} else {
				result[key] = "false"
			}
		case map[string]interface{}:
			flattenMap(key, val, result)
		default:
			if val != nil {
				result[key] = fmt.Sprintf("%v", val)
			}
		}
	}
}

// parseDuration converts duration strings like "1m", "1h", "1d", "30s" to time.Duration.
func parseDuration(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// Try standard Go duration first
	d, err := time.ParseDuration(s)
	if err == nil {
		return d
	}

	// Handle "d" suffix for days
	if strings.HasSuffix(s, "d") {
		numStr := strings.TrimSuffix(s, "d")
		var days int
		if _, err := fmt.Sscanf(numStr, "%d", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}

	return 0
}

// ValidateDSL validates a scenario DSL for correctness.
func ValidateDSL(dsl *ScenarioDSL) []string {
	var errors []string

	if dsl.Trigger.Type == "" {
		errors = append(errors, "trigger.type is required")
	}

	validTriggers := map[string]bool{
		TriggerWebhookEvent: true,
		TriggerBotCommand:   true,
		TriggerButtonClick:  true,
		TriggerSchedule:     true,
		TriggerManual:       true,
	}
	if !validTriggers[dsl.Trigger.Type] {
		errors = append(errors, fmt.Sprintf("unknown trigger type: %s", dsl.Trigger.Type))
	}

	if len(dsl.Steps) == 0 {
		errors = append(errors, "at least one step is required")
	}

	validSteps := map[string]bool{
		StepSendTelegram: true,
		StepHTTPRequest:  true,
		StepDelay:        true,
		StepCondition:    true,
		StepSetVariable:  true,
		StepUniBeeAPI:    true,
		StepSendEmail:    true,
		StepLog:          true,
	}

	stepIds := map[string]bool{}
	for _, step := range dsl.Steps {
		if step.ID == "" {
			errors = append(errors, "step.id is required for all steps")
		}
		if stepIds[step.ID] {
			errors = append(errors, fmt.Sprintf("duplicate step id: %s", step.ID))
		}
		stepIds[step.ID] = true

		if !validSteps[step.Type] {
			errors = append(errors, fmt.Sprintf("unknown step type: %s in step %s", step.Type, step.ID))
		}
	}

	return errors
}
