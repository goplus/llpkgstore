package generator

type Generator interface {
	Generate(toDir string) error
	Check(baseDir string) error
}
