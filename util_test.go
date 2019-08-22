package go2chef

import (
	"io/ioutil"
	"log"
	"testing"
)

func doesFunctionPanic(f func()) bool {
	var rec interface{}

	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("captured panic: %#v", r)
				rec = r
			}
		}()

		f()
	}()

	return rec != nil
}

func createTempFile(t *testing.T, content string) string {
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("error creating tempfile: %s", err)
	}
	defer tf.Close()
	if _, err := tf.WriteString(`{"key":"value"}`); err != nil {
		t.Fatalf("failed to write tempfile: %s", err)
	}
	return tf.Name()
}
