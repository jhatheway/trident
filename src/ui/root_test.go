package TriUI_test

import (
	"net/http"
	"testing"
	urltest "ui/urltest"
)

func TestUI_Main_Misc(t *testing.T) {
	tests := []urltest.URLTest{
		/* Root test */
		{"RootTest",
			"GET", "/",
			"",
			nil,
			nil,
			http.StatusOK, []string{}, []string{}},

		/* Missing pages check */
		urltest.URLTest_404("/gfx/404"),
	}

	/* Our Root */
	root := NewPfRootUI(pu.TestingUI)

	for _, u := range tests {
		urltest.Test_URL(t, root.H_root, u)
	}
}
