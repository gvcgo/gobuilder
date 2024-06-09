package builder

type Builder struct {
	WorkDir    string   `json:"work_dir"`
	ArchOSList []string `json:"arch_os_list"`
	BuildArgs  []string `json:"build_args"`
}
