package main

type CapsLock struct {
	CapabilityInfo []struct {
		Capability     string `json:"capability,omitzero"`
		CapabilityType string `json:"capabilityType,omitzero"`
		DepPath        string `json:"depPath,omitzero"`
		PackageDir     string `json:"packageDir,omitzero"`
		PackageName    string `json:"packageName,omitzero"`
		Path           []struct {
			Name    string `json:"name,omitzero"`
			Package string `json:"package,omitzero"`
			Site    *struct {
				Column   string `json:"column,omitzero"`
				Filename string `json:"filename,omitzero"`
				Line     string `json:"line,omitzero"`
			} `json:"site,omitempty,omitzero"`
		} `json:"path"`
	} `json:"capabilityInfo"`
	ModuleInfo []struct {
		Path    string `json:"path,omitzero"`
		Version string `json:"version,omitzero"`
	} `json:"moduleInfo"`
	PackageInfo []struct {
		IgnoredFiles []string `json:"ignoredFiles,omitempty"`
		Path         string   `json:"path,omitzero"`
	} `json:"packageInfo"`
}
