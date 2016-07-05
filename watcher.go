package main

// /**
//  * ScanCallback type
//  */
// type ScanCallback func(path string)

// /**
//  * Scan changes and invoke callback on file changes
//  *
//  * @param  {ScanCallback}   cb ScanCallback
//  */
// func scanChanges(cb ScanCallback) {
// 	for {
// 		for _, dir := range options.Watches {
// 			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
// 				if isIgnorable(path) {
// 					return filepath.SkipDir
// 				}

// 				if isAcceptable(path) && info.ModTime().After(modifiedTime) {
// 					modifiedTime = time.Now()
// 					cb(path)
// 				}
// 				return nil
// 			})
// 		}

// 		time.Sleep(1000 * time.Millisecond)
// 	}
// }

// /**
//  * Check whether path is ignorable
//  *
//  * @param {string} path
//  */
// func isIgnorable(path string) bool {
// 	ignorable := false
// 	for _, ignore := range options.Ignores {
// 		if matched, _ := filepath.Match(ignore, path); matched {
// 			ignorable = true
// 			break
// 		}
// 	}
// 	return ignorable
// }

// /**
//  * Check whether path having acceptable extension
//  *
//  * @param {string} path
//  */
// func isAcceptable(path string) bool {
// 	for _, ext := range options.Extensions {
// 		if strings.HasSuffix(path, "."+ext) {
// 			return true
// 		}
// 	}

// 	return false
// }
