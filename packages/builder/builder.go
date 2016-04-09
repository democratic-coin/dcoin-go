// builder
package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"net/http"
	"runtime"
	"io"
	"path/filepath"
	"os/exec"
	"strings"
)

const (
	GITPATH = `github.com/democratic-coin/dcoin-go`
)

var (
    options Settings
)

type Settings struct {
	Branch    string 
	GitRoot   string 
	TempPath  string
	OutFile   string
	GoPath   string
	BinData   string
	BinDebug  bool
}

func exit( err error ) {  
	if err != nil {
		fmt.Println( err )
	}
	fmt.Println( `Press Enter to exit...`)
    fmt.Scanln( )
	if err != nil {
		os.Exit(1)
	}
}

func download( zfile string ) ( destfile string ) {
	srcfile := fmt.Sprintf("%s/archive/%s", options.GitRoot, zfile )
	fmt.Println(`Downloading `, srcfile )
	destfile = filepath.Join( options.TempPath, zfile )
	out, err := os.Create( destfile )
  	if err != nil  {
  		exit( err )
    }
  	defer out.Close()
	check := http.Client{
                 CheckRedirect: func(r *http.Request, via []*http.Request) error {
                         r.URL.Opaque = r.URL.Path
                         return nil },
     }
    resp, err := check.Get( srcfile )
  	if err != nil {
  		exit(err)
  	}
	defer resp.Body.Close()
	  _, err = io.Copy(out, resp.Body)
  	if err != nil  {
  		exit(err)
  	}
	fmt.Println(`Downloaded successfully` )
	return
}

func  extract(f *zip.File) error {
    rc, err := f.Open()
    if err != nil {
        return err
    }
    defer rc.Close()
	fname := f.Name[strings.IndexRune( f.Name, '/' ) + 1:]
    path := filepath.Join( options.GoPath, GITPATH, fname )
	fmt.Println(`Decompressing`, fname )
    if f.FileInfo().IsDir() {
        return os.MkdirAll(path, f.Mode())
    } else {
        f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            return err
        }
        defer f.Close()
		
        _, err = io.Copy(f, rc)
        if err != nil {
            return err
        }
    }
    return nil
}

func main() {
	var ( settings map[string]Settings
	)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		exit(err)
	}
	params, err := ioutil.ReadFile(filepath.Join(dir, `builder.json`))
	if err != nil {
		exit(err)
	}
	if err = json.Unmarshal(params, &settings); err != nil {
		exit(err)
	}
	options = settings[`default`]
	if len(os.Args ) > 1 {
		if cmdopt, found := settings[ os.Args[1]]; found {
			if len( cmdopt.Branch )> 0 {
				options.Branch = cmdopt.Branch
			}
		} else {
			exit( fmt.Errorf( `Cannot find %s settings`, os.Args[1]))
		}
	}
	srcPath := filepath.Join( options.GoPath, GITPATH )
	if err = os.MkdirAll( srcPath, 0755); err != nil {
		exit(err)
	}
	zfile := options.Branch + `.zip`
	srcfile := download( zfile )
	
	z, err := zip.OpenReader(srcfile)
	if err != nil {
		exit( err )
	}
	defer z.Close()
	if _, err := os.Stat( filepath.Join( srcPath, "dcoinwindows.go")); err == nil {
		fmt.Println(`Removing `, srcPath )		
		if err = os.RemoveAll( srcPath); err!=nil {
			exit(err)
		}
	}
	
	for _, f := range z.File {
		if err = extract( f ); err != nil {
			exit( err )
		}
	}
	if err = os.Chdir( srcPath ); err != nil {
		exit( err )
	}
	fmt.Println(`Creating static.go`)
	args := []string{ `-o=packages/static/static.go`, `-pkg=static\` }
	if options.BinDebug {
		args = append( args, `-debug=true`)
	}
	cmd := exec.Command( options.BinData, append( args, `static/...` )...)
	if err := cmd.Run(); err != nil {
		exit( err )
	}
	fmt.Println(`Compiling dcoin.go`)
	if err = os.MkdirAll( filepath.Dir(options.OutFile), 0755); err != nil {
		exit(err)
	}
	args = []string{ `build`, `-o`, options.OutFile, `-ldflags` }
	if runtime.GOOS == `windows` {
		args = append( args, `-H windowsgui`)
	}
	cmd = exec.Command( `go`, append( args, GITPATH )... )
	if err = cmd.Run(); err != nil {
		exit( err )
	}
	exit(nil)
}
