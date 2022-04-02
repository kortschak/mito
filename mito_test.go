package mito

import (
	"flag"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/kortschak/mito/lib"
	"github.com/rogpeppe/go-internal/testscript"
)

var update = flag.Bool("update", false, "update testscript output files")

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"mito": Main,
	}))
}

func TestScripts(t *testing.T) {
	t.Parallel()

	p := testscript.Params{
		Dir:           filepath.Join("testdata"),
		UpdateScripts: *update,
	}
	testscript.Run(t, p)
}

func TestSend(t *testing.T) {
	chans := map[string]chan interface{}{"ch": make(chan interface{})}
	send := lib.Send(chans)

	var got interface{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		got = <-chans["ch"]
	}()

	res, err := eval(`42.send_to("ch").close("ch")`, "", nil, send)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if res != "true" {
		t.Errorf("unexpected false result")
	}
	wg.Wait()
	if got != int64(42) {
		t.Errorf("unexpected sent result: got:%v want:42", got)
	}
}
