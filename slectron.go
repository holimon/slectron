package slectron

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	DefaultElectronVersion = "11.0.0"
	DefaultElectronMirror  = "http://npm.taobao.org/mirrors/electron"
)

type BasicDirType string

var DefaultBasicExec = BasicDirType(executalePath())
var DefaultBasicTerm = BasicDirType(terminalPath())
var DefaultBasicTemp = BasicDirType(tempPath())

type Options struct {
	ElectronVersion string
	ElectronMirror  string
	BasicPath       BasicDirType
	ElectronParam   string
	CustomVendorer  Vendorer
}

type Slectron struct {
	opt      Options
	paths    *Paths
	cwdwatch *cmdWatch
}

func New(opt Options) (s *Slectron, err error) {
	if !isValid(runtime.GOOS, runtime.GOARCH) {
		fmt.Printf("%s %s is invalid\n", runtime.GOOS, runtime.GOARCH)
		return nil, fmt.Errorf("%s %s is invalid", runtime.GOOS, runtime.GOARCH)
	}
	if opt.ElectronMirror == "" {
		opt.ElectronMirror = DefaultElectronMirror
	}
	if opt.ElectronVersion == "" {
		opt.ElectronVersion = DefaultElectronVersion
	}
	if opt.BasicPath == "" {
		opt.BasicPath = DefaultBasicTemp
	}
	s = &Slectron{opt: opt, paths: newPaths(runtime.GOOS, runtime.GOARCH, opt), cwdwatch: newcwdWatch()}
	//verify electron vendor and creat it.
	if vendorVerify(s.paths.runtimePath, opt.ElectronVersion) {
		return s, nil
	}
	if vendorVerify("electron", opt.ElectronVersion) {
		return s, nil
	}
	if s.opt.CustomVendorer != nil {
		if s.customVendor(s.opt.CustomVendorer) == nil {
			return s, nil
		}
	}
	return s, s.defaultVendor()
}

func (s *Slectron) Start() error {
	if s.opt.ElectronParam == "" {
		return fmt.Errorf("electron execution parameters have not been setted")
	}
	s.cwdwatch.cmdRun(s.paths.runtimePath, s.opt.ElectronParam)
	return nil
}

func (s *Slectron) Stop() {
	if s == nil {
		return
	}
	s.cwdwatch.cmdStop()
}

func (s *Slectron) Wait() {
	if s == nil {
		return
	}
	s.cwdwatch.cmdWait()
}

func (s *Slectron) AssetsWrite(embed func() (name string, content []byte, err error)) error {
	if s == nil {
		return fmt.Errorf("nil instance")
	}
	f, b, e := embed()
	dst := filepath.Join(s.paths.assetsPath, f)
	if e == nil {
		return writeFile(b, dst)
	}
	return fmt.Errorf("write bytes to %s failed", dst)
}

func (s *Slectron) AssetsQuote(name string) (string, error) {
	if s == nil {
		return "", fmt.Errorf("nil instance")
	}
	abs := filepath.Join(s.paths.assetsPath, name)
	if _, err := os.Stat(abs); err == nil {
		return abs, nil
	} else {
		return "", err
	}
}

func (s *Slectron) SetExecuteArgs(args ...string) error {
	if s == nil {
		return fmt.Errorf("nil instance")
	}
	if s.cwdwatch.isRun() {
		return fmt.Errorf("after Start method")
	}
	s.opt.ElectronParam = ""
	for _, a := range args {
		s.opt.ElectronParam += a
	}
	return nil
}
