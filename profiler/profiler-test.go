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
    //"sync"
    "io"
    "os/signal"
    "syscall"
    "unique"//custom package .. used to remove duplicates from array
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
var PGID int
var hertz float64

var Resources []Memory
var Processes []CPU
var Pids []int

var totalMem uint64
var totalCPU float64

func main() {
    hertz = 1000000
	defer TimeTrack(time.Now(), "profiler")
    defer func () {
        fmt.Printf("\n")
        fmt.Println(Resources)
        fmt.Println(Processes)
        fmt.Println(Pids)
        killProcess()
        fmt.Printf("Program finished\n")
        /*
        Clean up work
        Delete all the sub-processes and current process started
        */
    }()

    if len(os.Args[1:]) <= 0 {
        log.Fatal("Please type command after binary's name seperated by space")
        os.Exit(1)
    }


    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    signal.Notify(c, syscall.SIGTERM)
    go func() {
        <-c
        log.Println("Receved Ctrl + C")
        killProcess()
    }()

    command := strings.Join(os.Args[1:], " ")
    go takeSnapshots()
    exe_cmd(command)
}

//https://github.com/golang/go/issues/23152
func exe_cmd(cmd string) {
    go getCPUHZ()

    fmt.Printf("Entered command : %s\n", cmd)
    cmnd := exec.Command("sh", "-c", cmd)
    defer cmnd.Wait()
    
    stdin, err := cmnd.StdinPipe()
    log.Println(err)

    stderr, err := cmnd.StderrPipe()
    log.Println(err)

    stdout, err := cmnd.StdoutPipe()
    log.Println(err)

    cmnd.Stdout = os.Stdout
    err = cmnd.Start()
    fmt.Println("Process started!")
    pid = cmnd.Process.Pid
    fmt.Printf("Parent Pid: %d\n", pid)

    //get group pid
    getParentGroupId()
    //get child pids
    getChildPids()
    //append current process id
    Pids = append(Pids,pid)

    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
    go io.Copy(os.Stdout, stdout)
    go io.Copy(stdin, os.Stdin)
    go io.Copy(os.Stderr, stderr)
}

//https://unix.stackexchange.com/questions/58539/top-and-ps-not-showing-the-same-cpu-result
func CalculateCPU(ind int, cpid int) {
    //fmt.Printf("Inside CalculateCPU, pid: %d\n", opid)
    f, err := os.Open(fmt.Sprintf("/proc/%d/stat", cpid))
    if err != nil {
        //log.Println(err)
    }
    defer f.Close()
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
        //fmt.Printf("starttime: %f\n", starttime)

        total_time := utime + stime + cuttime + cstime
        //fmt.Printf("total cpu clocks: %f\n", total_time)
        //fmt.Printf("total cpu usage in hertz: %f\n", total_time/hertz)

        t := time.Now()
        pr:= CPU{}
        pr.dateTime = t.Format(time.RFC3339)
        pr.pid = pid
        totalCPU += total_time/hertz
        pr.cpu = totalCPU
        Processes = append(Processes,pr)
    }
    if err := r.Err(); err != nil {
        //log.Println(err)
    }
}

func CalculateMemory(ind int, cpid int) {
    f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", cpid))
    if err != nil {
        //log.Println(err)
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
                //log.Println(err)
            }
            res += size
        }
    }
    if err := r.Err(); err != nil {
        //log.Println(err)
    }
    totalMem += res;
    t := time.Now()
    rs := Memory{}
    //rs.memoryUsed = res
    rs.memoryUsed = totalMem
    rs.dateTime = t.Format(time.RFC3339)
    rs.pid = pid
    Resources = append(Resources,rs)
}

func TimeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s\n", name, elapsed)
}

func takeSnapshots() {
    sleep_time := 15
    for i:=0;i<100;i++ {
        totalMem = 0
        totalCPU = 0
        go getChildPids()
        if i > 10 {
            sleep_time = 30
        }
        //lop for pids
        for index, child_pid := range Pids {
            go CalculateCPU(index, child_pid)
            go CalculateMemory(1, child_pid)
            time.Sleep(10 * time.Millisecond)
        }
        // go CalculateMemory(1, pid)
        time.Sleep(time.Duration(sleep_time) * time.Millisecond)
    }
}

func getChildPids() {
    if pid == 0 || pid == 1 { return }
    var cpid int
    //out, err := exec.Command("sh", "-c", fmt.Sprintf("pgrep -P %d", pid)).Output()
    out, err := exec.Command("sh", "-c", fmt.Sprintf("ps xh -o pgrp,pid | awk '$1==%d{print $2}'", PGID)).Output()
    if err != nil {
        //log.Fatal(err)
    }
    //fmt.Printf("Output: %s\n", out)
    temp := strings.Split(string(out),"\n")

    for index, child_pid := range temp {
        _, err1 := fmt.Sscanf(string(child_pid), "%d", &cpid)
        if err1 != nil && index > 0 {
            //log.Fatal(err1)
        }
        if cpid != 0 && cpid != 1 {
            Pids = append(Pids,cpid)
            Pids = unique.Ints(Pids)
        }
    }
}

func getParentGroupId() {
    if pid == 0 || pid == 1 { return }
    out, err := exec.Command("sh", "-c", fmt.Sprintf("ps xh -o pgrp,pid | awk '$2==%d{print $1}'", pid)).Output()
    if err != nil {
        //log.Fatal(err)
    }
    fmt.Printf("Parent Groupid: %s\n", out)
    _, err1 := fmt.Sscanf(string(out), "%d", &PGID)
    if err1 != nil {
        //log.Fatal(err)
    }    
}

func getCPUHZ() {
    line := float64(0)
    out, err := exec.Command("sh", "-c", "lscpu | grep -m1 MHz").Output()
    if err != nil {
        log.Println(err)
    }
    _, err1 := fmt.Sscanf(string(out[8:]), "%f", &line)
    if err1 != nil {
        log.Println(err1)
    }
    hertz = line*1000000
}

func killProcess() {
    fmt.Println("Killing process");
    for index, child_pid := range Pids {
        out, err := exec.Command("sh", "-c", fmt.Sprintf("kill %d", child_pid)).Output()
        if err != nil && index > 0 {
            //log.Fatal(err)
        }
        fmt.Printf("%s",out);
    }
}

// func getChildPids() {
//     //f, err := os.Open(fmt.Sprintf("sudo ps xh -o pgrp,pid | awk '$1==%d{print $2}'", pid))
//     fmt.Printf("pgrep -P %d\n", pid)
//     f, err := os.Open(fmt.Sprintf("pgrep -P %d", pid))
//     if err != nil {
//         //log.Println(err)
//     }
//     defer f.Close()
//     r := bufio.NewScanner(f)
//     for r.Scan() {
//         line := r.Bytes()
//         fmt.Println("--------------------------")
//         fmt.Println(line)
//         var cpid int
//         _, err := fmt.Sscanf(string(line), "%d", &cpid)
//         if err != nil {
//             //log.Println(err)
//         }
//         fmt.Printf("cpids: %d\n", cpid)
//         Pids = append(Pids,cpid)
//     }
//     fmt.Println(Pids)
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