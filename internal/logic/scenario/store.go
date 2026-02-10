package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	dao "unibee/internal/dao/default"
	entity "unibee/internal/model/entity/default"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ──────────────────────────────────────────
// Scenario CRUD
// ──────────────────────────────────────────

// CreateScenario creates a new scenario from DSL JSON.
func CreateScenario(ctx context.Context, merchantId uint64, name, description string, dsl *ScenarioDSL) (*entity.MerchantScenario, error) {
	dslBytes, err := json.Marshal(dsl)
	if err != nil {
		return nil, fmt.Errorf("marshal scenario json: %w", err)
	}

	now := gtime.Now()
	row := &entity.MerchantScenario{
		MerchantId:   merchantId,
		Name:         name,
		Description:  description,
		ScenarioJson: string(dslBytes),
		Enabled:      0,
		TriggerType:  dsl.Trigger.Type,
		TriggerValue: dsl.Trigger.Event,
		GmtCreate:    now,
		GmtModify:    now,
		CreateTime:   time.Now().Unix(),
	}

	result, err := dao.MerchantScenario.Ctx(ctx).Data(row).OmitEmpty().Insert()
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	row.Id = uint64(id)
	return row, nil
}

// UpdateScenario updates an existing scenario.
func UpdateScenario(ctx context.Context, merchantId, scenarioId uint64, name, description string, dsl *ScenarioDSL) error {
	dslBytes, err := json.Marshal(dsl)
	if err != nil {
		return fmt.Errorf("marshal scenario json: %w", err)
	}

	_, err = dao.MerchantScenario.Ctx(ctx).
		Where(dao.MerchantScenario.Columns().Id, scenarioId).
		Where(dao.MerchantScenario.Columns().MerchantId, merchantId).
		Where(dao.MerchantScenario.Columns().IsDeleted, 0).
		Data(g.Map{
			dao.MerchantScenario.Columns().Name:         name,
			dao.MerchantScenario.Columns().Description:  description,
			dao.MerchantScenario.Columns().ScenarioJson: string(dslBytes),
			dao.MerchantScenario.Columns().TriggerType:  dsl.Trigger.Type,
			dao.MerchantScenario.Columns().TriggerValue: dsl.Trigger.Event,
			dao.MerchantScenario.Columns().GmtModify:    gtime.Now(),
		}).Update()
	return err
}

// DeleteScenario soft-deletes a scenario.
func DeleteScenario(ctx context.Context, merchantId, scenarioId uint64) error {
	_, err := dao.MerchantScenario.Ctx(ctx).
		Where(dao.MerchantScenario.Columns().Id, scenarioId).
		Where(dao.MerchantScenario.Columns().MerchantId, merchantId).
		Data(g.Map{
			dao.MerchantScenario.Columns().IsDeleted: 1,
			dao.MerchantScenario.Columns().GmtModify: gtime.Now(),
		}).Update()
	return err
}

// ToggleScenario enables/disables a scenario.
func ToggleScenario(ctx context.Context, merchantId, scenarioId uint64, enabled bool) error {
	val := 0
	if enabled {
		val = 1
	}
	_, err := dao.MerchantScenario.Ctx(ctx).
		Where(dao.MerchantScenario.Columns().Id, scenarioId).
		Where(dao.MerchantScenario.Columns().MerchantId, merchantId).
		Where(dao.MerchantScenario.Columns().IsDeleted, 0).
		Data(g.Map{
			dao.MerchantScenario.Columns().Enabled:   val,
			dao.MerchantScenario.Columns().GmtModify: gtime.Now(),
		}).Update()
	return err
}

// GetScenario retrieves a single scenario by ID.
func GetScenario(ctx context.Context, merchantId, scenarioId uint64) (*entity.MerchantScenario, error) {
	var row entity.MerchantScenario
	err := dao.MerchantScenario.Ctx(ctx).
		Where(dao.MerchantScenario.Columns().Id, scenarioId).
		Where(dao.MerchantScenario.Columns().MerchantId, merchantId).
		Where(dao.MerchantScenario.Columns().IsDeleted, 0).
		Scan(&row)
	if err != nil {
		return nil, err
	}
	if row.Id == 0 {
		return nil, nil
	}
	return &row, nil
}

