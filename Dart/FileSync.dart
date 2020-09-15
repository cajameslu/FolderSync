import "dart:async";
import 'dart:convert';
import "dart:io";
import "package:mutex/mutex.dart";
import "dart:math";
import 'package:path/path.dart' as path;

Map folderMap = {'D:\\James\\Test\\src' : 'D:\\James\\Test\\dest', 'D:\\James\\Test\\src2' : 'D:\\James\\Test\\dest2', "aaa": "bbb"};
bool quit = false;

// Key: file name
// Value: false, sync in progress; true, sync finished
Map filesInSync = <String, Map<String, DateTime>> {};
Map<String, Mutex> locks = {};

void main() async{
  initData();

  StreamSubscription<String> inListener;

  inListener = readLine().listen((line) {
    if (line.toLowerCase() == "q") {
      quit = true;
    }
  });

  var printSync = Timer.periodic(Duration(milliseconds: 20000), (Timer t) {
    var msg = "File sync map:";
    filesInSync.forEach((k, v) => msg += "\n\t\t" + k.toString() + ": " + v.toString());

    logInfo("main", msg);
  });

  while (!quit){
    folderMap.forEach((k, v) => syncFolder(k, v));
    await Future.delayed(Duration(milliseconds: 500));
  }

  await inListener.cancel();
  await printSync.cancel();

  logInfo("main", "Quitting...");
}

printSyncMap(){
  var msg = "File sync map:";
  filesInSync.forEach((k, v) => msg += "\n\t\t" + k.toString() + ": " + v.toString());

  logInfo("main", msg);
}

Stream<String> readLine() => stdin
    .transform(utf8.decoder)
    .transform(const LineSplitter());

void initData(){
  folderMap.forEach((srcFolder, destFolder) {
    locks[srcFolder] = Mutex();
    filesInSync[srcFolder] = <String, DateTime>{};
  });
}

Future syncFolder(String srcFolder, String destFolder) async {
//  return Future(() {
  logInfo("syncFolder", 'sync $srcFolder -- > $destFolder');

  var srcDir = Directory(srcFolder);
  var destdir = Directory(destFolder);

  if (srcDir.existsSync() && destdir.existsSync()) {
    // find all files in this folder
    var currentFiles = <File>{};
    List contents = srcDir.listSync();
    for(var item in contents){
      if (item is File) {
        currentFiles.add(item);
      }
    }
    logDebug("syncFolder", "Current Files: $currentFiles");
    //check and remove deleted files
    await removeDeletedFiles(srcFolder, currentFiles);

    // for each file, start sync
    currentFiles.forEach((File file) {
      syncFile(srcFolder, destFolder, file);
    });
  }
  else{
    logWarning("syncFolder", "Kipped: Src or Dest folder doest not exists: $srcFolder : $destFolder");
  }
//  });
}

Future removeDeletedFiles(String srcFolder, Set currentFiles) async{
  if (filesInSync.containsKey(srcFolder)) {
    Map<String, DateTime> fileEntries = filesInSync[srcFolder];
    Iterable<String> currentFileNames = currentFiles.map<String>((file) => file.path);
    
    try {
      await locks[srcFolder].acquire();

      // Remive entry if it is not in current listed files and entry value is finished sync
      fileEntries.removeWhere((fileName, lastModified) => !currentFileNames.contains(fileName) && lastModified != null);
    }finally {
      locks[srcFolder].release();
    }
  }
}

Future syncFile(String srcFolder, String destFolder, File file) async{
//  return Future(() {
    if (await StartSyncFile(srcFolder, destFolder, file)){
      // copy
//      var random = Random.secure();
//      int sleepTime = random.nextInt(1000);
//      await Future.delayed(Duration(milliseconds: sleepTime));

      try {
        var baseName = path.basename(file.path);
        var lastModified = file.lastModifiedSync();
        logAction("syncFile", "File copying: $srcFolder => $destFolder : (${baseName})");
        file.copySync(destFolder + "\\" + baseName);
        logAction("syncFile", "File copied: $srcFolder => $destFolder : (${baseName})");

        await FinishSyncFile(srcFolder, file.path, lastModified);
      }catch(e){
        logError("syncFile", e.toString());
        await RemoveSyncFile(srcFolder, file.path);
      }
    }else{
      logInfo("syncFile", "Skipped: ${file.path}");
    }

//  });
}

Future<bool> StartSyncFile(String srcFolder, String destFolder, File file) async{
  if (filesInSync.containsKey(srcFolder)) {
    try {
      await locks[srcFolder].acquire();

      if (!filesInSync[srcFolder].containsKey(file.path)) {
        var destFile = File(destFolder + "\\" + path.basename(file.path));
        if(! destFile.existsSync() ||
            (destFile.existsSync() && file.lastModifiedSync().isAfter(destFile.lastModifiedSync()))) {
          // Add to map if destFile does not exist or src file is later than dest file
          filesInSync[srcFolder][file.path] = null;
          printSyncMap();
          return true;
        }else{
          // dest file is up-to-date, add to synced
          filesInSync[srcFolder][file.path] = destFile.lastModifiedSync();
          printSyncMap();
        }
      }else{
        if (filesInSync[srcFolder][file.path] != null){
          // sync finished
          if (file.lastModifiedSync().isAfter(filesInSync[srcFolder][file.path])){
            // File modified after last sync
            filesInSync[srcFolder][file.path] = null;
            printSyncMap();
            return true;
          }
        }
      }
    }catch(e){
      logError("StartSyncFile", e.toString());
    }
    finally {
      locks[srcFolder].release();
    }
  }

  return false;
}

Future FinishSyncFile(String srcFolder, String fileName, DateTime fileLastModified) async{
  if (filesInSync.containsKey(srcFolder)){
    try {
        // update sync finished
      await locks[srcFolder].acquire();
      if (filesInSync[srcFolder].containsKey(fileName)) {
        filesInSync[srcFolder][fileName] = fileLastModified;

        printSyncMap();
      }
    }catch(e){
      logError("FinishSyncFile", e.toString());
    }finally{
      locks[srcFolder].release();
    }
  }
}

Future RemoveSyncFile(String srcFolder, String fileName) async{
  if (filesInSync.containsKey(srcFolder)) {
    try {
      await locks[srcFolder].acquire();
      if (filesInSync[srcFolder].containsKey(fileName)) {
        // remove
        filesInSync[srcFolder].remove(fileName);

        printSyncMap();
      }
    }catch(e){
      logError("RemoveSyncFile", e.toString());
    }
    finally {
      locks[srcFolder].release();
    }
  }
}

enum LogLevel{
  Debug,
  Info,
  Action,
  Warning,
  Error
}

String getLogLevelText(LogLevel level){
  return level.toString().split(".").last;
}

logDebug(String header, String msg){
  log(LogLevel.Debug, header, msg);
}

logInfo(String header, String msg){
  log(LogLevel.Info, header, msg);
}

logAction(String header, String msg){
  log(LogLevel.Action, header, msg);
}

logWarning(String header, String msg){
  log(LogLevel.Warning, header, msg);
}

logError(String header, String msg){
  log(LogLevel.Error, header, msg);
}

LogLevel outputLevel = LogLevel.Info;

log(LogLevel level, String header, String msg){
  if(level.index >= outputLevel.index) {
    print("[${DateTime.now()}]\t[${getLogLevelText(level)}]\t[$header]\t$msg");
  }
}