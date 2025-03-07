package llcppg

import "testing"

func TestFindGoMod(t *testing.T) {
	goModFile = "ggg.test"
	l := &llcppgGenerator{dir: "."}
	if ret := l.findGoMod(); ret != "testfind" {
		t.Errorf("unexpected find result: got %s want: testfind", ret)
	}
}