// ListScenarios returns all scenarios for a merchant.
func ListScenarios(ctx context.Context, merchantId uint64) ([]*entity.MerchantScenario, error) {
	var list []*entity.MerchantScenario
	err := dao.MerchantScenario.Ctx(ctx).
		Where(dao.MerchantScenario.Columns().MerchantId, merchantId).
		Where(dao.MerchantScenario.Columns().IsDeleted, 0).
		OrderDesc(dao.MerchantScenario.Columns().Id).
		Scan(&list)
	return list, err
}

// GetScenariosByTrigger finds enabled scenarios matching a trigger type+value for a merchant.
func GetScenariosByTrigger(ctx context.Context, merchantId uint64, triggerType, triggerValue string) ([]*entity.MerchantScenario, error) {
	var list []*entity.MerchantScenario
	q := dao.MerchantScenario.Ctx(ctx).
		Where(dao.MerchantScenario.Columns().MerchantId, merchantId).
		Where(dao.MerchantScenario.Columns().Enabled, 1).
		Where(dao.MerchantScenario.Columns().IsDeleted, 0).
		Where(dao.MerchantScenario.Columns().TriggerType, triggerType)

	if triggerValue != "" {
		q = q.Where(dao.MerchantScenario.Columns().TriggerValue, triggerValue)
	}

	err := q.Scan(&list)
	return list, err
}

// ParseDSL parses scenario_json into ScenarioDSL.
func ParseDSL(jsonStr string) (*ScenarioDSL, error) {
	var dsl ScenarioDSL
	if err := json.Unmarshal([]byte(jsonStr), &dsl); err != nil {
		return nil, fmt.Errorf("invalid scenario JSON: %w", err)
	}
	return &dsl, nil
}

// ──────────────────────────────────────────
// Execution CRUD
// ──────────────────────────────────────────

