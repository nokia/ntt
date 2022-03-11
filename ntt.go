package ntt

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/nokia/ntt/build"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/doc"
	"github.com/spf13/pflag"
)

// BasketFlags returns a flagset with all flags for filtering objects.
//
// BasketFlags are regular expressions to filter objects. If you pass multiple regular
// expressions, all of them must match (AND). Example:
//
// 	$ cat example.ttcn3
// 	testcase foo() ...
// 	testcase bar() ...
// 	testcase foobar() ...
// 	...
//
// 	$ ntt list --regex=foo --regex=bar
// 	example.foobar
//
// 	$ ntt list --regex='foo|bar'
// 	example.foo
// 	example.bar
// 	example.foobar
//
//
// Similarly, you can also specify regular expressions for documentation tags.
// Example:
//
// 	$ cat example.ttcn3
// 	// @one
// 	// @two some-value
// 	testcase foo() ...
//
// 	// @two: some-other-value
// 	testcase bar() ...
// 	...
//
// 	$ ntt list --tags-regex=@one --tags-regex=@two
// 	example.foo
//
// 	$ ntt list --tags-regex='@two: some'
// 	example.foo
// 	example.bar
//
func BasketFlags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("basket", pflag.ContinueOnError)
	fs.StringSliceP("regex", "r", nil, "list objects matching regular * expression.")
	fs.StringSliceP("exclude", "x", nil, "exclude objects matching regular * expresion.")
	fs.StringSliceP("tags-regex", "R", nil, "list objects with tags matching regular * expression")
	fs.StringSliceP("tags-exclude", "X", nil, "exclude objects with tags matching * regular expression")
	return fs
}

// A Basket is a filter for objects. It can be used to filter objects by name
// and tags.
//
// Baskets are also filters defined by environment variables of the form:
//
//         NTT_LIST_BASKETS_<name> = <filters>
//
// For example, to define a basket "stable" which excludes all objects with @wip
// or @flaky tags:
//
// 	export NTT_LIST_BASKETS_stable="-X @wip|@flaky"
//
// Baskets become active when they are listed in colon separated environment
// variable NTT_LIST_BASKETS. If you specify multiple baskets, at least of them
// must match (OR).
//
// Rule of thumb: all baskets are ORed, all explicit filter options are ANDed.
// Example:
//
// 	$ export NTT_LIST_BASKETS_stable="--tags-exclude @wip|@flaky"
// 	$ export NTT_LIST_BASKETS_ipv6="--tags-regex @ipv6"
// 	$ NTT_LIST_BASKETS=stable:ipv6 ntt list -R @flaky
//
//
// Above example will output all tests with a @flaky tag and either @wip or @ipv6 tag.
//
// If a basket is not defined by an environment variable, it's equivalent to a
// "--tags-regex" filter. For example, to lists all tests, which have either a
// @flaky or a @wip tag:
//
// 	# Note, flaky and wip baskets are not specified explicitly.
// 	$ NTT_LIST_BASKETS=flaky:wip ntt list
//
// 	# This does the same:
// 	$ ntt list --tags-regex="@wip|@flaky"
//
type Basket struct {
	// Name is the name of the basket. The basket is used to filter objects
	// by tag, if no explicit filters are given.
	Name string

	// Regular expressions the object name must match.
	NameRegex []string

	// Regular expressions the object name must not match.
	NameExclude []string

	// Regular expressions the object tags must match.
	TagsRegex []string

	// Regular expressions the object tags must not match.
	TagsExclude []string

	// Baskets are sub-baskets to be ORed.
	Baskets []Basket
}

// NewBasket creates a new basket and parses the given arguments.
func NewBasket(name string, args ...string) (Basket, error) {

	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	fs.AddFlagSet(BasketFlags())
	if err := fs.Parse(args); err != nil {
		return Basket{}, err
	}
	return NewBasketWithFlags(name, fs)
}

func NewBasketWithFlags(name string, fs *pflag.FlagSet) (Basket, error) {
	b := Basket{Name: name}

	var err error

	b.NameRegex, err = fs.GetStringSlice("regex")
	if err != nil {
		return b, err
	}

	b.NameExclude, err = fs.GetStringSlice("exclude")
	if err != nil {
		return b, err
	}
	b.TagsRegex, err = fs.GetStringSlice("tags-regex")
	if err != nil {
		return b, err
	}
	b.TagsExclude, err = fs.GetStringSlice("tags-exclude")
	if err != nil {
		return b, err
	}
	return b, nil
}

