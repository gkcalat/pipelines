package model

type Task struct {
	UUID              string `gorm:"column:UUID; not null; primary_key"`
	Namespace         string `gorm:"column:Namespace; not null;"`
	PipelineName      string `gorm:"column:PipelineName; not null;"`
	RunUUID           string `gorm:"column:RunUUID; not null;"`
	MLMDExecutionID   string `gorm:"column:MLMDExecutionID; not null;"`
	CreatedTimestamp  int64  `gorm:"column:CreatedTimestamp; not null;"`
	StartedTimestamp  int64  `gorm:"column:StartedTimestamp;"`
	FinishedTimestamp int64  `gorm:"column:FinishedTimestamp;"`
	Fingerprint       string `gorm:"column:Fingerprint; not null;"`
	Name              string `gorm:"column:Name; default:null"`
	ParentTaskUUID    string `gorm:"column:ParentTaskUUID; default:null"`
	State             string `gorm:"column:State; default:null;"`
	StateHistory      string `gorm:"column:StateHistory; default:null;"`
	MLMDInputs        string `gorm:"column:MLMDInputs; default:null; size:65535;"`
	MLMDOutputs       string `gorm:"column:MLMDOutputs; default:null; size:65535;"`
}

func (t Task) PrimaryKeyColumnName() string {
	return "UUID"
}

func (t Task) DefaultSortField() string {
	return "CreatedTimestamp"
}

func (t Task) APIToModelFieldMap() map[string]string {
	return taskAPIToModelFieldMap
}

func (t Task) GetModelName() string {
	return "tasks"
}

func (t Task) GetSortByFieldPrefix(s string) string {
	return "tasks."
}

func (t Task) GetKeyFieldPrefix() string {
	return "tasks."
}

var taskAPIToModelFieldMap = map[string]string{
	"task_id":        "UUID",
	"namespace":      "Namespace",
	"pipeline_name":  "PipelineName",
	"run_id":         "RunUUID ",
	"execution_id":   "MLMDExecutionID",
	"create_time":    "CreatedTimestamp",
	"start_time":     "StartedTimestamp",
	"end_time":       "FinishedTimestamp",
	"fingerprint":    "Fingerprint",
	"state":          "State",
	"state_history":  "StateHistory",
	"display_name":   "Name",
	"parent_task_id": "ParentTaskUUID",
}

func (t Task) GetField(name string) (string, bool) {
	if field, ok := taskAPIToModelFieldMap[name]; ok {
		return field, true
	}
	return "", false
}

func (t Task) GetFieldValue(name string) interface{} {
	switch name {
	case "UUID":
		return t.UUID
	case "Namespace":
		return t.Namespace
	case "PipelineName":
		return t.PipelineName
	case "RunUUID":
		return t.RunUUID
	case "MLMDExecutionID":
		return t.MLMDExecutionID
	case "CreatedTimestamp":
		return t.CreatedTimestamp
	case "FinishedTimestamp":
		return t.FinishedTimestamp
	case "Fingerprint":
		return t.Fingerprint
	case "ParentTaskUUID":
		return t.ParentTaskUUID
	case "State":
		return t.State
	case "StateHistory":
		return t.StateHistory
	case "Name":
		return t.Name
	case "MLMDInputs":
		return t.MLMDInputs
	case "MLMDOutputs":
		return t.MLMDOutputs
	default:
		return nil
	}
}
