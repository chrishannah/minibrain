package agent

type PermissionState struct {
	Project    ProjectConfig
	AllowRead  bool
	AllowWrite bool
	DenyWrite  bool
}

func ResolvePermissionState(root string, envRead, envWrite bool) PermissionState {
	proj := LoadProjectConfig(root)
	allowRead := envRead || proj.AllowReadAlways
	allowWrite := envWrite || proj.AllowWriteAlways
	denyWrite := proj.DenyWriteAlways
	if denyWrite {
		allowWrite = false
	}
	return PermissionState{
		Project:    proj,
		AllowRead:  allowRead,
		AllowWrite: allowWrite,
		DenyWrite:  denyWrite,
	}
}
