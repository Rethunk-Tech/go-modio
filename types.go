package modio

import (
	"cmp"
	"strconv"
	"strings"
)

// File is one uploaded build of a mod.
type File struct {
	ID        int    `json:"id"`
	Filesize  int64  `json:"filesize"`
	Filename  string `json:"filename"`
	Version   string `json:"version"`
	Changelog string `json:"changelog"`
	Filehash  struct {
		MD5 string `json:"md5"`
	} `json:"filehash"`
	Download struct {
		// BinaryURL can expire: read the file again right before downloading.
		BinaryURL string `json:"binary_url"`
	} `json:"download"`
}

// VersionName is the file's own version, or its id when the author left it blank.
func (f File) VersionName() string { return cmp.Or(strings.TrimSpace(f.Version), strconv.Itoa(f.ID)) }

// MD5 is the lowercase hex digest the API publishes, its only hash.
func (f File) MD5() string { return strings.ToLower(f.Filehash.MD5) }

// Downloadable reports whether the file carries an address.
func (f File) Downloadable() bool { return f.ID != 0 && f.Download.BinaryURL != "" }

// Mod is one mod listing.
type Mod struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Summary        string `json:"summary"`
	Description    string `json:"description"`
	ProfileURL     string `json:"profile_url"`
	DateUpdated    int64  `json:"date_updated"`
	MaturityOption int    `json:"maturity_option"`
	SubmittedBy    struct {
		Username string `json:"username"`
	} `json:"submitted_by"`
	Logo struct {
		Thumb string `json:"thumb_320x180"`
	} `json:"logo"`
	Stats struct {
		Downloads       int `json:"downloads_total"`
		RatingsPositive int `json:"ratings_positive"`
	} `json:"stats"`
	Tags []struct {
		Name string `json:"name"`
	} `json:"tags"`
	// Modfile is the current file.
	Modfile File `json:"modfile"`
}

// Explicit reports the mod's maturity option includes explicit content.
func (m Mod) Explicit() bool { return m.MaturityOption&MaturityExplicit != 0 }

// ModList is one page of mods.
type ModList struct {
	Data  []Mod `json:"data"`
	Total int   `json:"result_total"`
}

// User is the account a token belongs to.
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}
