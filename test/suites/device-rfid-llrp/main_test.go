package test

import (
	"edgex-snap-testing/test/utils"
	"log"
	"os"
	"testing"
)

const (
	deviceRfidLlrpSnap = "edgex-device-rfid-llrp"
	deviceRfidLlrpApp  = "device-rfid-llrp"
)

func TestMain(m *testing.M) {
	teardown, err := utils.SetupServiceTests(deviceRfidLlrpSnap)
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
		Snap:                 deviceRfidLlrpSnap,
		App:                  deviceRfidLlrpApp,
	})

	utils.TestConfig(t, deviceRfidLlrpSnap, utils.Config{
		TestChangePort: utils.ConfigChangePort{
			App:                      deviceRfidLlrpApp,
			DefaultPort:              utils.ServicePort(deviceRfidLlrpApp),
			TestAppConfig:            true,
			TestGlobalConfig:         true,
			TestMixedGlobalAppConfig: utils.FullConfigTest,
		},
		TestAutoStart: true,
	})

	utils.TestNet(t, deviceRfidLlrpSnap, utils.Net{
		StartSnap:        true,
		TestOpenPorts:    []string{utils.ServicePort(deviceRfidLlrpApp)},
		TestBindLoopback: []string{utils.ServicePort(deviceRfidLlrpApp)},
	})

	utils.TestPackaging(t, deviceRfidLlrpSnap, utils.Packaging{
		TestSemanticSnapVersion: true,
	})
}
