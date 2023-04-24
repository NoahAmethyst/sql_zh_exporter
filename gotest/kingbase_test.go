package gotest

import (
	kingbase "github.com/team-ide/go-driver/db_kingbase_v8r6"
	"testing"
)

func Test_KingBase(t *testing.T) {
	cdn := kingbase.GetDSN("SYSTEM", "123456", "127.0.0.1", 54321, "TEST")
	t.Logf("%s", cdn)
}
