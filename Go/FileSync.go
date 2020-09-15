package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var folderMap = map[string]string{"D:/James/Test/src": "D:/James/Test/dest", "D:/James/Test/src2": "D:/James/Test/dest2", "aaa": "bbb"}

var folderSyncMap = map[string]map[string]*time.Time{}
var locks = map[string]*sync.Mutex{}

var waitGroup sync.WaitGroup

var quit bool = false

const FILE_COPY_INFO_INTERVAL = 5000

func main() {
	const (
		// In milliseconds
		INFO_REFRESH_INTERVAL = 5000
		QUIT_INFO_INTERVAL    = 2000
		SYNC_FOLDER_INTERVAL  = 2000
	)

	logAction("main", "Starting File Sync...")

	initData()

	logVerbose("main", fmt.Sprintf("Folder Map: %s", folderMap))

	// check if user request quit
	go checkQuit()

	// print Sync map
	go func() {
		waitGroup.Add(1)
		for !quit {
			time.Sleep(INFO_REFRESH_INTERVAL * time.Millisecond)
			printSyncMap(Info)
		}
		waitGroup.Done()
	}()

	// loop for sync folders
	for !quit {

		for srcFolder, destFolder := range folderMap {
			go syncFolder(srcFolder, destFolder)
		}

		time.Sleep(SYNC_FOLDER_INTERVAL * time.Millisecond)
	}

	waitFinished := false
	// wait for routines to finish
	go func() {
		for !waitFinished {
			logAction("main", "Quitting: waiting for routines to finish")
			time.Sleep(QUIT_INFO_INTERVAL * time.Millisecond)
		}
	}()

	waitGroup.Wait()

	waitFinished = true

	// quitting
	logAction("main", "Quit system")
}

func initData() {
	for srcFolder, _ := range folderMap {
		locks[srcFolder] = new(sync.Mutex)
		folderSyncMap[srcFolder] = map[string]*time.Time{}
	}

	//	fmt.Println(fmt.Sprint("locks ====%s", locks))
}

func checkQuit() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		text = strings.Replace(text, "\r", "", -1)

		if strings.Compare("q", text) == 0 {
			quit = true

			break
		}
	}
}

func printSyncMap(level LogLevel) {
	msg := "File sync map:"

	for folerName, fileMap := range folderSyncMap {
		msg += fmt.Sprintf("\n\t\t%s:", folerName)
		for fileName, lastModified := range fileMap {
			msg += fmt.Sprintf("\n\t\t\t\t[%s] : [%s]", fileName, lastModified)
		}
	}

	log(level, "printSyncMap", msg)
}

func syncFolder(srcFolder, destFolder string) {
	waitGroup.Add(1)

	logInfo("syncFolder", fmt.Sprintf("sync %s -- > %s", srcFolder, destFolder))

	if isDir(srcFolder) && isDir(destFolder) {
		// find all files in this folder

		currentFiles, err := listFiles(srcFolder)
		if err != "" {
			logError("syncFolder", fmt.Sprint("List file error: %s", err))
			return
		}

		msg := "Current Files: " + srcFolder
		for k, _ := range currentFiles {
			msg += "\n\t" + k
		}

		logDebug("syncFolder", msg)

		//check and remove deleted files
		removeDeletedFiles(srcFolder, currentFiles)

		// for each file, start sync
		for file, lastModified := range currentFiles {
			go syncFile(srcFolder, destFolder, file, lastModified)
		}
	} else {
		logWarning("syncFolder", fmt.Sprintf("Kipped: Src or Dest folder doest not exists: %s --> %s ", srcFolder, destFolder))
	}

	waitGroup.Done()
}

func removeDeletedFiles(srcFolder string, currentFiles map[string]*time.Time) {
	if fileMap, ok := folderSyncMap[srcFolder]; ok {
		locks[srcFolder].Lock()

		// Remive entry if it is not in current listed files and entry value is finished sync

		toRemove := []string{}

		for fileName, lastModified := range fileMap {
			if lastModified != nil {
				if _, found := currentFiles[fileName]; !found {
					toRemove = append(toRemove, fileName)
				}
			}
		}

		for _, v := range toRemove {
			delete(fileMap, v)
			logInfo("removeDeletedFiles", "file removed from map "+v)
		}

		locks[srcFolder].Unlock()

	}
}

