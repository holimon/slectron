package slectron

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	validMap = map[string][]string{
		"darwin":  {"arm64", "amd64"},
		"linux":   {"arm64", "amd64", "arm"},
		"windows": {"arm64", "amd64"},
	}
)

func executalePath() string {
	app, _ := exec.LookPath(os.Args[0])
	abs, _ := filepath.Abs(app)
	path, _ := filepath.Split(abs)
	return path
}

func terminalPath() string {
	path, _ := os.Getwd()
	return path
}

func tempPath() string {
	if runtime.GOOS == "linux" {
		if u, e := user.Current(); e == nil {
			return filepath.Join(u.HomeDir, ".tmp-slectron")
		}
		return filepath.Join(os.TempDir(), ".tmp-slectron")
	}
	return filepath.Join(os.TempDir(), ".tmp-slectron")
}

func electronDownloadSrc(os, arch, version, mirror string) string {
	var o string
	switch strings.ToLower(os) {
	case "darwin":
		o = "darwin"
	case "linux":
		o = "linux"
	case "windows":
		o = "win32"
	}
	var a = "ia32"
	if strings.ToLower(arch) == "amd64" {
		a = "x64"
	} else if strings.ToLower(arch) == "arm" && o == "linux" {
		a = "armv7l"
	} else if strings.ToLower(arch) == "arm64" {
		a = "arm64"
	}
	return fmt.Sprintf("%s/v%s/electron-v%s-%s-%s.zip", mirror, version, version, o, a)
}

type writeCounter struct {
	content uint64
	loaded  uint64
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	l := len(p)
	wc.loaded += uint64(l)
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %.2f %% complete", float64(wc.loaded)/float64(wc.content)*100)
	if wc.content == wc.loaded {
		fmt.Printf("\n")
	}
	return l, nil
}

func httpDownload(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		if e := os.Remove(dst); e != nil {
			return e
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stating %s failed: %w", dst, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), os.FileMode(0755)); err != nil {
		return err
	}
	client := http.Client{Transport: &http.Transport{Dial: func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, time.Second*5)
	}, ResponseHeaderTimeout: time.Second * 5}}
	resp, err := client.Get(src)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	counter := &writeCounter{content: uint64(resp.ContentLength)}
	if _, err = io.Copy(f, io.TeeReader(resp.Body, counter)); err != nil {
		return err
	}
	return nil
}

func zipDecompress(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		if e := os.RemoveAll(dst); e != nil {
			return e
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stating %s failed: %w", dst, err)
	}
	z, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer z.Close()
	for _, item := range z.File {
		if item.FileInfo().IsDir() {
			if err := os.MkdirAll(filepath.Join(dst, item.Name), os.ModePerm); err != nil {
				return err
			}
		} else {
			err := os.MkdirAll(filepath.Dir(filepath.Join(dst, item.Name)), os.ModePerm)
			if err != nil {
				return err
			}
			fsrc, err := item.Open()
			if err != nil {
				return err
			}
			defer fsrc.Close()
			fdst, err := os.OpenFile(filepath.Join(dst, item.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, item.Mode())
			if err != nil {
				return err
			}
			defer fdst.Close()
			if _, err = io.Copy(fdst, fsrc); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeFile(b []byte, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return err
	}
	return ioutil.WriteFile(dst, b, os.ModePerm)
}

func isValid(os, arch string) bool {
	if o, b := validMap[runtime.GOOS]; b {
		for _, v := range o {
			if v == runtime.GOARCH {
				return true
			}
		}
	}
	return false
}

func vendorVerify(execute string, version string) bool {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, execute, "-v")
	if out, err := cmd.CombinedOutput(); err == nil {
		runVer := strings.ReplaceAll(string(out), "\r", "")
		runVer = strings.ReplaceAll(runVer, "\n", "")
		if strings.Compare(runVer, "v"+version) >= 0 {
			fmt.Printf("Electron %s already exists in the vendor.\n", runVer)
			return true
		}
	}
	return false
}
