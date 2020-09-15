use std::collections::HashMap;
use std::path::Path;
use std::fs;
use std::thread;

pub fn run() {
    let folder_map = init_folder_map();

    let mut th_vec = Vec::new();

    for (src, dest) in folder_map{
        // let src_s = src.as_str().clone();
        // let dest_s = dest.as_str().clone();

        let src_path = Path::new(src);
        let dest_path = Path::new(dest);

        if src_path.is_dir() && dest_path.is_dir() {
            let th = thread::spawn( move||
                {
                    sync_folder(src, dest);
                });
            th_vec.push(th);
        }else {
            println!("Folder not exists or not folder: {} -> {}", src, dest)
        }
    }

    for th in th_vec{
        th.join();
    }
}

fn init_folder_map() -> HashMap<&'static str, &'static str> {
    let mut folder_map: HashMap<&str, &str> = HashMap::new();

    folder_map.insert("/home/james/tmp/copyFrom/AA", "/home/james/tmp/copyTo/AA");
    folder_map.insert("/home/james/tmp/copyFrom/BB", "/home/james/tmp/copyTo/BB");
    folder_map.insert("/aaa", "/bbb");

    folder_map
}

fn sync_folder(src: &'static str, dest: &'static str){
    println!("Start Sync Folder: {} -> {}", src, dest);
    let mut files = Vec::new();

    let src_path = Path::new(src);

    for entry in fs::read_dir(src_path).unwrap() {
        let path = entry.unwrap().path();
        if path.is_file() {
            let file_name = String::from(path.file_name().unwrap().to_str().unwrap());
            println!("Found file: {}", file_name);
            files.push(file_name);
        }
    }

    let mut th_vec = Vec::new();

    for file_name in files{
        let th = thread::spawn(move ||
            {
                sync_file(src, dest, file_name.as_str());
            });

           th_vec.push(th);
    }

    for th in th_vec{
        th.join().expect("Couldn't join on the associated thread");
    }
}

fn sync_file(src: & str, dest:  & str, file_name : &str) {
    println!("Start Sync file: {} from {} to {}", file_name, src, dest);

    let src_folder = Path::new(src);
    let dest_folder = Path::new(dest);

    let mut src_file_name = String::from(src_folder.to_str().unwrap());
    src_file_name.push('/');
    src_file_name.push_str(file_name);
    let src_file_path = Path::new(src_file_name.as_str());

    let mut dest_file_name = String::from(dest_folder.to_str().unwrap());
    dest_file_name.push('/');
    dest_file_name.push_str(file_name);
    let dest_file_path = Path::new(dest_file_name.as_str());

    if src_file_path.exists() {
        if dest_file_path.exists(){
            let src_file_meta = fs::metadata(src_file_path);
            let dest_file_meta = fs::metadata(dest_file_path);

            match src_file_meta.unwrap().modified(){
                Ok(src_time) => {
                    match dest_file_meta.unwrap().modified() {
                        Ok(dest_time) => {
                            if src_time.gt(&dest_time) {
                                // src file is newer
                                copy_file(src_file_path, dest_file_path);
                            } else {
                                println!("Ignored, dest file is up-to-date {}", dest_file_name);
                            }
                        },
                        Err(err) => println!("Error when Check dest file modified date: {}  \n {} ", dest_file_name, err.to_string()),
                    }
                },
                Err(err) => println!("Error when Check src file modified date: {}  \n {} ", src_file_name, err.to_string()),
            }
        }else {
            // dest not exits, just copy
            copy_file(src_file_path, dest_file_path);
        }

    }else{
        println!("Can't find source file: {}", src_file_name);
    }
}

fn copy_file(src_file: & Path, dest_file : & Path){
    println!("Copy file Started: from {} => {}",  src_file.to_str().unwrap(), dest_file.to_str().unwrap());

    match fs::copy(src_file, dest_file){
        Ok(_n) => println!("Copy file Finished: from {} => {}",  src_file.to_str().unwrap(), dest_file.to_str().unwrap()),
        Err(err) => {
            println!("Copy file ERROR: from {} => {}  \n {}", src_file.to_str().unwrap(), dest_file.to_str().unwrap(), err.to_string());
        },
    }
}
// for each folder pair
    // list files in source folder
    // check if it's newer than dest folder
        // if yes, copy file from source to dest