// CreateExecution starts a new scenario execution run.
func CreateExecution(ctx context.Context, merchantId, scenarioId uint64, triggerData map[string]interface{}, variables map[string]string) (uint64, error) {
	triggerJSON, _ := json.Marshal(triggerData)
	varsJSON, _ := json.Marshal(variables)
	now := gtime.Now()

	result, err := dao.MerchantScenarioExecution.Ctx(ctx).Data(&entity.MerchantScenarioExecution{
		MerchantId:  merchantId,
		ScenarioId:  scenarioId,
		TriggerData: string(triggerJSON),
		Status:      StatusRunning,
		Variables:   string(varsJSON),
		StartedAt:   time.Now().Unix(),
		GmtCreate:   now,
		GmtModify:   now,
		CreateTime:  time.Now().Unix(),
	}).OmitEmpty().Insert()
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

// UpdateExecutionStatus updates the status of an execution.
func UpdateExecutionStatus(ctx context.Context, executionId uint64, status, currentStep, errorMsg string, variables map[string]string) error {
	data := g.Map{
		dao.MerchantScenarioExecution.Columns().Status:    status,
		dao.MerchantScenarioExecution.Columns().GmtModify: gtime.Now(),
	}
	if currentStep != "" {
		data[dao.MerchantScenarioExecution.Columns().CurrentStep] = currentStep
	}
	if errorMsg != "" {
		data[dao.MerchantScenarioExecution.Columns().ErrorMessage] = errorMsg
	}
	if variables != nil {
		varsJSON, _ := json.Marshal(variables)
		data[dao.MerchantScenarioExecution.Columns().Variables] = string(varsJSON)
	}
	if status == StatusCompleted || status == StatusFailed {
		data[dao.MerchantScenarioExecution.Columns().FinishedAt] = time.Now().Unix()
	}

	_, err := dao.MerchantScenarioExecution.Ctx(ctx).
		Where(dao.MerchantScenarioExecution.Columns().Id, executionId).
		Data(data).Update()
	return err
}

// GetExecution retrieves an execution by ID.
func GetExecution(ctx context.Context, executionId uint64) (*entity.MerchantScenarioExecution, error) {
	var row entity.MerchantScenarioExecution
	err := dao.MerchantScenarioExecution.Ctx(ctx).
		Where(dao.MerchantScenarioExecution.Columns().Id, executionId).
		Scan(&row)
	if err != nil {
		return nil, err
	}
	if row.Id == 0 {
		return nil, nil
	}
	return &row, nil
}

// ListExecutions returns recent executions for a scenario.
func ListExecutions(ctx context.Context, merchantId uint64, scenarioId uint64, page, size int) ([]*entity.MerchantScenarioExecution, int, error) {
	q := dao.MerchantScenarioExecution.Ctx(ctx).
		Where(dao.MerchantScenarioExecution.Columns().MerchantId, merchantId)
	if scenarioId > 0 {
		q = q.Where(dao.MerchantScenarioExecution.Columns().ScenarioId, scenarioId)
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	var list []*entity.MerchantScenarioExecution
	err = q.OrderDesc(dao.MerchantScenarioExecution.Columns().Id).
		Page(page, size).
		Scan(&list)
	return list, total, err
}

// ──────────────────────────────────────────
// Step Log
// ──────────────────────────────────────────

// CreateStepLog records a single step execution.
func CreateStepLog(ctx context.Context, executionId uint64, stepId, stepType string, inputData, outputData interface{}, status string, durationMs int, errorMsg string) error {
	inJSON, _ := json.Marshal(inputData)
	outJSON, _ := json.Marshal(outputData)

	_, err := dao.MerchantScenarioStepLog.Ctx(ctx).Data(&entity.MerchantScenarioStepLog{
		ExecutionId:  executionId,
		StepId:       stepId,
		StepType:     stepType,
		InputData:    string(inJSON),
		OutputData:   string(outJSON),
		Status:       status,
		DurationMs:   durationMs,
		ErrorMessage: errorMsg,
		GmtCreate:    gtime.Now(),
	}).OmitEmpty().Insert()
	return err
}

// GetStepLogs returns all step logs for an execution.
func GetStepLogs(ctx context.Context, executionId uint64) ([]*entity.MerchantScenarioStepLog, error) {
	var list []*entity.MerchantScenarioStepLog
	err := dao.MerchantScenarioStepLog.Ctx(ctx).
		Where(dao.MerchantScenarioStepLog.Columns().ExecutionId, executionId).
		OrderAsc(dao.MerchantScenarioStepLog.Columns().Id).
		Scan(&list)
	return list, err
}

// ──────────────────────────────────────────
// Delayed Tasks
// ──────────────────────────────────────────

// CreateDelayedTask schedules a step execution for later.
func CreateDelayedTask(ctx context.Context, merchantId, executionId uint64, stepId string, executeAt int64) error {
	_, err := dao.MerchantScenarioDelayedTask.Ctx(ctx).Data(&entity.MerchantScenarioDelayedTask{
		MerchantId:  merchantId,
		ExecutionId: executionId,
		StepId:      stepId,
		ExecuteAt:   executeAt,
		Status:      StatusPending,
		GmtCreate:   gtime.Now(),
		GmtModify:   gtime.Now(),
	}).OmitEmpty().Insert()
	return err
}

// GetPendingDelayedTasks returns all tasks ready to execute.
func GetPendingDelayedTasks(ctx context.Context) ([]*entity.MerchantScenarioDelayedTask, error) {
	var list []*entity.MerchantScenarioDelayedTask
	err := dao.MerchantScenarioDelayedTask.Ctx(ctx).
		Where(dao.MerchantScenarioDelayedTask.Columns().Status, StatusPending).
		WhereLTE(dao.MerchantScenarioDelayedTask.Columns().ExecuteAt, time.Now().Unix()).
		OrderAsc(dao.MerchantScenarioDelayedTask.Columns().ExecuteAt).
		Scan(&list)
	return list, err
}

// MarkDelayedTaskExecuted marks a delayed task as executed.
func MarkDelayedTaskExecuted(ctx context.Context, taskId uint64) error {
	_, err := dao.MerchantScenarioDelayedTask.Ctx(ctx).
		Where(dao.MerchantScenarioDelayedTask.Columns().Id, taskId).
		Data(g.Map{
			dao.MerchantScenarioDelayedTask.Columns().Status:    "executed",
			dao.MerchantScenarioDelayedTask.Columns().GmtModify: gtime.Now(),
		}).Update()
	return err
}

// ──────────────────────────────────────────
// Telegram User Mapping
// ──────────────────────────────────────────

// UpsertTelegramUser creates or updates a telegram_chat_id ↔ user mapping.
func UpsertTelegramUser(ctx context.Context, merchantId uint64, telegramChatId int64, username, firstName, lastName string) error {
	chatIdStr := strconv.FormatInt(telegramChatId, 10)
	var existing entity.MerchantTelegramUser
	err := dao.MerchantTelegramUser.Ctx(ctx).
		Where(dao.MerchantTelegramUser.Columns().MerchantId, merchantId).
		Where(dao.MerchantTelegramUser.Columns().TelegramChatId, chatIdStr).
		Where(dao.MerchantTelegramUser.Columns().IsDeleted, 0).
		Scan(&existing)
	if err != nil {
		return err
	}

	now := gtime.Now()
	if existing.Id > 0 {
		_, err = dao.MerchantTelegramUser.Ctx(ctx).
			Where(dao.MerchantTelegramUser.Columns().Id, existing.Id).
			Data(g.Map{
				dao.MerchantTelegramUser.Columns().TelegramUsername: username,
				dao.MerchantTelegramUser.Columns().FirstName:        firstName,
				dao.MerchantTelegramUser.Columns().LastName:         lastName,
				dao.MerchantTelegramUser.Columns().GmtModify:        now,
			}).Update()
		return err
	}

	_, err = dao.MerchantTelegramUser.Ctx(ctx).Data(&entity.MerchantTelegramUser{
		MerchantId:       merchantId,
		TelegramChatId:   chatIdStr,
		TelegramUsername: username,
		FirstName:        firstName,
		LastName:         lastName,
		GmtCreate:        now,
		GmtModify:        now,
		CreateTime:       time.Now().Unix(),
	}).OmitEmpty().Insert()
	return err
}

// GetTelegramUserByChatId looks up a user by telegram chat id.
func GetTelegramUserByChatId(ctx context.Context, merchantId uint64, chatId int64) (*entity.MerchantTelegramUser, error) {
	chatIdStr := strconv.FormatInt(chatId, 10)
	var row entity.MerchantTelegramUser
	err := dao.MerchantTelegramUser.Ctx(ctx).
		Where(dao.MerchantTelegramUser.Columns().MerchantId, merchantId).
		Where(dao.MerchantTelegramUser.Columns().TelegramChatId, chatIdStr).
		Where(dao.MerchantTelegramUser.Columns().IsDeleted, 0).
		Scan(&row)
	if err != nil {
		return nil, err
	}
	if row.Id == 0 {
		return nil, nil
	}
	return &row, nil
}

// LinkTelegramToUser links a telegram chat to a UniBee userId.
func LinkTelegramToUser(ctx context.Context, merchantId uint64, chatId int64, userId uint64) error {
	_, err := dao.MerchantTelegramUser.Ctx(ctx).
		Where(dao.MerchantTelegramUser.Columns().MerchantId, merchantId).
		Where(dao.MerchantTelegramUser.Columns().TelegramChatId, chatId).
		Data(g.Map{
			dao.MerchantTelegramUser.Columns().UserId:    userId,
			dao.MerchantTelegramUser.Columns().GmtModify: gtime.Now(),
		}).Update()
	return err
}
