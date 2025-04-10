package conan

type properties struct {
	PkgName string `json:"pkg_config_name"`
}

type cppInfo struct {
	Properties properties `json:"properties"`
}

type packageInfo struct {
	Name    string             `json:"name"`
	CppInfo map[string]cppInfo `json:"cpp_info"`
}

type installOutput struct {
	Graph struct {
		Nodes map[string]packageInfo `json:"nodes"`
	} `json:"graph"`
}

type dependency struct {
	Ref     string `json:"ref"`
	IsBuild bool   `json:"build"`
}

type dependencyInfo struct {
	Requires []string `json:"requires"`
}

type graphInfo struct {
	Name         string                `json:"name"`
	Info         dependencyInfo        `json:"info"`
	Dependencies map[string]dependency `json:"dependencies"`
}

type graphOutput struct {
	Graph struct {
		Nodes map[string]graphInfo `json:"nodes"`
	} `json:"graph"`
}
