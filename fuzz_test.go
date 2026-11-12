package modio

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

func FuzzMod(f *testing.F) {
	f.Add([]byte(timberAPI))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(`[]`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > maxBody {
			return
		}
		var m Mod
		if err := json.Unmarshal(raw, &m); err != nil {
			return
		}
		_ = m.Explicit()
		checkFile(t, m.Modfile)
		if _, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzFile(f *testing.F) {
	f.Add([]byte(`{"id":9,"filesize":2048,"filename":"t.zip","version":"","filehash":{"md5":"0123456789ABCDEF0123456789abcdef"},"download":{"binary_url":"https://g.modapi.io/f/9"}}`))
	f.Add([]byte(`{"id":9,"version":"0.8"}`))
	f.Add([]byte(`{"id":7}`))
	f.Add([]byte(`{}`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > maxBody {
			return
		}
		var file File
		if err := json.Unmarshal(raw, &file); err != nil {
			return
		}
		checkFile(t, file)
		if _, err := json.Marshal(file); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzModList(f *testing.F) {
	f.Add([]byte(`{"data":[` + timberAPI + `],"result_total":3002}`))
	f.Add([]byte(`{"data":[{"id":5,"name":"Harmony"}]}`))
	f.Add([]byte(`{}`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > maxBody {
			return
		}
		var list ModList
		if err := json.Unmarshal(raw, &list); err != nil {
			return
		}
		for i := range list.Data {
			_ = list.Data[i].Explicit()
			checkFile(t, list.Data[i].Modfile)
		}
		if _, err := json.Marshal(list); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzUser(f *testing.F) {
	f.Add([]byte(`{"id":7,"username":"nomad"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > maxBody {
			return
		}
		var u User
		if err := json.Unmarshal(raw, &u); err != nil {
			return
		}
		if !utf8.ValidString(u.Username) {
			t.Fatalf("username not utf-8")
		}
		if _, err := json.Marshal(u); err != nil {
			t.Fatal(err)
		}
	})
}

func checkFile(t *testing.T, file File) {
	t.Helper()
	if file.VersionName() == "" {
		t.Fatal("VersionName empty")
	}
	md5 := file.MD5()
	if md5 != strings.ToLower(file.Filehash.MD5) {
		t.Fatalf("MD5 %q != lower(%q)", md5, file.Filehash.MD5)
	}
	if file.Downloadable() {
		if file.ID == 0 {
			t.Fatal("Downloadable with id 0")
		}
		if len(md5) != md5HexLen {
			t.Fatalf("Downloadable MD5 len %d", len(md5))
		}
		if !strings.HasPrefix(strings.ToLower(file.Download.BinaryURL), "https://") {
			t.Fatalf("Downloadable url %q", file.Download.BinaryURL)
		}
	}
}