// Load baskets from given environment variable.
func (b *Basket) LoadFromEnv(key string) error {
	s := env.Getenv(key)
	if s == "" {
		return nil
	}

	for _, name := range strings.Split(s, ":") {
		// Ignore empty fields
		if name == "" {
			continue
		}
		args := strings.Fields(env.Getenv(fmt.Sprintf("%s_%s", key, name)))
		if len(args) == 0 {
			args = []string{"-R", "@" + name}
		}

		sb, err := NewBasket(name, args...)
		if err != nil {
			return err
		}
		b.Baskets = append(b.Baskets, sb)
	}
	return nil
}

// Match returns true if the given name and tags match the basket or sub-basket filters.
func (b *Basket) Match(name string, tags [][]string) bool {
	ok := b.match(name, tags)
	if len(b.Baskets) == 0 {
		return ok
	}

	for _, basket := range b.Baskets {
		if basket.Match(name, tags) && ok {
			return true
		}
	}
	return false
}

// match returns true if the given name and tags match the basket filters.
func (b *Basket) match(name string, tags [][]string) bool {
	if !b.matchAll(b.NameRegex, name) {
		return false
	}
	if len(b.NameExclude) > 0 && b.matchAll(b.NameExclude, name) {
		return false
	}

	if len(b.TagsRegex) > 0 {
		if len(tags) == 0 {
			return false
		}
		if !b.matchAllTags(b.TagsRegex, tags) {
			return false
		}
	}

	if len(b.TagsExclude) > 0 && b.matchAllTags(b.TagsExclude, tags) {
		return false
	}

	return true
}

// matchAll returns true if all regular expressions match the given string.
func (b *Basket) matchAll(regexes []string, s string) bool {
	for _, r := range regexes {
		if ok, _ := regexp.Match(r, []byte(s)); !ok {
			return false
		}
	}
	return true
}

// matchAllTags returns true if all regular expressions match the all given tags.
func (b *Basket) matchAllTags(regexes []string, tags [][]string) bool {
next:
	for _, r := range regexes {
		f := strings.SplitN(r, ":", 2)
		for i := range f {
			f[i] = strings.TrimSpace(f[i])
		}
		for _, tag := range tags {
			if ok, _ := regexp.Match(f[0], []byte(tag[0])); !ok {
				continue
			}
			if len(f) > 1 {
				if ok, _ := regexp.Match(f[1], []byte(tag[1])); !ok {
					continue
				}
			}
			continue next
		}
		return false
	}
	return true
}

// SplitQualifiedName splits a qualified name into module and test name.
func SplitQualifiedName(name string) (string, string) {
	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		return "", name
	}
	return parts[0], strings.Join(parts[1:], ".")
}

// NewSuite creates a new suite from the given files. It expects either
// a single directory as argument or a list of regular .ttcn3 files.
//
// Calling NewSuite with an empty argument list will create a suite from
// current working directory or, if set, from NTT_SOURCE_DIR.
//
// NewSuite will read manifest (package.yml) if any.
func NewSuite(files ...string) (*Suite, error) {
	oldSuite, err := ntt.NewFromArgs(files...)
	if err != nil {
		return nil, fmt.Errorf("loading test suite failed: %w", err)
	}

	name, err := oldSuite.Name()
	if err != nil {
		return nil, fmt.Errorf("retrieving test suite name failed: %w", err)
	}

	srcs, err := oldSuite.Sources()
	if err != nil {
		return nil, fmt.Errorf("retrieving TTCN-3 sources failed: %w", err)
	}

	imports, err := oldSuite.Imports()
	if err != nil {
		return nil, fmt.Errorf("retrieving TTCN-3 imports failed: %w", err)
	}

	var paths []string
	if s := env.Getenv("NTT_CACHE"); s != "" {
		paths = append(paths, strings.Split(s, ":")...)
	}
	paths = append(imports, k3.FindAuxiliaryDirectories()...)
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, cwd)
	}

	if err != nil {
		return nil, fmt.Errorf("retrieving runtime paths failed: %w", err)
	}

	t, err := oldSuite.Timeout()
	if err != nil {
		return nil, fmt.Errorf("retrieving test suite timeout failed: %w\n", err)
	}

	d, err := time.ParseDuration(fmt.Sprintf("%fs", t))
	if err != nil {
		return nil, fmt.Errorf("retrieving test suite timeout failed: %w\n", err)
	}

	var parametersFile string
	f, err := oldSuite.ParametersFile()
	if err != nil {
		return nil, fmt.Errorf("retrieving parameters file failed: %w", err)
	}
	if f != nil {
		parametersFile = f.Path()
	}

	dir, err := oldSuite.ParametersDir()
	if err != nil {
		return nil, fmt.Errorf("retrieving parameters file failed: %w", err)
	}
	v, err := oldSuite.Variables()
	if err != nil {
		return nil, fmt.Errorf("retrieving test suite variables failed: %w", err)
	}
	return &Suite{
		Vars:           v,
		Name:           name,
		Sources:        srcs,
		RuntimePaths:   paths,
		timeout:        d,
		parametersFile: parametersFile,
		parametersDir:  dir,
	}, nil

}

