package fileutil

import (
	"os"
	"strings"

	"github.com/alexflint/go-filemutex"
)

func Lock(file string) func() {
	file = file + ".lock"

	m, err := filemutex.New(file)
	if err != nil {
		panic(err)
	}

	err = m.Lock()
	if err != nil {
		panic(err)
	}

	return func() {
		errUnlock := m.Unlock()
		if errUnlock != nil {
			panic(err)
		}
	}
}

func MkdirAll(dirPath string) {
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func MkdirAllFromFile(filePath string) {
	idx := strings.LastIndex(filePath, "/")
	MkdirAll(filePath[:idx])
}

func WriteFile(filePath string, data []byte) {
	WriteFilePerm(filePath, data, os.ModePerm)
}

func WriteFilePerm(filePath string, data []byte, perm os.FileMode) {
	err := os.WriteFile(filePath, data, perm)
	if err != nil {
		panic(err)
	}
}

func ReadFile(filePath string) []byte {
	b, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return b
}

func Remove(filePath string) {
	err := os.Remove(filePath)
	if err != nil {
		panic(err)
	}
}
