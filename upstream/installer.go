package upstream

// Installer represents a package installer that can download, install, and locate binaries from a remote repository.
// It provides methods to install packages to specific directories and search for installed package information.
type Installer interface {
	Name() string
	Config() map[string]string
	// Install downloads and installs the specified package.
	// The outputDir is where build artifacts (e.g., .pc files, headers) are stored.
	// Returns an error if installation fails, all the pkgConfigFiles if success.
	Install(pkg Package, outputDir string) (pkgConfigFiles []string, err error)
	// Search checks remote repository for the specified package availability.
	// Returns the search results text and any encountered errors.
	Search(pkg Package) ([]string, error)

	// Dependency retrieves the list of dependencies for the specified package.
	// It queries the package manager's repository to determine required packages
	// and their versions. The returned list includes both direct and transitive
	// dependencies. An error is returned if the package is not found or dependency
	// resolution fails.
	Dependency(pkg Package) (dependencies []Package, err error)
}
