package scenario

// ScenarioDSL is the top-level JSON structure of a scenario.
type ScenarioDSL struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Enabled   bool              `json:"enabled"`
	Trigger   TriggerDSL        `json:"trigger"`
	Variables map[string]string `json:"variables,omitempty"`
	Steps     []StepDSL         `json:"steps"`
}

// TriggerDSL describes what starts the scenario.
type TriggerDSL struct {
	Type  string `json:"type"`  // webhook_event, bot_command, button_click, schedule, manual
	Event string `json:"event"` // event name, command, cron expression
}

// StepDSL describes a single step in the scenario.
type StepDSL struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"` // send_telegram, http_request, delay, condition, set_variable, unibee_api, send_email, log
	Params map[string]interface{} `json:"params"`
}

// ButtonDSL describes a Telegram inline button.
type ButtonDSL struct {
	Text   string `json:"text"`
	Action string `json:"action"`
}

// Trigger types
const (
	TriggerWebhookEvent = "webhook_event"
	TriggerBotCommand   = "bot_command"
	TriggerButtonClick  = "button_click"
	TriggerSchedule     = "schedule"
	TriggerManual       = "manual"
)

// Step types
const (
	StepSendTelegram = "send_telegram"
	StepHTTPRequest  = "http_request"
	StepDelay        = "delay"
	StepCondition    = "condition"
	StepSetVariable  = "set_variable"
	StepUniBeeAPI    = "unibee_api"
	StepSendEmail    = "send_email"
	StepLog          = "log"
)

// Execution statuses
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusWaiting   = "waiting" // waiting for delayed task / callback
)

// Step log statuses
const (
	StepStatusSuccess = "success"
	StepStatusFailed  = "failed"
	StepStatusSkipped = "skipped"
)

// ExecutionContext holds runtime state while executing a scenario.
type ExecutionContext struct {
	ExecutionID uint64
	MerchantID  uint64
	ScenarioID  uint64
	Variables   map[string]string
	TriggerData map[string]interface{}
}
