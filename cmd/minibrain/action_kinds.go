package main

type ActionKind string

const (
	ActionRead            ActionKind = "READ"
	ActionReadRequest     ActionKind = "READ REQUEST"
	ActionReadApproved    ActionKind = "READ APPROVED"
	ActionReadDenied      ActionKind = "READ DENIED"
	ActionReadAlways      ActionKind = "READ ALWAYS APPROVED"
	ActionWrite           ActionKind = "WRITE"
	ActionDelete          ActionKind = "DELETE"
	ActionPatch           ActionKind = "PATCH"
	ActionPatchFailed     ActionKind = "PATCH FAILED"
	ActionChangesBlocked  ActionKind = "CHANGES BLOCKED"
	ActionChangesDenied   ActionKind = "CHANGES DENIED"
	ActionChangesAuto     ActionKind = "CHANGES AUTO-APPLY ENABLED"
	ActionError           ActionKind = "ERROR"
	ActionModel           ActionKind = "MODEL"
	ActionMemory          ActionKind = "MEMORY"
	ActionRaw             ActionKind = "RAW OUTPUT"
	ActionInfo            ActionKind = "INFO"
)

func formatAction(kind ActionKind, detail string) string {
	if detail == "" {
		return string(kind)
	}
	return string(kind) + ": " + detail
}
