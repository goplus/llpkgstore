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

type conanOutput struct {
	Graph struct {
		Nodes map[string]packageInfo `json:"nodes"`
	} `json:"graph"`
}
