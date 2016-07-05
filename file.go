package main

// func copy_folder(source string, dest string) (err error) {

//  sourceinfo, err := os.Stat(source)
//  if err != nil {
//    return err
//  }

//  err = os.MkdirAll(dest, sourceinfo.Mode())
//  if err != nil {
//    return err
//  }

//  directory, _ := os.Open(source)

//  objects, err := directory.Readdir(-1)

//  for _, obj := range objects {

//    sourcefilepointer := source + "/" + obj.Name()

//    destinationfilepointer := dest + "/" + obj.Name()

//    if obj.IsDir() {
//      if obj.Name() != ".gopath" {
//        err = copy_folder(sourcefilepointer, destinationfilepointer)
//        if err != nil {
//          fmt.Println(err)
//        }
//      }
//    } else {
//      err = copy_file(sourcefilepointer, destinationfilepointer)
//      if err != nil {
//        fmt.Println(err)
//      }
//    }

//  }
//  return
// }

// func copy_file(source string, dest string) (err error) {
//  sourcefile, err := os.Open(source)
//  if err != nil {
//    return err
//  }

//  defer sourcefile.Close()

//  destfile, err := os.Create(dest)
//  if err != nil {
//    return err
//  }

//  defer destfile.Close()

//  _, err = io.Copy(destfile, sourcefile)
//  if err == nil {
//    sourceinfo, err := os.Stat(source)
//    if err != nil {
//      err = os.Chmod(dest, sourceinfo.Mode())
//    }

//  }

//  return
// }
