package k3

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nokia/ntt/build"
	"github.com/nokia/ntt/internal/compdb"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
)

// DefaultEnv is the default environment to build k3-based test suites.
var DefaultEnv = map[string]string{
	"CXX":        "g++",
	"CC":         "gcc",
	"ASN1C":      "asn1",
	"ASN1CFLAGS": "-reservedWords ffs -c -charIntegers -listingFile -messageFormat emacs -noDefines -valuerefs -debug -root -soed",
	"ASN2TTCN":   "asn1tottcn3",
	"OSSINFO":    filepath.Join(DataDir(), "asn1"),
	"K3C":        Compiler(),
	"K3R":        Runtime(),
}

// NewPluginBuilder returns a builder for building plugins.
func NewPluginBuilder(name string, srcs ...string) *Plugin {
	return &Plugin{
		name:    name,
		target:  build.Pathf("k3r-%s-plugin.so", name),
		sources: srcs,
	}
}

type Plugin struct {
	name    string
	target  string
	sources []string
}

func (p *Plugin) Targets() []string {
	return []string{p.target}
}

func (p *Plugin) command() *exec.Cmd {
	cmd := "$CXX $CXXFLAGS -shared -fPIC"
	args := build.FieldsExpandWithDefault(cmd, DefaultEnv)
	args = append(args, "-o", p.target)
	args = append(args, p.sources...)
	args = append(args, build.FieldsExpandWithDefault("$LDFLAGS $EXTRA_LDFLAGS -lk3-plugin", DefaultEnv)...)
	return build.CommandWithEnv(DefaultEnv, args...)
}

func (p *Plugin) Build() error {
	b, err := build.NeedsRebuild(p.target, p.sources...)
	if b && err == nil {
		cmd := p.command()
		log.Verboseln("+", cmd.String())
		err = cmd.Run()
	}
	return err
}

func (p *Plugin) Commands() []compdb.Command {
	cmd := p.command()
	var ret []compdb.Command
	for _, src := range p.sources {
		ret = append(ret, compdb.Command{
			Command: cmd.String(),
			File:    src,
			Output:  p.target,
		})
	}
	return ret
}

func NewT3XFBuilder(name string, srcs ...string) *T3XF {
	return &T3XF{
		name:    name,
		target:  build.Pathf("%s.t3xf", name),
		sources: srcs,
	}
}

// T3XF is a T3XF builder.
type T3XF struct {
	name    string
	target  string
	sources []string
}

func (t *T3XF) Targets() []string {
	return []string{t.target}
}

func (t *T3XF) Sources() []string {
	return t.sources
}

func (t *T3XF) command() *exec.Cmd {
	args := []string{Compiler(), "-o", t.target}
	if env := build.FieldsExpandWithDefault("$K3CFLAGS", DefaultEnv); env != nil {
		args = append(args, env...)
	}
	visited := make(map[string]bool)
	for _, src := range t.sources {
		dir := filepath.Dir(src)
		if !visited[dir] {
			args = append(args, "-I", dir)
			visited[dir] = true
		}
	}
	for _, dir := range FindAuxiliaryDirectories() {
		args = append(args, "-I", dir)
	}
	args = append(args, t.sources...)
	return build.CommandWithEnv(DefaultEnv, args...)
}

func (t *T3XF) Build() error {
	b, err := build.NeedsRebuild(t.target, t.sources...)
	if b && err == nil {
		// T3XF file should be removed before building to force a new inode.
		// This ensures that already open T3XF files stay intact.
		os.Remove(t.target)

		cmd := t.command()
		log.Verboseln("+", cmd.String())
		err = cmd.Run()
	}
	return err
}

func (t *T3XF) Commands() []compdb.Command {
	cmd := t.command()
	var ret []compdb.Command
	for _, src := range t.sources {
		ret = append(ret, compdb.Command{
			Command: cmd.String(),
			File:    src,
			Output:  t.target,
		})
	}
	return ret
}