// Suite represents a test suite.
type Suite struct {
	Name         string
	Sources      []string
	RuntimePaths []string
	Vars         map[string]string

	timeout        time.Duration
	parametersFile string
	parametersDir  string
}

// Parameters returns the module parameters and timeout to be used for execution.
func (s *Suite) Parameters() (map[string]string, time.Duration, error) {
	return s.parameters("")
}

// TestParameters returns module module parameters and timeout to be used for a single test case.
func (s *Suite) TestParameters(name string) (map[string]string, time.Duration, error) {
	return s.parameters(name)
}

func (s *Suite) parameters(name string) (map[string]string, time.Duration, error) {
	m := make(map[string]string)

	if path := s.parametersFile; path != "" {
		if err := readParametersFile(path, &m); err != nil {
			return nil, 0, fmt.Errorf("reading %q failed: %w", path, err)
		}
		log.Debugf("Read parameters from %q\n", path)
	}

	if mod, test := SplitQualifiedName(name); mod != "" {
		testPars := filepath.Join(s.parametersDir, mod, test+".parameters")
		if err := readParametersFile(testPars, &m); err != nil {
			return nil, 0, fmt.Errorf("reading %qw failed: %w", testPars, err)
		}
		log.Debugf("Read test specific parameters from %q\n", testPars)
	}

	d := s.timeout
	if t, ok := m["TestcaseExecutor.time_out"]; ok {
		delete(m, "TestcaseExecutor.time_out")
		d2, err := time.ParseDuration(fmt.Sprintf("%ss", t))
		if err != nil {
			return nil, 0, err
		}
		d = d2
	}
	return m, d, nil
}

func readParametersFile(path string, v interface{}) error {
	b, err := fs.Content(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("reading parameters file failed: %w", err)
	}
	_, err = toml.Decode(string(b), v)
	if err != nil {
		return fmt.Errorf("decoding parameters file failed: %w", err)
	}
	return nil
}

// ErrNoSources is returned when no source files are found.
var ErrNoSources = fmt.Errorf("no sources")

// BuildProject builds a project with the given name. It the build of the
// project or any of its dependencies fails, the error is returned.
func BuildProject(name string, p project.Interface) error {
	start := time.Now()
	defer func() {
		log.Debugf("build %s took %s\n", name, time.Since(start))
	}()
	builders, err := PlanProject(name, p)
	if err != nil {
		return err
	}
	for _, b := range builders {
		if err := b.Build(); err != nil {
			return err
		}
	}
	return err
}

