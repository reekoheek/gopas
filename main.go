package main

// /**
//  * Install action
//  */
// func actionInstall() {
// 	dependencies := readDependencies(options.GopasFile)
// 	for _, dep := range dependencies {
// 		cmd := exec.Command("go", "get", dep.Name)
// 		env := []string{fmt.Sprintf("GOPATH=%s", cwd+"/.gopath")}
// 		for _, v := range os.Environ() {
// 			if !strings.HasPrefix(v, "GOPATH=") {
// 				env = append(env, v)
// 			}
// 		}
// 		cmd.Env = env
// 		cmd.Start()
// 		if err := cmd.Wait(); err != nil {
// 			fmt.Fprintf(os.Stderr, ">> Install Error: %s\n", err.Error())
// 			os.Exit(1)
// 		}
// 		fmt.Printf(">> Installing \"go get %s\"\n", dep.Name)
// 	}
// }

/**
 * Imports
 */
import (
	"os"

	"gopkg.in/urfave/cli.v1"
)

/**
 * main function
 */
func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}

	tool := &Tool{
		Project: &ProjectImpl{
			Cwd:        cwd,
			Watches:    []string{"."},
			Ignores:    []string{".git", ".gopath"},
			Extensions: []string{"go"},
		},
	}

	tool.Bootstrap()

	app := cli.NewApp()
	app.Name = "gopas"
	app.Usage = "Go build tool outside GOPATH"
	app.Version = "0.1.0"

	app.Commands = []cli.Command{
		{
			Name:   "build",
			Usage:  "build project",
			Action: tool.DoBuild,
		},
		{
			Name:   "clean",
			Usage:  "clean gopath",
			Action: tool.DoClean,
		},
		{
			Name:   "install",
			Usage:  "install dependencies",
			Action: tool.DoInstall,
		},
		{
			Name:   "list",
			Usage:  "list dependencies",
			Action: tool.DoList,
		},
		{
			Name:   "run",
			Usage:  "run executable",
			Action: tool.DoRun,
		},
		{
			Name:   "test",
			Usage:  "test project",
			Action: tool.DoTest,
		},
	}
	app.Run(os.Args)

	//	if len(os.Args) > 1 {
	//		switch os.Args[1] {
	//		case "build":
	//			tool.DoBuild()
	//			return
	//		case "clean":
	//			tool.DoClean()
	//			return
	//		case "install":
	//			tool.DoInstall()
	//			return
	//		case "list":
	//			tool.DoList()
	//			return
	//		case "run":
	//			tool.DoRun()
	//			return
	//		case "test":
	//			tool.DoTest()
	//			return
	//		}
	//	}
	//	tool.DoHelp()
}
