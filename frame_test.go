package jewl

import (
	"fmt"
	//"math/rand"
	"reflect"
	"testing"
)

func bubbleSort(data []int) ([]int, error) {
	rec, err := GetRecorder(JewlConfig)
	if err != nil {
		panic(err)
	}
	err = rec.Frame("bubbleSort")

	rec.AddData("data", data)
	for i := 0; i < len(data); i++ {
		err = rec.Frame("outer_loop")
		if err != nil {
			panic(err)
		}
		rec.AddData("i", i)
		for j := 0; j < len(data); j++ {
			err = rec.Frame("inner_loop")
			if err != nil {
				panic(err)
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
			err = rec.Stop()
            if err != nil{
                panic(err)
            }
		}
        err = rec.Stop()
        if err != nil{
            panic(err)
        }
	}
    err = rec.Stop()
    if err != nil{
        panic(err)
    }
	return data, nil
}

const FileLocation = "bubble_sort"

var JewlConfig = FileConfig(FileLocation)

func TestFrame(t *testing.T) {
	rec, err := NewRecorder(JewlConfig)
	defer rec.Close()
	if err != nil {
		panic(err)
	}
	err = rec.Frame("TestFrame")
	sorted, err := bubbleSort([]int{5, 1, 6, 4, 3, 7, 8, 9})
	if err != nil {
		panic(err)
	}

	expected := []int{1, 3, 4, 5, 6, 7, 8, 9}
	if !reflect.DeepEqual(expected, sorted) {
		fmt.Println(fmt.Sprintf("Sort result is not as expected: %s", fmt.Sprint(sorted)))
	}
	err = rec.Stop()
    if err != nil{
        panic(err)
    }
}

//func TestFrameMany(t *testing.T) {
//    rec, err := NewRecorder(JewlConfig)
//	defer rec.Close()
//	if err != nil {
//		panic(err)
//	}
//	err = rec.Frame("TestFrame")
//    for i := 0; i < 10; i++{
//        sorted, err := bubbleSort(rand.Perm(10))
//        if err != nil{
//            panic(err)
//        }
//        println(sorted)
//    }
//	rec.Stop()
//}
