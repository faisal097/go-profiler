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
    //"strconv"
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

var hertz float64

func main() {

	defer TimeTrack(time.Now(), "profiler")
    go exe_cmd(os.Args[1])
	Resources := []Memory{}
    Processes := []CPU{}
    hertz = 1000
	for i := 0; i < 3; i++ {
        fmt.Println("-----------------------------------------------------------------------------");
		rs := Memory{}
        pr := CPU{}
        go CalculateMemory(pid, rs, &Resources)
        go CalculateCPU(pid, pr, &Processes)
		// go CalculateMemory(12939, rs, &Resources)
        // go CalculateCPU(12939, pr, &Processes)
        // go CPUUsage(pid, pr, &Processes)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println(Resources)
    fmt.Println(Processes)
	fmt.Printf("Program finished\n")
}

//https://unix.stackexchange.com/questions/58539/top-and-ps-not-showing-the-same-cpu-result
func CalculateCPU(opid int, pr CPU, prArr *[]CPU) {
    f, err := os.Open(fmt.Sprintf("/proc/%d/stat", opid))
    if err != nil {
        //return 0, err
    }
    defer f.Close()

    //hertz := getCPUHZ()

    go getCPUHZ()

    fmt.Printf("Hertz: %f\n", hertz)
    utime := float64(0)
    stime := float64(0)
    cuttime := float64(0)
    cstime := float64(0)
    starttime := float64(0)
    var proctimes []string
    //var seconds float64

    r := bufio.NewScanner(f)
    for r.Scan() {
        line := r.Text()
        proctimes = strings.Split(line, " ")
        fmt.Sscanf(proctimes[13], "%f", &utime)
        fmt.Sscanf(proctimes[14], "%f", &stime)
        fmt.Sscanf(proctimes[15], "%f", &cuttime)
        fmt.Sscanf(proctimes[15], "%f", &cstime)
        fmt.Sscanf(proctimes[21], "%f", &starttime)

        // fmt.Printf("utime: %f\n", utime)
        // fmt.Printf("stime: %f\n", stime)
        // fmt.Printf("cuttime: %f\n", cuttime)
        // fmt.Printf("cstime: %f\n", cstime)
        // fmt.Printf("starttime: %f\n", starttime)

        total_time := utime + stime + cuttime + cstime
        fmt.Printf("total cpu clocks: %f\n", total_time)
        fmt.Printf("total cpu usage in hertz: %f\n", total_time/hertz)

        t := time.Now()
        pr.dateTime = t.Format(time.RFC3339)
        pr.pid = opid
        pr.cpu = total_time/hertz
        *prArr = append(*prArr,pr)

        //percentage calcualtioc
        //seconds = uptime - starttime/hertz  // get uptime from tail -f /proc/uptime
        //fmt.Printf("seconds: %f\n", seconds)
        // if seconds > 0 {
        //     total_time := utime + stime + cuttime + cstime
        //     fmt.Printf("total_time: %f\n", total_time)
        //     pcpu := ( total_time * 1000 / hertz) / seconds
        //     fmt.Printf("pcpu: %f\n", pcpu)
        // }
    }
    if err := r.Err(); err != nil {
        //return 0, err
    }
}

func getCPUHZ() {
    line := float64(0)
    out, err := exec.Command("sh", "-c", "lscpu | grep -m1 MHz").Output()
    if err != nil {
        log.Fatal(err)
        //return line
    }
    _, err1 := fmt.Sscanf(string(out[8:]), "%f", &line)
    if err1 != nil {
        log.Fatal(err1)
    }
    hertz = line*1000
    //fmt.Printf("MHz %f\n", line)
    //return line*1000
}

// func getCPUHZ() float64 {
//     line := float64(0)
//     out, err := exec.Command("sh", "-c", "lscpu | grep -m1 MHz").Output()
//     if err != nil {
//         log.Fatal(err)
//         return line
//     }
//     _, err1 := fmt.Sscanf(string(out[8:]), "%f", &line)
//     if err1 != nil {
//         log.Fatal(err1)
//     }
//     //fmt.Printf("MHz %f\n", line)
//     return line*1000
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
    *resAr = append(*resAr,rs)
    //return res, nil
}

func TimeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s\n", name, elapsed)
}

// func CPUUsage(opid int, pr CPU, prArr *[]CPU) {
//     cmd := exec.Command("ps", "aux")
//     var out bytes.Buffer
//     cmd.Stdout = &out
//     err := cmd.Run()
//     if err != nil {
//         log.Fatal(err)
//     }
//     for {
//         line, err := out.ReadString('\n')
//         if err!=nil {
//             break;
//         }
//         tokens := strings.Split(line, " ")
//         ft := make([]string, 0)
//         for _, t := range(tokens) {
//             if t!="" && t!="\t" {
//                 ft = append(ft, t)
//             }
//         }
//         pid, err := strconv.Atoi(ft[1])
//         if err!=nil {
//             continue
//         }
//         cpu, err := strconv.ParseFloat(ft[2], 64)
//         if err!=nil {
//             log.Fatal(err)
//         }
//         if pid == opid {
//             t := time.Now()
//             pr.dateTime = t.Format(time.RFC3339)
//             pr.pid = pid
//             pr.cpu = cpu
//             *prArr = append(*prArr,pr)
//         }
//     }
// }

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