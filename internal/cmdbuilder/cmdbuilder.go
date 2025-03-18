package cmdbuilder

import (
	"fmt"
	"os/exec"
	"strings"
)

type CmdBuilderSerilizer func(k, v string) string

type Options func(*CmdBuilder)

type CmdBuilder struct {
	serializer CmdBuilderSerilizer
	name       string
	subcommand string
	args       []string
	objs       []string
}

func WithConanSerializer() Options {
	return func(cb *CmdBuilder) {
		cb.serializer = func(k, v string) string {
			return fmt.Sprintf(`--%s=%s`, k, v)
		}
	}
}

func NewCmdBuilder(opts ...Options) *CmdBuilder {
	c := &CmdBuilder{}

	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *CmdBuilder) SetName(n string) {
	c.name = n
}

func (c *CmdBuilder) SetSubcommand(s string) {
	c.subcommand = s
}

func (c *CmdBuilder) SetArg(k, v string) {
	c.args = append(c.args, c.serializer(k, v))
}

func (c *CmdBuilder) SetObj(o string) {
	c.objs = append(c.objs, o)
}

func (c *CmdBuilder) Name() string {
	return c.name
}

func (c *CmdBuilder) Subcommand() string {
	return c.subcommand
}

func (c *CmdBuilder) Args() []string {
	return c.args
}

func (c *CmdBuilder) Objs() []string {
	return c.objs
}

func (c *CmdBuilder) String() string {
	return fmt.Sprintf("%s %s %s %s", c.name, c.subcommand, c.objs, strings.Join(c.Args(), " "))
}

func (c *CmdBuilder) Cmd() *exec.Cmd {
	cmds := append([]string{c.subcommand}, c.objs...)
	cmds = append(cmds, c.Args()...)
	return exec.Command(c.name, cmds...)
}
