package network

import (
	"testing"
)

func TestFormatConfigEmpty(t *testing.T) {
	lines := formatConfig("")
	if len(lines) != 0 {
		t.FailNow()
	}
}

func TestFormatConfigContinue(t *testing.T) {
	lines := formatConfig("line 1\\\nis long")
	if len(lines) != 1 {
		t.FailNow()
	}
}

func TestFormatConfigComment(t *testing.T) {
	lines := formatConfig("#comment")
	if len(lines) != 0 {
		t.FailNow()
	}
}

func TestFormatConfigCommentContinue(t *testing.T) {
	lines := formatConfig("#comment\\\ncomment")
	if len(lines) != 0 {
		t.FailNow()
	}
}

func TestFormatConfig(t *testing.T) {
	lines := formatConfig("  #comment \\\n comment\nline 1\nline 2\\\n is long")
	if len(lines) != 2 {
		t.FailNow()
	}
}

func TestProcessDebianNetconfNoConfig(t *testing.T) {
	interfaces, err := ProcessDebianNetconf("")
	if err != nil {
		t.FailNow()
	}
	if len(interfaces) != 0 {
		t.FailNow()
	}
}

func TestProcessDebianNetconfInvalidStanza(t *testing.T) {
	_, err := ProcessDebianNetconf("iface")
	if err == nil {
		t.FailNow()
	}
}

func TestProcessDebianNetconfNoInterfaces(t *testing.T) {
	interfaces, err := ProcessDebianNetconf("auto eth1\nauto eth2")
	if err != nil {
		t.FailNow()
	}
	if len(interfaces) != 0 {
		t.FailNow()
	}
}

func TestProcessDebianNetconf(t *testing.T) {
	interfaces, err := ProcessDebianNetconf("iface eth1 inet manual")
	if err != nil {
		t.FailNow()
	}
	if len(interfaces) != 1 {
		t.FailNow()
	}
}
