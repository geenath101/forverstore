package main

import (
	"fmt"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "bestpictures"
	pathName := CASPathTransformFunc(key)
	expectedPathName := "b2f8f/1dd50/fdeec/113ac/1b506/6d1d3/a10f7/0f1dc"
	if pathName != expectedPathName {
		t.Errorf("issue with path name , please check path creating logic")
	}
	fmt.Printf("pathName is %s", pathName)
}

// func TestStore(t *testing.T) {
// 	opts := StoreOpts{
// 		PathTransformFunc: DefaultPathTransformFunc,
// 	}
// 	s := NewStore(opts)
// }
