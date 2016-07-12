package util

import (
	"io"
	"os"
)

func copy_folder(source string, dest string) error {
	var (
		sourceinfo os.FileInfo
		err        error
	)

	if sourceinfo, err = os.Stat(source); err != nil {
		return err
	}

	if err = os.MkdirAll(dest, sourceinfo.Mode()); err != nil {
		return err
	}

	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourcefilepointer := source + "/" + obj.Name()

		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			if obj.Name() != ".gopath" {
				if err = copy_folder(sourcefilepointer, destinationfilepointer); err != nil {
					return err
				}
			}
		} else if err = copy_file(sourcefilepointer, destinationfilepointer); err != nil {
			return err
		}
	}

	return err
}

func copy_file(source string, dest string) error {
	var (
		sourcefile *os.File
		err        error
		destfile   *os.File
	)

	if sourcefile, err = os.Open(source); err != nil {
		return err
	}

	defer sourcefile.Close()

	if destfile, err = os.Create(dest); err != nil {
		return err
	}

	defer destfile.Close()

	if _, err = io.Copy(destfile, sourcefile); err == nil {
		if sourceinfo, err := os.Stat(source); err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}

	return err
}
