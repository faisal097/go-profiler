package profiler

/*
	Time tracker will take two parameters time now and function name in which it is called
*/

import (
	"fmt"
	"time"
)


func TimeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s\n", name, elapsed)
}