func NewASN1Codec(name string, encoding string, srcs ...string) *ASN1Codec {
	return &ASN1Codec{
		name:     name,
		encoding: encoding,
		c:        build.Pathf("%s.enc.c", name),
		h:        build.Pathf("%s.enc.h", name),
		lib:      build.Pathf("%slib.so", name),
		mod:      build.Pathf("%smod.ttcn3", name),
		sources:  srcs,
	}
}

type ASN1Codec struct {
	name           string
	encoding       string
	c, h, lib, mod string
	sources        []string
}

func (c *ASN1Codec) Targets() []string {
	return []string{c.lib, c.mod}
}

func (c *ASN1Codec) generateCommand() *exec.Cmd {
	args := build.FieldsExpandWithDefault("$ASN1C", DefaultEnv)
	args = append(args, fmt.Sprintf("-%s", c.encoding))
	args = append(args, build.FieldsExpandWithDefault("$ASN1CFLAGS", DefaultEnv)...)
	args = append(args, "-output", strings.TrimSuffix(c.c, ".c"), "-prefix", strings.TrimSuffix(c.c, ".enc.c"))
	args = append(args, "$OSSINFO/asn1dflt.linux-x86_64")
	args = append(args, c.sources...)
	return build.CommandWithEnv(DefaultEnv, args...)
}

func (c *ASN1Codec) buildCommand() *exec.Cmd {
	args := []string{"$CC"}
	args = append(args, "-fPIC", "-shared")
	args = append(args, "-D_OSSGETHEADER", "-DOSSPRINT")
	if env := build.FieldsExpandWithDefault("$CFLAGS", DefaultEnv); env != nil {
		args = append(args, env...)
	}
	if env := build.FieldsExpandWithDefault("$LDFLAGS", DefaultEnv); env != nil {
		args = append(args, env...)
	}
	if env := build.FieldsExpandWithDefault("$EXTRA_LDFLAGS", DefaultEnv); env != nil {
		args = append(args, env...)
	}
	args = append(args, c.c, c.h)
	args = append(args, "-l:libasn1code.a", "-Wl,-Bdynamic", "-o", c.lib)
	return build.CommandWithEnv(DefaultEnv, args...)
}

func (c *ASN1Codec) moduleCommand() *exec.Cmd {
	return build.CommandWithEnv(DefaultEnv, "$ASN2TTCN", "-o", c.mod, c.lib, fs.Stem(c.mod), c.encoding)
}

func (c *ASN1Codec) Build() error {
	updateC, err := build.NeedsRebuild(c.c, c.sources...)
	if err != nil {
		return err
	}
	updateH, err := build.NeedsRebuild(c.h, c.sources...)
	if err != nil {
		return err
	}
	if updateC || updateH {
		cmd := c.generateCommand()
		log.Verboseln("+", cmd.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	updateLib, err := build.NeedsRebuild(c.lib, c.c, c.h)
	if updateLib && err == nil {
		cmd := c.buildCommand()
		log.Verboseln("+", cmd.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	updateMod, err := build.NeedsRebuild(c.mod, c.lib)
	if updateMod && err == nil {
		cmd := c.moduleCommand()
		log.Verboseln("+", cmd.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return err
}

func (c *ASN1Codec) Commands() []compdb.Command {
	var ret []compdb.Command
	for _, src := range c.sources {
		ret = append(ret, compdb.Command{
			Command: c.generateCommand().String(),
			File:    src,
			Output:  c.c,
		}, compdb.Command{
			Command: c.generateCommand().String(),
			File:    src,
			Output:  c.h,
		})
	}

	ret = append(ret,
		compdb.Command{
			Command: c.buildCommand().String(),
			File:    c.c,
			Output:  c.lib,
		},
		compdb.Command{
			Command: c.buildCommand().String(),
			File:    c.h,
			Output:  c.lib,
		},
		compdb.Command{
			Command: c.moduleCommand().String(),
			File:    c.lib,
			Output:  c.mod,
		})

	return ret
}
