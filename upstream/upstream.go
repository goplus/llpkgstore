package upstream

// Upstream represents a binary package and its installation configuration.
// It encapsulates an Installer responsible for the installation process and a Package
// defining the target library's metadata. The Installer uses the Package's details
// to download, build, and install the binary into the system or designated directories.
type Upstream struct {
	// Installer is the package installer implementation responsible for executing the installation process
	Installer Installer
	// Pkg defines the target library's metadata including name and version required for installation
	Pkg Package
}

// Package defines the metadata required to identify and install a software library.
// The Name and Version fields provide precise identification of the library.
type Package struct {
	Name    string
	Version string
}