// PlanProject returns a list of builders required to build the given project.
//
// Using a different backend than k3 is currently not supported.
func PlanProject(name string, p project.Interface) ([]build.Builder, error) {
	var ret []build.Builder

	srcs, err := p.Sources()
	if err != nil {
		return nil, err
	}

	imports, err := p.Imports()
	if err != nil {
		return nil, err
	}

	for _, dir := range imports {
		builders, err := PlanImport(dir)
		if err != nil {
			return nil, err
		}
		for _, b := range builders {
			for _, o := range b.Targets() {
				if fs.HasTTCN3Extension(o) {
					srcs = append(srcs, o)
				}
			}

			// Skip T3XF Builders, because we'll build t3xf speparately.
			if t, ok := b.(*k3.T3XF); ok {
				srcs = append(srcs, t.Sources()...)
			} else {
				ret = append(ret, b)
			}
		}
	}

	if len(srcs) == 0 {
		return nil, fmt.Errorf("test root folder %s: %w", p.Root(), ErrNoSources)
	}

	ret = append(ret, k3.NewT3XFBuilder(name, srcs...))
	return ret, nil
}

// PlanImport returns a list of builders required to build the given import
// directory. An import directory can be a TTCN-3 library, a ASN.1 codec or a
// k3 plugin.
//
// Using a different backend than k3 is currently not supported.
func PlanImport(dir string) ([]build.Builder, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var (
		builders                      []build.Builder
		asn1Files, ttcn3Files, cFiles []string
		processed                     int
	)

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		switch path := filepath.Join(dir, f.Name()); filepath.Ext(path) {
		case ".asn1", ".asn":
			asn1Files = append(asn1Files, path)
			processed++
		case ".ttcn3", ".ttcn":
			ttcn3Files = append(ttcn3Files, path)
			processed++
		case ".c", ".cxx", ".cpp", ".cc":
			// Skip ASN1 codecs
			if strings.HasSuffix(path, ".enc.c") {
				continue
			}
			cFiles = append(cFiles, path)
			processed++
		}
	}
	if processed == 0 {
		return nil, fmt.Errorf("%s: %w", dir, ErrNoSources)
	}
	name := fs.Slugify(fs.Stem(dir))
	if len(asn1Files) > 0 {
		builders = append(builders, k3.NewASN1Codec(name, encoding(name), asn1Files...))
	}
	if len(cFiles) > 0 {
		builders = append(builders, k3.NewPluginBuilder(name, cFiles...))
	}
	if len(ttcn3Files) > 0 {
		builders = append(builders, k3.NewT3XFBuilder(name, ttcn3Files...))
	}
	return builders, nil
}

func encoding(name string) string {
	if strings.Contains(strings.ToLower(name), "rrc") {
		return "uper"
	}
	return "per"
}

// GenerateIDs emits given IDs to a channel.
func GenerateIDs(ids ...string) <-chan string {
	return GenerateIDsWithContext(context.Background(), ids...)
}

// GenerateIDs emits given IDs to a channel.
func GenerateIDsWithContext(ctx context.Context, ids ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, id := range ids {
			out <- id
		}
	}()
	return out
}

// GenerateTests emits all test ids from given TTCN-3 files to a channel.
func GenerateTests(files ...string) <-chan string {
	return GenerateTestsWithContext(context.Background(), Basket{}, files...)
}

// GenerateTestsWithContext emits all test ids from given TTCN-3 files to a channel.
func GenerateTestsWithContext(ctx context.Context, b Basket, files ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, src := range files {
			tree := ttcn3.ParseFile(src)
			for _, def := range tree.Funcs() {
				if n := def.Node.(*ast.FuncDecl); n.IsTest() {
					if !generate(ctx, def, b, out) {
						return
					}
				}
			}
		}
	}()
	return out
}

// GenerateControls emits all control function ids from given TTCN-3 files to a channel.
func GenerateControls(files ...string) <-chan string {
	return GenerateControlsWithContext(context.Background(), Basket{}, files...)
}

// GenerateControls emits all control function ids from given TTCN-3 files to a channel.
func GenerateControlsWithContext(ctx context.Context, b Basket, files ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, src := range files {
			tree := ttcn3.ParseFile(src)
			for _, n := range tree.Controls() {
				if !generate(ctx, n, b, out) {
					return
				}
			}

		}
	}()
	return out
}

func generate(ctx context.Context, def *ttcn3.Definition, b Basket, c chan<- string) bool {
	id := def.Tree.QualifiedName(def.Ident)
	tags := doc.FindAllTags(ast.FirstToken(def.Node).Comments())
	if b.Match(id, tags) {
		select {
		case c <- id:
		case <-ctx.Done():
			return false
		}
	}
	return true
}
