package context

// InputMode defines the various text input modes
type InputMode int

const (
	InputNone InputMode = iota
	InputSearch
	InputRename
	InputGoto
	InputAuth
	InputFuzzySearch
	InputZip
	InputUnzip
	InputCreate
	InputConflictRename
	InputKeybinding
)

// LogLevel defines the severity level of a log entry
type LogLevel int

const (
	LogInfo LogLevel = iota
	LogSuccess
	LogWarn
	LogError
)

// LogStatus defines the current state of an operation
type LogStatus int

const (
	StatusPending LogStatus = iota
	StatusRunning
	StatusSuccess
	StatusError
)

// HelpState holds state for the help view
type HelpState struct {
	Cursor int
	Offset int
}
