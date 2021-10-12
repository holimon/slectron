package slectron

import "path/filepath"

type Paths struct {
	basicPath     string
	cachePath     string
	vendorPath    string
	downloadSrc   string
	downloadDst   string
	electronUnzip string
	assetsPath    string
	runtimePath   string
}

func newPaths(os, arch string, opt Options) *Paths {
	var p = &Paths{}
	p.downloadSrc = electronDownloadSrc(os, arch, opt.ElectronVersion, opt.ElectronMirror)
	p.basicPath = string(opt.BasicPath)
	p.cachePath = filepath.Join(p.basicPath, "cache")
	p.downloadDst = filepath.Join(p.cachePath, "electron.zip")

	p.vendorPath = filepath.Join(p.basicPath, "vendor")
	p.electronUnzip = filepath.Join(p.vendorPath, "electron")
	p.runtimePath = filepath.Join(p.electronUnzip, "electron")
	p.assetsPath = filepath.Join(p.vendorPath, "assets")

	return p
}
