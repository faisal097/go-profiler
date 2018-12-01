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
    "strconv"
    "strings"
    //"sync"
)

type Memory struct {
    pid int
    memoryUsed uint64
    dateTime string
}

type CPU struct {
    pid int
    cpu float64
    dateTime string
}

var pid int

func main() {
	defer TimeTrack(time.Now(), "profiler")
	
	// fmt.Println(os.Args[1])
	// take inputs from cmd
	
    go exe_cmd(os.Args[1])

	// cmd := exec.Command("sh", "-c", os.Args[1])
 //    cmd.Stdout = os.Stdout
 //    err := cmd.Start()
 //    if err != nil {
 //       log.Fatal(err)
 //    }
 //    log.Printf("Just ran subprocess %d, exiting\n", cmd.Process.Pid)
	// //ends
	
	// pid := cmd.Process.Pid

    //pid = 6548
	Resources := []Memory{}
    Processes := []CPU{}

	for i := 0; i < 5; i++ {
		rs := Memory{}
        pr := CPU{}
		go CalculateMemory(pid, rs, &Resources)
        go CPUUsage(pid, pr, &Processes)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println(Resources)
    fmt.Println(Processes)
	fmt.Printf("Program finished\n")
}

// func main() {
//     defer TimeTrack(time.Now(), "profiler")
//     fmt.Println(os.Args[1])
//     // take inputs from cmd
//     command_test(os.Args[1]);
//     fmt.Printf("Program finished\n")
// }


// func command_test(cmd string) {
//     done := make(chan bool, 1)
//     go exe_cmd(cmd, done)
//     Resources := []Resource{}
//     for {
//         if <-done {
//             fmt.Println("Channel ends");
//             break
//         }
//         rs := Resource{}
//         fmt.Println("pid", pid)
//         go CalculateMemory(pid, rs, &Resources)
//         time.Sleep(100 * time.Millisecond)    
//     }
// }

func exe_cmd(cmd string) {
    fmt.Println(cmd)
    cmnd := exec.Command("sh", "-c", cmd)
    cmnd.Stdout = os.Stdout
    err := cmnd.Start()
    pid = cmnd.Process.Pid
    if err != nil {
        fmt.Println("error occured")
        fmt.Printf("%s", err)
    }
    fmt.Printf("%s", err)
}

// func exe_cmd(cmd string, done chan bool) {
//     fmt.Println(cmd)
//     cmnd := exec.Command("sh", "-c", cmd)
//     cmnd.Stdout = os.Stdout
//     err := cmnd.Start()
//     pid = cmnd.Process.Pid
//     if err != nil {
//         fmt.Println("error occured")
//         fmt.Printf("%s", err)
//     }
//     fmt.Printf("ddd %s", err)
//     done <- true
// }


func CalculateMemory(pid int, rs Memory, resAr *[]Memory) {
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
    t := time.Now()
    rs.memoryUsed = res
    rs.dateTime = t.Format(time.RFC3339)
    rs.pid = pid
   	// rs.dateTime = time.Now().Format("Y-m-d H:i:s")
    *resAr = append(*resAr,rs)
    //return res, nil
}

func CPUUsage(opid int, pr CPU, prArr *[]CPU) {
    cmd := exec.Command("ps", "aux")
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
    for {
        line, err := out.ReadString('\n')
        if err!=nil {
            break;
        }
        tokens := strings.Split(line, " ")
        ft := make([]string, 0)
        for _, t := range(tokens) {
            if t!="" && t!="\t" {
                ft = append(ft, t)
            }
        }
        pid, err := strconv.Atoi(ft[1])
        if err!=nil {
            continue
        }
        cpu, err := strconv.ParseFloat(ft[2], 64)
        if err!=nil {
            log.Fatal(err)
        }
        if pid == opid {
            t := time.Now()
            pr.dateTime = t.Format(time.RFC3339)
            pr.pid = pid
            pr.cpu = cpu
            *prArr = append(*prArr,pr)
        }
    }
}

// func CalculateCPU(pid int, rs Resource, resAr *[]Resource) {
//     f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", pid))
//     if err != nil {
//         //return 0, err
//     }
//     defer f.Close()

//     res := uint64(0)
//     pfx := []byte("Pss:")
//     r := bufio.NewScanner(f)
//     for r.Scan() {
//         line := r.Bytes()
//         if bytes.HasPrefix(line, pfx) {
//             var size uint64
//             _, err := fmt.Sscanf(string(line[4:]), "%d", &size)
//             if err != nil {
//                 //return 0, err
//             }
//             res += size
//         }
//     }
//     if err := r.Err(); err != nil {
//         //return 0, err
//     }
//     rs.memoryUsed = res
//     rs.dateTime = time.Now().Format("Y-m-d H:i:s")
//     *resAr = append(*resAr,rs)
//     //return res, nil
// }

func TimeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s\n", name, elapsed)
}

//https://stackoverflow.com/questions/20437336/how-to-execute-system-command-in-golang-with-unknown-arguments

// func setInterval(someFunc func(), milliseconds int, async bool) chan bool {

//     // How often to fire the passed in function 
//     // in milliseconds
//     interval := time.Duration(milliseconds) * time.Millisecond

//     // Setup the ticket and the channel to signal
//     // the ending of the interval
//     ticker := time.NewTicker(interval)
//     clear := make(chan bool)

//     // Put the selection in a go routine
//     // so that the for loop is none blocking
//     go func() {
//         for {

//             select {
//             case <-ticker.C:
//                 if async {
//                     // This won't block
//                     go someFunc()
//                 } else {
//                     // This will block
//                     someFunc()
//                 }
//             case <-clear:
//                 ticker.Stop()
//                 return
//             }

//         }
//     }()

//     // We return the channel so we can pass in 
//     // a value to it to clear the interval
//     return clear

// }








































//CPU usage: https://stackoverflow.com/questions/11356330/getting-cpu-usage-with-golang