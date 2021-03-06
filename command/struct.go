package command

import (
	"time"
)

const (
	DEFAULT_PREFIX = "/home/r"
	WORKDIR        = "/var/rogi"
	DATAFILE       = WORKDIR + "/rogi.db"
	SETTING_BUCKET = "setting"
)

var (
	_ENV  []string
	_HOME string
)

type PackageObject struct {
	Name      string          `toml:"name"`
	Version   string          `toml:"version"`
	LongDesc  string          `toml:"long_desc"`
	ShortDesc string          `toml:"short_desc"`
	Custodian string          `toml:"custodian"`
	Homepage  string          `toml:"homepage,omitempty"`
	Scripts   ScriptObject    `toml:"scripts,omitempty"`
	Settings  *SettingObjects `toml:"settings,omitempty"`
	Owner     string          `toml:"owner,omitempty"`
	Group     string          `toml:"group,omitempty"`
	Mode      string          `toml:"mode,omitempty"`
	Files     []string        `toml:"files,omitempty"`
	FDetail   *FileObjects    `toml:"fdetail,omitempty"`
}

type ScriptObject struct {
	Start       string `toml:"start,omitempty"`
	PreInstall  string `toml:"preinstall,omitempty"`
	PostInstall string `toml:"postinstall,omitempty"`
	Stop        string `toml:"stop,omitempty"`
}

type StateObject struct {
	Packages map[string]string                 `toml:"packages"`
	Settings map[string]map[string]interface{} `toml:"settings"`
}

type FileType int
type TemplateType int

const (
	FTDir = iota
	FTFile
	FTBinary
	FTGlob
	FTConfig
)

const (
	TTNone = iota
	TTTemplate
	TTExpand
)

type FileObject struct {
	FTType  FileType
	TType   TemplateType
	Mode    int64
	Size    int64
	ModTime time.Time
	Uname   string
	Gname   string
	Target  string
	Source  string
}

type FileObjects []FileObject

type SettingObject struct {
	Package string      `db:"pkg"`
	Key     string      `db:"key"`
	Value   interface{} `db:"value"`
}

type settings map[string]string

type SettingObjects []SettingObject
