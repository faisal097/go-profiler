package main

import (
	"fmt"
	"time"
	"os"
	"os/exec"
	"log"
	"bufio"
	"bytes"
    "strings"
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
    Resources := []Memory{}
    Processes := []CPU{}
    hertz = 1000
    if len(os.Args[1:]) <= 0 {
        log.Fatal("Please type command after binary's name seperated by space")
        os.Exit(0)
    }
    //Join(a []string, sep string) string
    command := strings.Join(os.Args[1:], " ")
    
    go exe_cmd(command)
	for i := 0; i < 15; i++ {
		rs := Memory{}
        pr := CPU{}
        go CalculateMemory(pid, rs, &Resources)
        go CalculateCPU(pid, pr, &Processes)
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println(Resources)
    fmt.Println(Processes)
	fmt.Printf("Program finished\n")
}

func exe_cmd(cmd string) {
    fmt.Printf("Entered command : %s", cmd)
    cmnd := exec.Command("sh", "-c", cmd)
    cmnd.Stdout = os.Stdout
    err := cmnd.Start()
    pid = cmnd.Process.Pid
    if err != nil {
        log.Fatal(err)
        os.Exit(0)
    }
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

    //fmt.Printf("Hertz: %f\n", hertz)
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
        //fmt.Printf("total cpu clocks: %f\n", total_time)
        //fmt.Printf("total cpu usage in hertz: %f\n", total_time/hertz)

        t := time.Now()
        pr.dateTime = t.Format(time.RFC3339)
        pr.pid = opid
        pr.cpu = total_time/hertz
        *prArr = append(*prArr,pr)
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
    }
    _, err1 := fmt.Sscanf(string(out[8:]), "%f", &line)
    if err1 != nil {
        log.Fatal(err1)
    }
    hertz = line*1000000
}

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
//     return line*1000000
// }

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








































//CPU usage: https://stackoverflow.com/questions/11356330/getting-cpu-usage-with-golang