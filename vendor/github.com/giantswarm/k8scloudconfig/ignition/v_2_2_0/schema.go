package ignition

// This schema structure is based on github.com/coreos/ignition/config/v2_2/types/schema.go
// Due to issue with unmarshalling embedded anonymous nested structures,
// this file removes such structures.
// Changed types: Directory, File, Link.
type CaReference struct {
	Source       string       `json:"source,omitempty" yaml:"source,omitempty"`
	Verification Verification `json:"verification,omitempty" yaml:"verification,omitempty"`
}
type Config struct {
	Ignition Ignition `json:"ignition" yaml:"ignition,omitempty"`
	Networkd Networkd `json:"networkd,omitempty" yaml:"networkd,omitempty"`
	Passwd   Passwd   `json:"passwd,omitempty" yaml:"passwd,omitempty"`
	Storage  Storage  `json:"storage,omitempty" yaml:"storage,omitempty"`
	Systemd  Systemd  `json:"systemd,omitempty" yaml:"systemd,omitempty"`
}
type ConfigReference struct {
	Source       string       `json:"source,omitempty" yaml:"source,omitempty"`
	Verification Verification `json:"verification,omitempty" yaml:"verification,omitempty"`
}
type Create struct {
	Force   bool           `json:"force,omitempty" yaml:"force,omitempty"`
	Options []CreateOption `json:"options,omitempty" yaml:"options,omitempty"`
}
type CreateOption string
type Device string
type Directory struct {
	Filesystem string     `json:"filesystem,omitempty" yaml:"filesystem,omitempty"`
	Group      *NodeGroup `json:"group,omitempty" yaml:"group,omitempty"`
	Mode       *int       `json:"mode,omitempty" yaml:"mode,omitempty"`
	Overwrite  *bool      `json:"overwrite,omitempty" yaml:"overwrite,omitempty"`
	Path       string     `json:"path,omitempty" yaml:"path,omitempty"`
	User       *NodeUser  `json:"user,omitempty" yaml:"user,omitempty"`
}
type Disk struct {
	Device     string      `json:"device,omitempty" yaml:"device,omitempty"`
	Partitions []Partition `json:"partitions,omitempty" yaml:"partitions,omitempty"`
	WipeTable  bool        `json:"wipeTable,omitempty" yaml:"wipeTable,omitempty"`
}
type File struct {
	Append     bool         `json:"append,omitempty" yaml:"append,omitempty"`
	Contents   FileContents `json:"contents,omitempty" yaml:"contents,omitempty"`
	Filesystem string       `json:"filesystem,omitempty" yaml:"filesystem,omitempty"`
	Mode       int          `json:"mode,omitempty" yaml:"mode,omitempty"`
	Group      *NodeGroup   `json:"group,omitempty" yaml:"group,omitempty"`
	Overwrite  *bool        `json:"overwrite,omitempty" yaml:"overwrite,omitempty"`
	Path       string       `json:"path,omitempty" yaml:"path,omitempty"`
	User       *NodeUser    `json:"user,omitempty" yaml:"user,omitempty"`
}
type FileContents struct {
	Compression  string       `json:"compression,omitempty" yaml:"compression,omitempty"`
	Source       string       `json:"source,omitempty" yaml:"source,omitempty"`
	Verification Verification `json:"verification,omitempty" yaml:"verification,omitempty"`
}
type Filesystem struct {
	Mount *Mount  `json:"mount,omitempty" yaml:"mount,omitempty"`
	Name  string  `json:"name,omitempty" yaml:"name,omitempty"`
	Path  *string `json:"path,omitempty" yaml:"path,omitempty"`
}
type Group string
type Ignition struct {
	Config   IgnitionConfig `json:"config,omitempty" yaml:"config,omitempty"`
	Security Security       `json:"security,omitempty" yaml:"security,omitempty"`
	Timeouts Timeouts       `json:"timeouts,omitempty" yaml:"timeouts,omitempty"`
	Version  string         `json:"version,omitempty" yaml:"version,omitempty"`
}
type IgnitionConfig struct {
	Append  []ConfigReference `json:"append,omitempty" yaml:"append,omitempty"`
	Replace *ConfigReference  `json:"replace,omitempty" yaml:"replace,omitempty"`
}
type Link struct {
	Filesystem string     `json:"filesystem,omitempty" yaml:"filesystem,omitempty"`
	Group      *NodeGroup `json:"group,omitempty" yaml:"group,omitempty"`
	Hard       bool       `json:"hard,omitempty" yaml:"hard,omitempty"`
	Overwrite  *bool      `json:"overwrite,omitempty" yaml:"overwrite,omitempty"`
	Path       string     `json:"path,omitempty" yaml:"path,omitempty"`
	Target     string     `json:"target,omitempty" yaml:"target,omitempty"`
	User       *NodeUser  `json:"user,omitempty" yaml:"user,omitempty"`
}
type Mount struct {
	Create         *Create       `json:"create,omitempty" yaml:"create,omitempty"`
	Device         string        `json:"device,omitempty" yaml:"device,omitempty"`
	Format         string        `json:"format,omitempty" yaml:"format,omitempty"`
	Label          *string       `json:"label,omitempty" yaml:"label,omitempty"`
	Options        []MountOption `json:"options,omitempty" yaml:"options,omitempty"`
	UUID           *string       `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	WipeFilesystem bool          `json:"wipeFilesystem,omitempty" yaml:"wipeFilesystem,omitempty"`
}
type MountOption string
type Networkd struct {
	Units []Networkdunit `json:"units,omitempty" yaml:"units,omitempty"`
}
type NetworkdDropin struct {
	Contents string `json:"contents,omitempty" yaml:"contents,omitempty"`
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
}
type Networkdunit struct {
	Contents string           `json:"contents,omitempty" yaml:"contents,omitempty"`
	Dropins  []NetworkdDropin `json:"dropins,omitempty" yaml:"dropins,omitempty"`
	Name     string           `json:"name,omitempty" yaml:"name,omitempty"`
}
type Node struct {
	Filesystem string     `json:"filesystem,omitempty" yaml:"filesystem,omitempty"`
	Group      *NodeGroup `json:"group,omitempty" yaml:"group,omitempty"`
	Overwrite  *bool      `json:"overwrite,omitempty" yaml:"overwrite,omitempty"`
	Path       string     `json:"path,omitempty" yaml:"path,omitempty"`
	User       *NodeUser  `json:"user,omitempty" yaml:"user,omitempty"`
}
type NodeGroup struct {
	ID   *int   `json:"id,omitempty" yaml:"id,omitempty"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}
