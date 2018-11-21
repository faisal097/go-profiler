package profiler

import (
	"fmt"
	"os"
	"bufio"
	"bytes"
)

 	

// type Resources struct {
//     memoryUsed uint64
//     cpuUsed  int
// }

//type Resources map[string]interface{}


//https://stackoverflow.com/questions/31879817/golang-os-exec-realtime-memory-usage
//https://unix.stackexchange.com/questions/33381/getting-information-about-a-process-memory-usage-from-proc-pid-smaps
// func CalculateMemory(pid int) (uint64, error) {
//     f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", pid))
//     if err != nil {
//         return 0, err
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
//                 return 0, err
//             }
//             res += size
//         }
//     }
//     if err := r.Err(); err != nil {
//         return 0, err
//     }

//     return res, nil
// }

func CalculateMemory(pid int, rs *Resource, resAr *[]Resource) {
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
    resAr = append(resAr,rs)
    //rs := Resources{memoryUsed: res}
    //fmt.Println(rs)
    //fmt.Printf("Memory used: %d KB\n", res)
    //return res, nil
}

//must see: https://golang.org/doc/diagnostics.html