func syncFile(srcFolder string, destFolder string, fileName string, lastModified *time.Time) {
	waitGroup.Add(1)

	if StartSyncFile(srcFolder, destFolder, fileName, lastModified) {

		logAction("syncFile", fmt.Sprintf("File copy started: %s => %s : (%s)", srcFolder, destFolder, fileName))

		copied := false
		err := ""
		finished := false

		go func() {
			copied, err = fileCopy(srcFolder+"/"+fileName, destFolder+"/"+fileName)
			finished = true
		}()

		for !finished {
			time.Sleep(FILE_COPY_INFO_INTERVAL * time.Millisecond)
			logAction("syncFile", fmt.Sprintf("File copy in progress: %s => %s : (%s)", srcFolder, destFolder, fileName))
		}

		if copied {
			logAction("syncFile", fmt.Sprintf("File copy finished: %s => %s : (%s)", srcFolder, destFolder, fileName))
			FinishSyncFile(srcFolder, fileName, lastModified)
		} else {
			logError("syncFile", fmt.Sprintf("File copy failed with error: %s => %s : (%s) : %s ", srcFolder, destFolder, fileName, err))
			RemoveSyncFile(srcFolder, fileName)
		}
	} else {
		logVerbose("syncFile", fmt.Sprintf("Skipped: %s", fileName))
	}

	waitGroup.Done()
}

func StartSyncFile(srcFolder string, destFolder string, fileName string, lastModified *time.Time) bool {
	ret := false

	if fileMap, ok := folderSyncMap[srcFolder]; ok {

		locks[srcFolder].Lock()
		logDebug("StartSyncFile", "Locked >>>> "+srcFolder)
		if recLastModified, ok := fileMap[fileName]; !ok {
			var destFileName = destFolder + "/" + path.Base(fileName)
			var destExist, destLastModified = fileInfo(destFileName)

			if !destExist ||
				(destExist && lastModified.After(*destLastModified)) {
				// Add to map if destFile does not exist or src file is later than dest file
				fileMap[fileName] = nil
				printSyncMap(Debug)
				ret = true
			} else {
				// dest file is up-to-date, add to synced
				fileMap[fileName] = destLastModified
				printSyncMap(Debug)
			}
		} else {
			if recLastModified != nil {
				// sync finished
				if lastModified.After(*recLastModified) {
					// File modified after last sync
					fileMap[fileName] = nil
					printSyncMap(Debug)
					ret = true
				}
			}
		}

		locks[srcFolder].Unlock()
		logDebug("StartSyncFile", "UnLocked <<<< "+srcFolder)
	}

	return ret
}

func FinishSyncFile(srcFolder string, fileName string, fileLastModified *time.Time) {
	if fileMap, ok := folderSyncMap[srcFolder]; ok {

		// update sync finished
		locks[srcFolder].Lock()
		logDebug("FinishSyncFile", "Locked >>>> "+srcFolder)

		if _, ok := fileMap[fileName]; ok {
			fileMap[fileName] = fileLastModified

			printSyncMap(Debug)
		}

		locks[srcFolder].Unlock()
		logDebug("FinishSyncFile", "UnLocked <<<< "+srcFolder)
	}
}

func RemoveSyncFile(srcFolder string, fileName string) {
	if fileMap, ok := folderSyncMap[srcFolder]; ok {
		locks[srcFolder].Lock()
		logDebug("RemoveSyncFile", "Locked >>>> "+srcFolder)

		if _, ok := fileMap[fileName]; ok {
			// remove
			delete(fileMap, fileName)

			printSyncMap(Debug)
		}

		locks[srcFolder].Unlock()
		logDebug("RemoveSyncFile", "UnLocked <<<< "+srcFolder)
	}
}
