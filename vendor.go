package slectron

import (
	"fmt"
)

type Vendorer func() (content []byte, err error)

func (s *Slectron) defaultVendor() error {
	fmt.Println("Electron does not exist in the system or vendor directory. Now create the vendor directory.")
	fmt.Printf("Download from %s to %s\n", s.paths.downloadSrc, s.paths.downloadDst)
	if err := httpDownload(s.paths.downloadSrc, s.paths.downloadDst); err != nil {
		return fmt.Errorf("download from %s to %s failed", s.paths.downloadSrc, s.paths.downloadDst)
	}
	fmt.Printf("Unzip from %s to %s\n", s.paths.downloadDst, s.paths.electronUnzip)
	if err := zipDecompress(s.paths.downloadDst, s.paths.electronUnzip); err != nil {
		return fmt.Errorf("unzip from %s to %s failed", s.paths.downloadDst, s.paths.electronUnzip)
	}
	fmt.Println("Electron vendor directory creation completed.")
	return nil
}

func (s *Slectron) customVendor(embed func() (content []byte, err error)) error {
	fmt.Println("Electron is specified by the user as the source. Now try to create the vendor directory")
	b, e := embed()
	if e != nil {
		return e
	}
	fmt.Printf("Writting from bytes to %s\n", s.paths.downloadDst)
	if e = writeFile(b, s.paths.downloadDst); e != nil {
		return e
	}
	fmt.Printf("Unzip from %s to %s\n", s.paths.downloadDst, s.paths.electronUnzip)
	if e = zipDecompress(s.paths.downloadDst, s.paths.electronUnzip); e != nil {
		return e
	}
	if !vendorVerify(s.paths.runtimePath, s.opt.ElectronVersion) {
		fmt.Println("Electron version is too low or the executable is corrupted.")
		return fmt.Errorf("electron version is too low or the executable is corrupted")
	}
	fmt.Println("Electron vendor directory creation completed.")
	return nil
}
