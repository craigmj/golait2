package parser

import (
	`bufio`
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	`regexp`
	"strings"
	"text/template"

	"github.com/craigmj/commander"
	`github.com/golang/glog`
	`github.com/juju/errors`
)

// findTemplate finds the first template file matching the
// template name given.
func findTemplate(t string) (string, error) {
	home := os.Getenv("GOLAIT2_HOME")
	if "" == home {
		return "", errors.New("You need to set the GOLAIT2_HOME environment variable")
	}
	// First we search for a file with exactly the templates name
	templates := filepath.Join(home, "templates")
	file := filepath.Join(templates, t)
	stat, err := os.Stat(file)
	if nil == err && !stat.IsDir() {
		return file, nil
	}
	// If we don't find that, we search for any file matching the
	// templates name without file extension
	glob := filepath.Join(templates, t+".*")
	files, err := filepath.Glob(glob)
	if nil != err {
		return "", err
	}
	if 0 == len(files) {
		return "", errors.New("No template files matching " + glob)
	}
	if 1 < len(files) {
		return "", fmt.Errorf("Found %d files matching %s", len(files), glob)
	}
	return files[0], nil
}

func EnvGOPATH() string {
	return os.Getenv("GOPATH")
}

func hasGOPATH() bool {
	return ``!=EnvGOPATH()
}

// findFileToRoot searches from the dir to the root of the filesystem
// looking for a file named filename.
func findFileToRoot(dir, filename string) (string, error) {
	fmt.Println("findFileToRoot dir = ", dir, ", filename=", filename)
	_, err := os.Stat(filepath.Join(dir, filename))
	if nil==err {
		return filepath.Join(dir, filename), nil
	}
	if !os.IsNotExist(err) {
		return ``, errors.Trace(err)
	}
	if `.`==dir {
		wd, err := os.Getwd()
		if nil!=err {
			return ``, errors.Trace(err)
		}
		dir = wd
	}
	if `/`==dir {
		return ``, errors.Trace(err)
	}
	return findFileToRoot(filepath.Dir(dir), filename)
}

// toDir returns the input if the input is a directory, or the directory of the input
// if the input is a file.
func toDir(in string) (string) {
	var err error
	if !filepath.IsAbs(in) {
		in, err = filepath.Abs(in)
		if nil!=err {
			panic(errors.Trace(err))
		} 
	}
	s, err := os.Stat(in)
	if nil!=err {
		if os.IsNotExist(err) {	// assume we're looking at a file
			return filepath.Dir(in)
		}
		panic(fmt.Errorf("FAILED to stat %s in toDir: %w", in, err))
	}
	if s.IsDir() {
		return in
	}
	return filepath.Dir(in)
}
/** Gets the module name and source root for the go build from the dir */
func getModule(dir string) (string, string, error) {
	dir = toDir(dir)
	modFile, err := findFileToRoot(dir, "go.mod")
	if nil!=err {
		return ``, ``, errors.Trace(fmt.Errorf("Failed to find go.mod on tree from %s: %w", dir, err))
	}
	in, err := os.Open(modFile)
	if nil!=err {
		return ``, ``, fmt.Errorf("Failed to read %s: %w", modFile, err)
	}
	defer in.Close()
	modRegex := regexp.MustCompile(`^module\s+([^\s]+)\s*$`)
	scan := bufio.NewScanner(in)
	for scan.Scan() {
		if m := modRegex.FindStringSubmatch(scan.Text()); nil!=m {
			return m[1], filepath.Dir(modFile), nil
		}
		fmt.Println("NO MATCH: " , scan.Text())
	}
	return ``, ``, fmt.Errorf(`Failed to parse %s successfully to find module name`, modFile)
}

// GetModuleName returns the module name for the given path, either 
// from the src/ directory of the GOPATH, or from the module
// root directory with the module name appended
func GetModuleName(infile string) (string, error) {
	infile = toDir(infile)
	if !filepath.IsAbs(infile) {
		wd, err := os.Getwd()
		if nil!=err {
			return ``, err
		}
		infile = filepath.Clean(filepath.Join(wd, infile))
	}
	if hasGOPATH() {
		srcDir := filepath.Join(EnvGOPATH(), `src`)
		return filepath.Dir(infile)[len(srcDir)+1:], nil
	} else {
		modName, modDir, err := getModule(filepath.Dir(infile))
		if nil!=err {
			return ``, errors.Trace(err)
		}
		inDir := filepath.Dir(infile)
		return filepath.Join(modName, inDir[len(modDir)+1:]), nil
	}
}

func FileOnGOPATHSrc(in string) (string, error) {
	glog.Infof("FileOnGOPATHSrc(%s)", in)
	if hasGOPATH() {
		return filepath.Join(EnvGOPATH(), `src`, in), nil
	}
	_, modDir, err := getModule(in)
	return filepath.Join(modDir, in), errors.Trace(err)
}

func FileOnGOPATH(in string) (string, error) {
	if hasGOPATH() {
		return filepath.Join(EnvGOPATH(), in), nil
	}
	_, modDir, err := getModule(in)
	return filepath.Join(modDir, in), errors.Trace(err)
}

func GOPATH() string {
	gopath := EnvGOPATH()
	if ``!=gopath {
		return gopath
	}
	wd, err := os.Getwd()
	if nil!=err {
		panic(err)
	}
	_, modDir, err := getModule(wd)
	if nil!=err {
		panic(err)
	}
	return modDir
}

func PackageNameFromFilePath(f string) string {
	pkg := filepath.Base(filepath.Dir(f))
	if ``==pkg && !hasGOPATH() {
		modName, err := GetModuleName(f)
		if nil!=err {
			glog.Errorf("PackageNameFromFilePath(%s) failed: %w", f, err)
			panic(err)
		}
		parts := filepath.SplitList(modName)
		pkg = parts[len(parts)-1]
	}
	return pkg
}

