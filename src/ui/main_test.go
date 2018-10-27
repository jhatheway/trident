package TriUI

import (
	"os"
	"testing"
	tr "lib"
	trtst "lib/test"
)

func testingctx() pf.PfCtx {
	return tr.NewPfCtx(nil, nil, nil, nil, nil)
}

func TestingUI() PfUI {
	ctx := testingctx()
	return NewPfUI(ctx, nil, nil, nil)
}

func TestMain(m *testing.M) {
	toolname := trtst.Test_setup()

	/* UI Setup */
	err := Setup(toolname, false)
	if err != nil {
		tr.Errf("Failed to setup server PU: %s", err.Error())
		os.Exit(1)
	}

	/* Services */
	tr.Starts()
	defer tr.Stops()

	os.Exit(m.Run())
}
