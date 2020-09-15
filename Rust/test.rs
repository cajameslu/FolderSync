use std::thread;

struct ThreadData<'a>{
    name : &'a String,
    id : i8,
}

pub fn run(){
    test1();

}

pub fn test1(){
    let mut ths = Vec::new();
    let name = String::from("James test");

    for i in 1..20{
        let mut tname = String::from(name.as_str());
        tname.push(' ');
        tname.push_str(i.to_string().as_str());



        let th = thread::spawn(move ||
            {
                // println!("{}", tname);
                let td = ThreadData {name: &tname, id: i};
                println!("Thread run {}", td.name);
            });

        // th.join();
        ths.push(th);
    }

    for th in ths{
        th.join();
    }
}