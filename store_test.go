package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	s.Clear()
}

func TestPathTransformFunc(t *testing.T) {
	key := "bestpictures"
	pathKey := CASPathTransformFunc(key)
	expectedFileName := "b2f8f1dd50fdeec113ac1b5066d1d3a10f70f1dc"
	expectedPathName := "b2f8f/1dd50/fdeec/113ac/1b506/6d1d3/a10f7/0f1dc"
	if pathKey.PathName != expectedPathName {
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathName)
	}
	if pathKey.FileName != expectedFileName {
		t.Errorf("have %s want %s", pathKey.FileName, expectedFileName)
	}
	//fmt.Printf("pathName is %s", pathName)
}

func TestStoreDeleteKey(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "momsspecials"
	data := []byte("some jpg bytes")

	//if err := s.writeStream("bestpictures", data); err != nil {
	//	t.Error(err)
	//}
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	err := s.Delete(key)
	if err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	//opts := StoreOpts{
	//	PathTransformFunc: CASPathTransformFunc,
	//}
	s := newStore()
	defer teardown(t, s)
	key := "momsspecials"
	data := []byte("some jpg bytes")

	//if err := s.writeStream("bestpictures", data); err != nil {
	//	t.Error(err)
	//}
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := ioutil.ReadAll(r)

	fmt.Println(string(b))

	if string(b) != string(data) {
		t.Errorf("want %s have %s", data, b)
	}
}
