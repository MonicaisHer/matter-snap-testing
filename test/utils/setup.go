package utils

import (
	"log"
	"time"
)

// SetupServiceTests setup up the environment for testing
// It returns a teardown function to be called at the end of the tests
func SetupServiceTests(snapName string) (teardown func(), err error) {
	log.Println("[CLEAN]")
	SnapRemove(nil,
		snapName,
	)

	log.Println("[SETUP]")
	start := time.Now()

	teardown = func() {
		log.Println("[TEARDOWN]")
		SnapDumpLogs(nil, start, snapName)

		log.Println("Removing installed snap:", !SkipTeardownRemoval)
		if !SkipTeardownRemoval {
			SnapRemove(nil,
				snapName,
			)
		}
	}

	if LocalServiceSnap() {
		err = SnapInstallFromFile(nil, LocalServiceSnapPath)
	} else {
		err = SnapInstallFromStore(nil, snapName, ServiceChannel)
	}
	if err != nil {
		teardown()
		return
	}

	return
}
