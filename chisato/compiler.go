package chisato

import (
	"os"
	"os/exec"
)

func CompileC(path string, code string) (command string, err error) {
	//open path, if not exist, create it
	//write code to path
	//compile code

	//open
	err = os.MkdirAll(path, 0777)
	if err != nil {
		return
	}

	source_path := path + "/main.c"
	exec_path := path + "/main"
	f, err := os.Create(source_path)
	if err != nil {
		return
	}

	//write
	_, err = f.WriteString(code)
	if err != nil {
		return
	}

	//compile
	//gcc -o main main.c
	cmd := exec.Command("gcc", "-o", exec_path, source_path)
	err = cmd.Run()
	if err != nil {
		return
	}

	command = path + "/main"
	return
}

func CompileCpp(path string, code string) (command string, err error) {
	//open path, if not exist, create it
	//write code to path
	//compile code

	//open
	err = os.MkdirAll(path, 0777)
	if err != nil {
		return
	}

	source_path := path + "/main.cpp"
	exec_path := path + "/main"
	f, err := os.Create(source_path)
	if err != nil {
		return
	}

	//write
	_, err = f.WriteString(code)
	if err != nil {
		return
	}

	//compile
	//g++ -o main main.cpp
	cmd := exec.Command("g++", "-o", exec_path, source_path)
	err = cmd.Run()
	if err != nil {
		return
	}

	command = path + "/main"
	return
}

//only create file, not compile
func CompileScriptLanguage(path string, code string) (command string, err error) {
	//open path, if not exist, create it
	//write code to path

	//open
	err = os.MkdirAll(path, 0777)
	if err != nil {
		return
	}

	source_path := path + "/main.py"
	f, err := os.Create(source_path)
	if err != nil {
		return
	}

	//write
	_, err = f.WriteString(code)
	if err != nil {
		return
	}

	command = source_path
	return
}

func CompilePython2(path string, code string) (command string, err error) {
	return CompileScriptLanguage(path, code)
}

func CompilePython3(path string, code string) (command string, err error) {
	return CompileScriptLanguage(path, code)
}
