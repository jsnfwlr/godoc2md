package goreadme

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/WillAbides/godoc2md"
)

//go:generate ../bin/goreadme github.com/WillAbides/godoc2md/goreadme

var generatedRegexp = regexp.MustCompile(`<!--- generated by goreadme for (.*)-->`)

//WriteReadme writes a README.md for pkgname to the given path
func WriteReadme(pkgName, readmePath string) (err error) {
	f, err := os.OpenFile(readmePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640) //nolint:gosec
	if err != nil {
		return
	}
	defer func() {
		err = f.Close()
	}()
	err = ReadmeMD(pkgName, f)
	return
}

//VerifyReadme checks that the file at readmePath has the correct content for pkgName
func VerifyReadme(pkgName, readmePath string) (bool, error) {
	var want bytes.Buffer
	err := ReadmeMD(pkgName, &want)
	if err != nil {
		return false, err
	}

	got, err := ioutil.ReadFile(readmePath) //nolint:gosec
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return false, err
	}
	ok := bytes.Equal(want.Bytes(), got)
	return ok, nil
}

//ReadmeMD writes readme content for the given package to writer
func ReadmeMD(pkgName string, writer io.Writer) error {
	var buf bytes.Buffer
	config := &godoc2md.Config{
		TabWidth:          4,
		DeclLinks:         true,
		Goroot:            runtime.GOROOT(),
		SrcLinkHashFormat: "#L%d",
	}
	godoc2md.Godoc2md([]string{pkgName}, &buf, config)
	mdContent := buf.String()
	mdContent = strings.Replace(mdContent, `/src/target/`, `./`, -1)
	mdContent = strings.Replace(mdContent, fmt.Sprintf("/src/%s/", pkgName), `./`, -1)
	mdContent += fmt.Sprintf(`

<!--- generated by goreadme for %s-->
`, pkgName)

	_, err := writer.Write([]byte(mdContent))
	return err
}

func getReadmePackage(file []byte) (string, bool) {
	m := generatedRegexp.FindSubmatch(file)
	if len(m) == 0 {
		return "", false
	}
	return string(m[1]), true
}

//CheckReadmes checks that the readmes in basePath are up to date
// the first return value is a boolean indicating whether all readmes are up to date (false if any are outdated)
// the second return is the list of outdated readmes found
func CheckReadmes(basePath, readmeName string, excludeDirs []string) (bool, []string, error) {
	var paths []string
	readmes, err := FindReadmes(basePath, readmeName, excludeDirs)
	if err != nil {
		return false, nil, err
	}
	for path, pkg := range readmes {
		ok, err := VerifyReadme(pkg, path)
		if err != nil {
			return false, nil, err
		}
		if !ok {
			paths = append(paths, path)
		}
	}
	ok := len(paths) == 0
	return ok, paths, nil
}

//FindReadmes finds all goreadme generated README.md files in basePath and child directories
//  returns a map of filepath:package
func FindReadmes(basePath, readmeName string, excludeDirs []string) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			var abs string
			abs, err = filepath.Abs(path)
			if err != nil {
				return err
			}
			for _, exDir := range excludeDirs {
				var exAbs string
				exAbs, err = filepath.Abs(exDir)
				if err != nil {
					return err
				}
				if exAbs == abs {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.EqualFold(info.Name(), readmeName) {
			return nil
		}
		content, err := ioutil.ReadFile(path) //nolint:gosec
		if err != nil {
			return err
		}
		pkg, ok := getReadmePackage(content)
		if !ok {
			return nil
		}
		files[path] = pkg
		return nil
	})
	return files, err
}
