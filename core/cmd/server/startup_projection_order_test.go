package main

import (
	"os"
	"strings"
	"testing"
)

func TestProductStartupInstallsProjectionBeforeDispatch(t *testing.T) {
	raw, err := os.ReadFile("startup_product.go")
	if err != nil {
		t.Fatalf("read startup_product.go: %v", err)
	}
	source := string(raw)
	projection := strings.Index(source, "server.StartTeamWorkSignalProjection(ctx, adminSrv)")
	dispatch := strings.Index(source, "server.StartConfirmedActionDispatch(ctx, adminSrv)")
	if projection < 0 || dispatch < 0 {
		t.Fatalf("startup must explicitly install team projection and confirmed dispatch: projection=%d dispatch=%d", projection, dispatch)
	}
	if projection > dispatch {
		t.Fatal("confirmed dispatch starts before team result projection; a fast result can be lost during startup")
	}
}
