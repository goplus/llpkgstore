package generator

type Generator interface {
	Generate() error
	Check() error
}
