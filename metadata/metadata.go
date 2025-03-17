package metadata

// MetadataMap represents llpkgstore.json
type MetadataMap map[string]*Metadata

type Metadata struct {
	VersionMappings []*VersionMapping `json:"versions"`
}

type VersionMapping struct {
	CVersion   string   `json:"c"`
	GoVersions []string `json:"go"`
}
