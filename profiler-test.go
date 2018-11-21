package main

import (
	"fmt"
	"time"
	"os"
	"os/exec"
	"log"
	//"code/go/test-project/profiler"
	"bufio"
	"bytes"
)


// func main() {
// 	pid := os.Getpid()
// 	m1, err := profiler.CalculateMemory(os.Getpid())
// 	Counter()
// 	m2, err := profiler.CalculateMemory(pid)
// 	if err!=nil {

// 	} else {
// 		fmt.Printf("Memory used in process(pid: %d): %d KB \n", pid, (m1+m2)/2)
// 	}
// }

type Resource struct {
    memoryUsed uint64
    cpuUsed  uint64
    dateTime string
}

func main() {
	defer TimeTrack(time.Now(), "profiler")
	// if len(os.Args) == 1 {
	// 	fmt.Println("command not given.");
	// }
	fmt.Println(os.Args[1:])
	// take inputs from cmd
	//fmt.Println(len(os.Args), os.Args[1])
	//ends
	// process spawn 
	//cmd := exec.Command(os.Args[1])
	cmd := exec.Command("/usr/bin/curl", "sh", os.Args[1:])
    cmd.Stdout = os.Stdout
    err := cmd.Start()
    if err != nil {
       log.Fatal(err)
    }
    log.Printf("Just ran subprocess %d, exiting\n", cmd.Process.Pid)
	//ends
	// pid := os.Getpid()
	pid := cmd.Process.Pid
	Resources := []Resource{}
	for i := 0; i < 5; i++ {
		rs := Resource{}
		go CalculateMemory(pid, rs, &Resources)
		time.Sleep(100 * time.Millisecond)
	}
	
	//go profiler.CalculateMemory(pid)
	//time.Sleep(2 * time.Second)
	//Counter()
	fmt.Println(Resources)
	fmt.Printf("Program finished\n")
}

// func Counter() {
// 	//defer profiler.TimeTrack(time.Now(), "counter")
// 	i := 0
// 	for i < 100 {
// 		i++;
// 	}
// 	fmt.Println("Counted till 100")
// }


// func main() {
// 	pid := os.Getpid();
// 	f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", pid))
//     if err != nil {
//         //return 0, err
//     }
//     //pfx := []byte("Pss:")
//     r := bufio.NewScanner(f)
//     for r.Scan() {
//     	line := r.Bytes()
//     	//if bytes.HasPrefix(line, pfx) {
//     		fmt.Printf("%s\n", line);
//     	//}
//     }
//     //fmt.Printf("%s\n", f);
// }

func CalculateMemory(pid int, rs Resource, resAr *[]Resource) {
    f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", pid))
    if err != nil {
        //return 0, err
    }
    defer f.Close()

    res := uint64(0)
    pfx := []byte("Pss:")
    r := bufio.NewScanner(f)
    for r.Scan() {
        line := r.Bytes()
        if bytes.HasPrefix(line, pfx) {
            var size uint64
            _, err := fmt.Sscanf(string(line[4:]), "%d", &size)
            if err != nil {
                //return 0, err
            }
            res += size
        }
    }
    if err := r.Err(); err != nil {
        //return 0, err
    }
    rs.memoryUsed = res
   	rs.dateTime = time.Now().Format("Y-m-d H:i:s")

    //elapsed := time.Since(start)
    //rs.dateTime = elapsed
    *resAr = append(*resAr,rs)
   
    //return res, nil
}

func TimeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s\n", name, elapsed)
}