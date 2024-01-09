package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type Config struct {
	TestAutoStart  bool
}

func TestConfig(t *testing.T, snapName string, conf Config) {
	t.Run("config", func(t *testing.T) {
		TestAutoStart(t, snapName, conf.TestAutoStart)
	})
}

func TestAutoStart(t *testing.T, snapName string, testAutoStart bool) {
	if testAutoStart {
		t.Run("autostart", func(t *testing.T) {
			TestAutostartGlobal(t, snapName)
		})
	}
}

func TestAutostartGlobal(t *testing.T, snapName string) {
	t.Run("set and unset global autostart", func(t *testing.T) {
		t.Cleanup(func() {
			SnapUnset(t, snapName, "autostart")
			SnapStop(t, snapName)
		})

		SnapStop(t, snapName)
		require.False(t, SnapServicesEnabled(t, snapName))
		require.False(t, SnapServicesActive(t, snapName))

		SnapSet(t, snapName, "autostart", "true")
		require.True(t, SnapServicesEnabled(t, snapName))
		require.True(t, SnapServicesActive(t, snapName))

		SnapUnset(t, snapName, "autostart")
		require.True(t, SnapServicesEnabled(t, snapName))
		require.True(t, SnapServicesActive(t, snapName))

		SnapSet(t, snapName, "autostart", "false")
		require.False(t, SnapServicesEnabled(t, snapName))
		require.False(t, SnapServicesActive(t, snapName))
	})
}

func WaitForLogMessage(t *testing.T, snap, expectedLog string, since time.Time) {
	const maxRetry = 10

	for i := 1; i <= maxRetry; i++ {
		time.Sleep(1 * time.Second)
		t.Logf("Retry %d/%d: Waiting for expected content in logs: %s", i, maxRetry, expectedLog)

		logs := SnapLogs(t, since, snap)
		if strings.Contains(logs, expectedLog) {
			t.Logf("Found expected content in logs: %s", expectedLog)
			return
		}
	}

	t.Fatalf("Time out: reached max %d retries.", maxRetry)
}
