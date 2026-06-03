package importer

import (
	"testing"
)

func TestDetectFormat(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    Format
	}{
		{
			name:    "alpine answerfile",
			content: "KEYMAPOPTS=\"us us\"\nHOSTNAMEOPTS=\"-n myhost\"\n",
			want:    FormatAlpineAnswerfile,
		},
		{
			name:    "ubuntu autoinstall",
			content: "#cloud-config\nautoinstall:\n  version: 1\n",
			want:    FormatUbuntuAutoinstall,
		},
		{
			name:    "ubuntu autoinstall keyword only",
			content: "autoinstall:\n  version: 1\n",
			want:    FormatUbuntuAutoinstall,
		},
		{
			name:    "debian preseed",
			content: "d-i netcfg/get_hostname string myhost\nd-i time/zone string Europe/London\n",
			want:    FormatDebianPreseed,
		},
		{
			name:    "kickstart",
			content: "#kickstart\nrootpw secret\n%packages\n@core\n%end\n",
			want:    FormatKickstart,
		},
		{
			name: "autoyast",
			content: `<?xml version="1.0"?>
<!DOCTYPE profile>
<profile xmlns="http://www.suse.com/1.0/yast2ns">
  <timezone><timezone>Europe/Berlin</timezone></timezone>
</profile>`,
			want: FormatAutoYaST,
		},
		{
			name:    "unknown",
			content: "this is just some random text",
			want:    FormatUnknown,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectFormat(tc.content)
			if got != tc.want {
				t.Errorf("DetectFormat() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestImport_Alpine(t *testing.T) {
	content := "KEYMAPOPTS=\"us us\"\nHOSTNAMEOPTS=\"-n alpinehost\"\nTIMEZONEOPTS=\"-z UTC\"\n"
	res, err := Import(content)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if res.Profile.OS.Family != "alpine" {
		t.Errorf("family = %q, want %q", res.Profile.OS.Family, "alpine")
	}
}

func TestImport_Ubuntu(t *testing.T) {
	content := "#cloud-config\nautoinstall:\n  version: 1\n  hostname: ubuntuhost\n  timezone: America/New_York\n"
	res, err := Import(content)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if res.Profile.System.Hostname != "ubuntuhost" {
		t.Errorf("hostname = %q, want %q", res.Profile.System.Hostname, "ubuntuhost")
	}
	if res.Profile.System.Timezone != "America/New_York" {
		t.Errorf("timezone = %q, want %q", res.Profile.System.Timezone, "America/New_York")
	}
}

func TestImport_DebianPreseed(t *testing.T) {
	content := "d-i netcfg/get_hostname string debianhost\nd-i time/zone select Europe/London\n"
	res, err := Import(content)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if res.Profile.System.Hostname != "debianhost" {
		t.Errorf("hostname = %q, want %q", res.Profile.System.Hostname, "debianhost")
	}
	if res.Profile.System.Timezone != "Europe/London" {
		t.Errorf("timezone = %q, want %q", res.Profile.System.Timezone, "Europe/London")
	}
}

func TestImport_Kickstart(t *testing.T) {
	content := "#kickstart\nnetwork --hostname=kshost\ntimezone US/Eastern\nrootpw secret\n%packages\n@core\n%end\n"
	res, err := Import(content)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if res.Profile.System.Hostname != "kshost" {
		t.Errorf("hostname = %q, want %q", res.Profile.System.Hostname, "kshost")
	}
	if res.Profile.System.Timezone != "US/Eastern" {
		t.Errorf("timezone = %q, want %q", res.Profile.System.Timezone, "US/Eastern")
	}
}

func TestImport_AutoYaST(t *testing.T) {
	content := `<?xml version="1.0"?>
<!DOCTYPE profile>
<profile xmlns="http://www.suse.com/1.0/yast2ns">
  <hostname>ayhost</hostname>
  <timezone>Europe/Berlin</timezone>
</profile>`
	res, err := Import(content)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if res.Profile.System.Hostname != "ayhost" {
		t.Errorf("hostname = %q, want %q", res.Profile.System.Hostname, "ayhost")
	}
	if res.Profile.System.Timezone != "Europe/Berlin" {
		t.Errorf("timezone = %q, want %q", res.Profile.System.Timezone, "Europe/Berlin")
	}
}

func TestImport_Unknown(t *testing.T) {
	_, err := Import("this is totally unrecognised content")
	if err == nil {
		t.Fatal("Import() expected error for unknown format, got nil")
	}
}
