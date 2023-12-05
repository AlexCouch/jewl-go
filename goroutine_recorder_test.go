package jewl

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func bubbleSort2(data []int, dataOut chan<-[]int, errorChan chan<-error) {
	rec, err := GetRecorder(JewlConfig)
	if err != nil {
		panic(err)
	}
	err = rec.Frame("bubbleSort")

	rec.AddData("data", data)
	for i := 0; i < len(data); i++ {
		err = rec.Frame("outer_loop")
		if err != nil {
            errorChan<-err
            rec.Stop()
            break
		}
		rec.AddData("i", i)
		for j := 0; j < len(data); j++ {
			err = rec.Frame("inner_loop")
			if err != nil {
                errorChan<-err
                rec.Stop()
                break
			}
			rec.AddData("j", j)
			rec.AddData("data[j]", data[j])
			rec.AddData("j+1", j+1)
			if j+1 >= len(data) {
				rec.Stop()
				continue
			}
			rec.AddData("data[j+1]", data[j+1])
			if data[j] > data[j+1] {
				temp := data[j]
				data[j] = data[j+1]
				data[j+1] = temp
			}
			rec.Stop()
		}
		rec.Stop()
	}
	rec.Stop()
    dataOut<-data
}

func waitForBubbleSorts(wg *sync.WaitGroup, dataOut chan []int, errorChan chan error) {
    wg.Wait()
    close(errorChan)
    close(dataOut)
}

const fileLocation = "bubble_sort2"

var jewlConfig = FileConfig(fileLocation)

func TestGoroutines(t *testing.T) {
	rec, err := NewRecorder(jewlConfig)
	defer rec.Close()
	if err != nil {
		panic(err)
	}
	err = rec.Frame("TestFrame")
    dataOut := make(chan []int)
    errorChan := make(chan error)
    var wg sync.WaitGroup
    wg.Add(10)
    for i := 0; i < 10; i++ {
        go func(){
            defer wg.Done()
            bubbleSort2([]int{5, 1, 6, 4, 3, 7, 8, 9}, dataOut, errorChan)
        }()
    }
    go waitForBubbleSorts(&wg, dataOut, errorChan)
    for out := range dataOut{
        expected := []int{1, 3, 4, 5, 6, 7, 8, 9}
        if !reflect.DeepEqual(expected, out) {
            fmt.Println(fmt.Sprintf("Sort result is not as expected: %s", fmt.Sprint(out)))
        }
    }
    for err := range errorChan{
        println(err)
    }

	rec.Stop()
}