type NodeUser struct {
	ID   *int   `json:"id,omitempty" yaml:"id,omitempty"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}
type Partition struct {
	GUID     string `json:"guid,omitempty" yaml:"guid,omitempty"`
	Label    string `json:"label,omitempty" yaml:"label,omitempty"`
	Number   int    `json:"number,omitempty" yaml:"number,omitempty"`
	Size     int    `json:"size,omitempty" yaml:"size,omitempty"`
	Start    int    `json:"start,omitempty" yaml:"start,omitempty"`
	TypeGUID string `json:"typeGuid,omitempty" yaml:"typeGUID,omitempty"`
}
type Passwd struct {
	Groups []PasswdGroup `json:"groups,omitempty" yaml:"groups,omitempty"`
	Users  []PasswdUser  `json:"users,omitempty" yaml:"users,omitempty"`
}
type PasswdGroup struct {
	Gid          *int   `json:"gid,omitempty" yaml:"gid,omitempty"`
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
	PasswordHash string `json:"passwordHash,omitempty" yaml:"passwordHash,omitempty"`
	System       bool   `json:"system,omitempty" yaml:"system,omitempty"`
}
type PasswdUser struct {
	Create            *Usercreate        `json:"create,omitempty" yaml:"create,omitempty"`
	Gecos             string             `json:"gecos,omitempty" yaml:"gecos,omitempty"`
	Groups            []Group            `json:"groups,omitempty" yaml:"groups,omitempty"`
	HomeDir           string             `json:"homeDir,omitempty" yaml:"homeDir,omitempty"`
	Name              string             `json:"name,omitempty" yaml:"name,omitempty"`
	NoCreateHome      bool               `json:"noCreateHome,omitempty" yaml:"noCreateHome,omitempty"`
	NoLogInit         bool               `json:"noLogInit,omitempty" yaml:"noLogInit,omitempty"`
	NoUserGroup       bool               `json:"noUserGroup,omitempty" yaml:"noUserGroup,omitempty"`
	PasswordHash      *string            `json:"passwordHash,omitempty" yaml:"passwordHash,omitempty"`
	PrimaryGroup      string             `json:"primaryGroup,omitempty" yaml:"primaryGroup,omitempty"`
	SSHAuthorizedKeys []SSHAuthorizedKey `json:"sshAuthorizedKeys,omitempty" yaml:"sshAuthorizedKeys,omitempty"`
	Shell             string             `json:"shell,omitempty" yaml:"shell,omitempty"`
	System            bool               `json:"system,omitempty" yaml:"system,omitempty"`
	UID               *int               `json:"uid,omitempty" yaml:"uid,omitempty"`
}
type Raid struct {
	Devices []Device     `json:"devices,omitempty" yaml:"devices,omitempty"`
	Level   string       `json:"level,omitempty" yaml:"level,omitempty"`
	Name    string       `json:"name,omitempty" yaml:"name,omitempty"`
	Options []RaidOption `json:"options,omitempty" yaml:"options,omitempty"`
	Spares  int          `json:"spares,omitempty" yaml:"spares,omitempty"`
}
type RaidOption string
type SSHAuthorizedKey string
type Security struct {
	TLS TLS `json:"tls,omitempty" yaml:"tls,omitempty"`
}
type Storage struct {
	Directories []Directory  `json:"directories,omitempty" yaml:"directories,omitempty"`
	Disks       []Disk       `json:"disks,omitempty" yaml:"disks,omitempty"`
	Files       []File       `json:"files,omitempty" yaml:"files,omitempty"`
	Filesystems []Filesystem `json:"filesystems,omitempty" yaml:"filesystems,omitempty"`
	Links       []Link       `json:"links,omitempty" yaml:"links,omitempty"`
	Raid        []Raid       `json:"raid,omitempty" yaml:"raid,omitempty"`
}
type Systemd struct {
	Units []Unit `json:"units,omitempty" yaml:"units,omitempty"`
}
type SystemdDropin struct {
	Contents string `json:"contents,omitempty" yaml:"contents,omitempty"`
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
}
type TLS struct {
	CertificateAuthorities []CaReference `json:"certificateAuthorities,omitempty" yaml:"certificateAuthorities,omitempty"`
}
type Timeouts struct {
	HTTPResponseHeaders *int `json:"httpResponseHeaders,omitempty" yaml:"httpResponseHeaders,omitempty"`
	HTTPTotal           *int `json:"httpTotal,omitempty" yaml:"httpTotal,omitempty"`
}
type Unit struct {
	Contents string          `json:"contents,omitempty" yaml:"contents,omitempty"`
	Dropins  []SystemdDropin `json:"dropins,omitempty" yaml:"dropins,omitempty"`
	Enable   bool            `json:"enable,omitempty" yaml:"enable,omitempty"`
	Enabled  *bool           `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Mask     bool            `json:"mask,omitempty" yaml:"mask,omitempty"`
	Name     string          `json:"name,omitempty" yaml:"name,omitempty"`
}
type Usercreate struct {
	Gecos        string            `json:"gecos,omitempty" yaml:"gecos,omitempty"`
	Groups       []UsercreateGroup `json:"groups,omitempty" yaml:"groups,omitempty"`
	HomeDir      string            `json:"homeDir,omitempty" yaml:"homeDir,omitempty"`
	NoCreateHome bool              `json:"noCreateHome,omitempty" yaml:"noCreateHome,omitempty"`
	NoLogInit    bool              `json:"noLogInit,omitempty" yaml:"noLogInit,omitempty"`
	NoUserGroup  bool              `json:"noUserGroup,omitempty" yaml:"noUserGroup,omitempty"`
	PrimaryGroup string            `json:"primaryGroup,omitempty" yaml:"primaryGroup,omitempty"`
	Shell        string            `json:"shell,omitempty" yaml:"shell,omitempty"`
	System       bool              `json:"system,omitempty" yaml:"system,omitempty"`
	UID          *int              `json:"uid,omitempty" yaml:"uid,omitempty"`
}
type UsercreateGroup string
type Verification struct {
	Hash *string `json:"hash,omitempty" yaml:"hash,omitempty"`
}
