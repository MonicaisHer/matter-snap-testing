package test

import (
	"fmt"
	"log"
	"matter-snap-testing/test/utils"
	"os"
	"strings"
	"testing"
	"time"
	"sync"
	"os/exec"
)

const chipToolSnap = "chip-tool"

var start = time.Now()

func TestMain(m *testing.M) {
	teardown, err := setup()
	if err != nil {
		log.Fatalf("Failed to setup tests: %s", err)
	}

	code := m.Run()
	teardown()

	os.Exit(code)
}


var cmd *exec.Cmd
func TestMatterDeviceOperations(t *testing.T) {
	//setup
	if err := os.Remove("./chip-all-clusters-minimal-app-ptpython-fix.log"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Error deleting log file: %s\n", err)
	}
	if err := os.Remove("./chip-tool.log"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Error deleting log file: %s\n", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	logFile, err := os.Create("chip-all-clusters-minimal-app-ptpython-fix.log")
	if err != nil {
		t.Fatalf("Error creating log file: %s\n", err)
	}
	defer logFile.Close()

	// run chip-all-clusters-minimal-app-ptpython-fix in the background
	go func() {
		defer wg.Done()

		cmd = exec.Command("./chip-all-clusters-minimal-app-ptpython-fix")
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		err := cmd.Start()
		if err != nil {
			fmt.Printf("Error starting application: %s\n", err)
		}
	}()

	defer func() {
        if err := cmd.Process.Kill(); err != nil {
            t.Fatalf("Error killing process: %s\n", err)
        }
    }()

	wg.Wait()


	t.Run("Commission", func(t *testing.T) {
		utils.Exec(t, "sudo chip-tool pairing onnetwork 110 20202021")
	})

	t.Run("Control", func(t *testing.T) {
		utils.Exec(t, "sudo chip-tool onoff toggle 110 1")
		WaitForAppMessage(t, "./chip-all-clusters-minimal-app-ptpython-fix.log", "CHIP:ZCL: Toggle ep1 on/off", start)
	})
}

func setup() (teardown func(), err error) {
	log.Println("[CLEAN]")
	utils.SnapRemove(nil, chipToolSnap)

	log.Println("[SETUP]")

	teardown = func() {
		log.Println("[TEARDOWN]")
		utils.SnapDumpLogs(nil, start, chipToolSnap)

		log.Println("Removing installed snap:", !utils.SkipTeardownRemoval)
		if !utils.SkipTeardownRemoval {
			utils.SnapRemove(nil, chipToolSnap)
		}
	}

	if utils.LocalServiceSnap() {
		err = utils.SnapInstallFromFile(nil, utils.LocalServiceSnapPath)
	} else {
		err = utils.SnapInstallFromStore(nil, chipToolSnap, utils.ServiceChannel)
	}
	if err != nil {
		teardown()
		return
	}

	// connect interfaces
	utils.SnapConnect(nil, chipToolSnap+":avahi-observe", "")
	utils.SnapConnect(nil, chipToolSnap+":bluez", "")
	utils.SnapConnect(nil, chipToolSnap+":process-control", "")

	return
}

func WaitForAppMessage(t *testing.T, appLogPath, expectedLog string, since time.Time) {
	const maxRetry = 10

	for i := 1; i <= maxRetry; i++ {
		time.Sleep(1 * time.Second)
		t.Logf("Retry %d/%d: Waiting for expected content in logs: %s", i, maxRetry, expectedLog)

		logs, err := readLogFile(appLogPath)
		if err != nil {
			fmt.Println("Error reading log file:", err)
			continue
		}

		if strings.Contains(logs, expectedLog) {
			t.Logf("Found expected content in logs: %s", expectedLog)
			return
		}
	}

	t.Fatalf("Time out: reached max %d retries.", maxRetry)
}

func readLogFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", err
	}

	fileSize := stat.Size()
	buffer := make([]byte, fileSize)

	_, err = file.Read(buffer)
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}
