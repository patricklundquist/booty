package components

import (
	"fmt"
	"os"
	"path/filepath"

	"go.amplifyedge.org/booty-v2/dep"
	"go.amplifyedge.org/booty-v2/internal/fileutil"

	"go.amplifyedge.org/booty-v2/internal/downloader"
	"go.amplifyedge.org/booty-v2/internal/osutil"
	"go.amplifyedge.org/booty-v2/internal/store"
)

const (
	// version -- version -- os_arch
	caddyUrlFormat = "https://github.com/caddyserver/caddy/releases/download/v%s/caddy_%s_%s.%s"
	// https://github.com/caddyserver/caddy/releases/download/v2.3.0/caddy_2.3.0_windows_amd64.zip
)

type Caddy struct {
	version string
	db      *store.DB
}

func NewCaddy(db *store.DB, version string) *Caddy {
	return &Caddy{
		version: version,
		db:      db,
	}
}

func (c *Caddy) Version() string {
	return c.version
}

func (c *Caddy) Name() string {
	return "caddy"
}

func (c *Caddy) Download() error {
	downloadDir := getDlPath(c.Name(), c.version)
	_ = os.MkdirAll(downloadDir, 0755)
	var osname string
	var ext string
	switch osutil.GetOS() {
	case "linux", "darwin":
		ext = "tar.gz"
		osname = fmt.Sprintf("%s_%s", osutil.GetAltOs(), osutil.GetArch())
	case "windows":
		ext = "zip"
		osname = fmt.Sprintf("%s_%s", osutil.GetOS(), osutil.GetArch())
	}
	fetchUrl := fmt.Sprintf(caddyUrlFormat, c.version, c.version, osname, ext)
	err := downloader.Download(fetchUrl, downloadDir)
	if err != nil {
		return err
	}
	return nil
}

func (c *Caddy) Install() error {
	var err error
	// install to global path
	binDir := osutil.GetBinDir()
	dlPath := getDlPath(c.Name(), c.version)

	// all files that are going to be installed
	executableName := c.Name()
	switch osutil.GetOS() {
	case "windows":
		executableName += ".exe"
	}
	filesMap := map[string][]interface{}{
		filepath.Join(dlPath, executableName): {filepath.Join(binDir, executableName), 0755},
	}
	ip := store.InstalledPackage{
		Name:     c.Name(),
		Version:  c.version,
		FilesMap: map[string]int{},
	}
	// copy file to the global bin directory
	for k, v := range filesMap {
		if err = fileutil.Copy(k, v[0].(string)); err != nil {
			return err
		}
		installedName := v[0].(string)
		installedMode := v[1].(int)
		if err = os.Chmod(installedName, os.FileMode(installedMode)); err != nil {
			return err
		}
		ip.FilesMap[installedName] = installedMode
	}
	if err = c.db.New(&ip); err != nil {
		return err
	}
	return os.RemoveAll(dlPath)
}

func (c *Caddy) Uninstall() error {
	var err error
	dlPath := getDlPath(c.Name(), c.version)
	// install to global path

	// all files that are going to be installed
	var pkg *store.InstalledPackage
	pkg, err = c.db.Get(c.Name())
	if err != nil {
		return err
	}
	var filesList []string
	for k := range pkg.FilesMap {
		filesList = append(filesList, k)
	}
	// uninstall listed files
	for _, file := range filesList {
		if err = os.RemoveAll(file); err != nil {
			return err
		}
	}
	// remove downloaded files
	return os.RemoveAll(dlPath)
}

func (c *Caddy) Update(version string) error {
	c.version = version
	if err := c.Uninstall(); err != nil {
		return err
	}
	if err := c.Download(); err != nil {
		return err
	}
	return c.Install()
}

func (c *Caddy) Run(args ...string) error {
	return nil
}

func (c *Caddy) Backup() error {
	return nil
}

func (c *Caddy) RunStop() error {
	return nil
}

func (c *Caddy) Dependencies() []dep.Component {
	return nil
}