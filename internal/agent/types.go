package agent

type FileRef struct {
	Mention string
	Path    string
	Content string
	Err     error
}

type WriteOp struct {
	Path    string
	Content string
}

type DeleteOp struct {
	Path string
}

type PatchOp struct {
	Path  string
	Patch string
}

type Result struct {
	LLMOutput         string
	ProposedWrites    []WriteOp
	ProposedDeletes   []DeleteOp
	ProposedPatches   []PatchOp
	AppliedWrites     []WriteOp
	AppliedDeletes    []DeleteOp
	AppliedPatches    []PatchOp
	Applied           bool
	PrefrontalPath    string
	Mentions          []string
	FileRefs          []FileRef
	FileList          []string
	FileListTruncated bool
	Memory            MemoryStats
	Condensed         bool
}

type Config struct {
	RootDir             string
	BrainDir            string
	Model               string
	TimeoutSec          int
	NeoPath             string
	PrefrontalPath      string
	StmMaxBytes         int
	StmContextBytes     int
	ConversationBytes   int
	ContextBudgetTokens int
	AllowReadAll        bool
	ApplyWrites         bool
	ReadPaths           []string
	MaxFilesListed      int
	MaxFileBytes        int
	MaxTotalReadBytes   int
}

type MemoryStats struct {
	LtmLines int
	StmLines int
	LtmBytes int
	StmBytes int
}
