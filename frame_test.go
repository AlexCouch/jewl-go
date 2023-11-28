package jewl

import (
	"fmt"
	"reflect"
	"testing"
)

func bubbleSort(data []uint16) []uint16{
    frame := G_Profiler.StartFrame("bubbleSort")
    defer frame.Stop()

    frame.AddArg("data", data)
    for i := 0; i < len(data); i++{
        outer_loop := frame.Subframe("outer_loop")
        outer_loop.AddArg("i", i)
        for j := 0; j < len(data); j++{
            inner_loop := outer_loop.Subframe("inner_loop")
            inner_loop.AddArg("j", j)
            inner_loop.AddArg("data[j]", data[j])
            inner_loop.AddArg("j+1", j+1)
            if j + 1 >= len(data){
                continue
            }
            inner_loop.AddArg("data[j+1]", data[j+1])
            if data[j] > data[j + 1]{
                temp := data[j]
                data[j] = data[j+1]
                data[j+1] = temp
            }
            inner_loop.Stop()
        }
        outer_loop.Stop()
    }
    return data
}

func TestFrame(t *testing.T){
    frame := G_Profiler.StartFrame("TestFrame")
    sorted := bubbleSort([]uint16{5, 1, 6, 4, 3, 7, 8, 9})
    expected := []uint16{1, 3, 4, 5, 6, 7, 8, 9}
    if !reflect.DeepEqual(expected, sorted){
        fmt.Println(fmt.Sprintf("Sort result is not as expected: %s", fmt.Sprint(sorted)))
    }
    frame.Stop()
    err := G_Profiler.Dump("dump.json")
    if err != nil{
        panic(err)
    }
}
