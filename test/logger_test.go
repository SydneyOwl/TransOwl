package test

import (
	"github.com/gookit/slog"
	_ "github.com/sydneyowl/TransOwl/pkg/logger"
	"testing"
)

func TestLogger(t *testing.T) {
	slog.Panicf("testy")
	slog.Warnf("dsd")
}
