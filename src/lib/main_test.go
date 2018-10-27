package trident_test

/*
 * Support for setting up tests
 *
 * TODO: requires active database connection and valid config for now...
 */

import (
	"os"
	"testing"
	trtst "lib/test"
)

func TestMain(m *testing.M) {
	trtst.Test_setup()

	/* Services */
	Starts()
	defer Stops()

	os.Exit(m.Run())
}