func mustHave(arg, value string) error {
	if "" == arg {
		return errors.New("You need to provide a value for the -" + arg + " argument")
	}
	return nil
}

func oneError(e ...error) error {
	for _, err := range e {
		if nil != err {
			return err
		}
	}
	return nil
}

func FullPackageNameForFile(infile string) (string, error) {
	var err error
	infile, err = filepath.Abs(infile)
	if nil!=err {
		return ``, errors.Trace(err)
	}
	if hasGOPATH() {
		// If we've got a GOPATH, we make the input file relative to that,
		// then strip the last path element (the filename), and the first
		// relative path element ('src') and return the resulting package name
		gopath := GOPATH()
		fullpath, err := filepath.Rel(gopath, infile)
		if nil!=err {
			return ``, errors.Trace(err)
		}
		parts := filepath.SplitList(fullpath)
		return filepath.Join(parts[1:len(parts)-1]...), nil
	}
	// Without a GOPATH, we find the module name and module root directory.
	// Then we make the infile relative to the module root directory, and
	// return the module name prepended to the DIRECTORY of the input file.
	modName, modRootDir, err := getModule(infile)
	if nil!=err {
		return ``, errors.Trace(err)
	}
	rel, err := filepath.Rel(modRootDir, infile)
	if nil!=err {
		return ``, errors.Trace(err)
	}
	return filepath.Join(modName, filepath.Dir(rel)), nil
}

func GenerateTemplate(outfile string, out io.Writer, infile, rpcType, rpcConstructor, packageName, templateFile string, recoverFlag bool, connectionClass, connectionConstructor string, jsApply bool) error {
	if !recoverFlag {
		fmt.Println("GENERATING GOLAIT2 template from " + infile + " without RECOVERY - use for DEBUG builds only")
	}
	file, err := parser.ParseFile(token.NewFileSet(),
		infile, nil, 0)
	if nil != err {
		return err
	}

	rootPackagePath, err := FullPackageNameForFile(infile)
	if nil!=err {
		return errors.Trace(err)
	}
	glog.Infof("rootPackagePath(%s) = %s", infile, rootPackagePath)

	class, err := NewClassDefinition(file, rootPackagePath, packageName, rpcType, rpcConstructor, recoverFlag, connectionClass, connectionConstructor)
	if nil != err {
		return err
	}
	class.JsApply = jsApply

	templateFile, err = findTemplate(templateFile)
	if nil != err {
		return err
	}
	t, err := template.New("").Funcs(map[string]interface{}{
		"ExprToGoVar" : ExprToGoVar,
		"ExprToString" : ExprToString,
		"Join": func(glue string, in []string) string {
			return strings.Join(in, glue)
		},
		"Prefix": func(prefix string, in []string) []string {
			a := make([]string, len(in))
			for i, s := range in {
				a[i] = prefix + s
			}
			return a
		},
		"Outfile": func() string {
			return outfile
		},
		"Base": func(in string) string {
			return filepath.Base(in)
		},
		"StripExt": func(in string) string {
			s := filepath.Ext(in)
			if "" == s {
				return in
			}
			return in[0 : len(in)-len(s)]
		},
	}).ParseFiles(templateFile)
	if nil != err {
		return err
	}
	if err := t.ExecuteTemplate(out, filepath.Base(templateFile), class); nil != err {
		return err
	}
	return nil
}

func GenerateCommand() *commander.Command {
	fs := flag.NewFlagSet("gen", flag.ExitOnError)
	argOut := fs.String("out", "", "Output file for generation")
	argPkg := fs.String("pkg", "", "Output package (go) or class name (js, php, etc). Default to out file directory name.")
	argType := fs.String("type", "", "Type which is the basic struct for the API")
	argConstructor := fs.String("constructor", "", "Constructor method to get a *Type")
	argIn := fs.String("in", "", "Input file")
	argTemplate := fs.String("tem", "go", "Template file to use (no suffix required)")
	recoverFlag := fs.Bool("recover", true, "Auto-recover from PANICS (set to true for production)")
	jsApply := fs.Bool("jsApply", false, "Use .apply when calling return functions")

	connClass := fs.String("connectionClass", "", "Connection type to set in API base type - API base type must enclose the Connection")
	connConstructor := fs.String("connectionConstructor", "", "Connection method to create the API type")

	return commander.NewCommand(
		"gen",
		"Generate a template from the API struct",
		fs,
		func([]string) error {
			var out io.Writer
			out = os.Stdout

			if err := oneError(
				mustHave("type", *argType),
				mustHave("in", *argIn),
				mustHave("tem", *argTemplate),
			); nil != err {
				return err
			}

			packageName := *argPkg
			if "" == packageName {
				if "" != *argOut {
					packageName = PackageNameFromFilePath(*argOut)
				} else {
					packageName = "jsonrpc"
				}
			}

			if "" != *argOut {
				var err error
				outfile := *argOut
				// Ensure we have the output directory for the output file
				if err = os.MkdirAll(filepath.Dir(outfile), 0755); nil != err {
					return err
				}
				outf, err := os.Create(outfile)
				if nil != err {
					return err
				}
				defer outf.Close()
				out = outf
			}

			wd, err := os.Getwd()
			if nil!=err {
				panic(err)
			}
			inputFile, err := filepath.Abs(*argIn)
			if nil!=err {
				return errors.Trace(err)
			}
			glog.Infof("WORKING Directory = %s", wd)

			return GenerateTemplate(*argOut, out, inputFile, *argType, *argConstructor, packageName, *argTemplate, *recoverFlag, *connClass, *connConstructor, *jsApply)
		})
}
