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

	err = cmd.Wait()
	if err != nil && err.Error() != "exec: Wait was already called" {
		return
	}
	err = nil

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

	err = cmd.Wait()
	if err != nil && err.Error() != "exec: Wait was already called" {
		return
	}
	err = nil

	command = path + "/main"
	return
}

func CompileGolang(path string, code string) (command string, err error) {
	//open path, if not exist, create it
	//write code to path
	//compile code

	//open
	err = os.MkdirAll(path, 0777)
	if err != nil {
		return
	}

	source_path := path + "/main.go"
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
	cmd := exec.Command("go", "build", "-o", exec_path, source_path)
	err = cmd.Run()
	if err != nil {
		return
	}

	err = cmd.Wait()
	if err != nil && err.Error() != "exec: Wait was already called" {
		return
	}
	err = nil

	command = path + "/main"
	return
}

func CompileJava(path string, code string) (command string, err error) {
	//open path, if not exist, create it
	//write code to path
	//compile code

	//open
	err = os.MkdirAll(path, 0777)
	if err != nil {
		return
	}

	source_path := path + "/" + JAVA_DEFAULT_CLASS + ".java"
	exec_path := path + "/" + JAVA_DEFAULT_PATH
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
	cmd := exec.Command("javac", source_path)

	err = cmd.Run()
	if err != nil {
		return
	}

	err = cmd.Wait()
	if err != nil && err.Error() != "exec: Wait was already called" {
		return
	}
	err = nil

	// mkdir -p cn/srmxy/chisato/main
	// mv Main.class cn/srmxy/chisato/main
	cmd = exec.Command("mkdir", "-p", exec_path)
	err = cmd.Run()
	if err != nil {
		return
	}

	if err != nil && err.Error() != "exec: Wait was already called" {
		return
	}
	err = nil

	class_file := path + "/" + JAVA_DEFAULT_CLASS + ".class"

	cmd = exec.Command("mv", class_file, exec_path)
	err = cmd.Run()
	if err != nil {
		return
	}

	err = cmd.Wait()
	if err != nil && err.Error() != "exec: Wait was already called" {
		return
	}
	err = nil

	command = path
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
