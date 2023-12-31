package test

import (
	"edgex-snap-testing/test/utils"
	"log"
	"os"
	"testing"
)

const (
	deviceRestSnap = "edgex-device-rest"
	deviceRestApp  = "device-rest"
)

func TestMain(m *testing.M) {
	teardown, err := utils.SetupServiceTests(deviceRestSnap)
	if err != nil {
		log.Fatalf("Failed to setup tests: %s", err)
	}

	code := m.Run()
	teardown()

	os.Exit(code)
}

func TestCommon(t *testing.T) {
	utils.TestContentInterfaces(t, utils.ContentInterfaces{
		TestSecretstoreToken: true,
		Snap:                 deviceRestSnap,
		App:                  deviceRestApp,
	})

	utils.TestConfig(t, deviceRestSnap, utils.Config{
		TestChangePort: utils.ConfigChangePort{
			App:                      deviceRestApp,
			DefaultPort:              utils.ServicePort(deviceRestApp),
			TestAppConfig:            true,
			TestGlobalConfig:         true,
			TestMixedGlobalAppConfig: utils.FullConfigTest,
		},
		TestAutoStart: true,
	})

	utils.TestNet(t, deviceRestSnap, utils.Net{
		StartSnap:        true,
		TestOpenPorts:    []string{utils.ServicePort(deviceRestApp)},
		TestBindLoopback: []string{utils.ServicePort(deviceRestApp)},
	})

	utils.TestPackaging(t, deviceRestSnap, utils.Packaging{
		TestSemanticSnapVersion: true,
	})
}
