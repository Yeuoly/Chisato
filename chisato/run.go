package chisato

import (
	"errors"
	"strconv"
	"strings"
)

var work_path string = "/home/ctf/"

func parseResult(result string) (uint64, string, error) {
	//with header execute_time:123456 ##result
	//try to parse result
	//if not match, return error
	//if match, return execute_time and result
	results := strings.Split(result, "##")
	if len(results) != 2 {
		return 0, "", errors.New("1: result format error, your program may be crashed or without any output")
	}
	//check if key execute_time exists
	if !strings.HasPrefix(results[0], "execute_time:") {
		return 0, "", errors.New("2: result format error, your program may be crashed")
	}
	//parse execute_time
	execute_time_text := results[0][13:]
	execute_time_text = execute_time_text[:len(execute_time_text)-1]
	execute_time, err := strconv.ParseUint(execute_time_text, 10, 64)
	if err != nil {
		return 0, "", errors.New("3: result format error, your program may be crashed")
	}
	//tirm line break
	strings.TrimRight(results[1], "\n")
	return execute_time, results[1], nil
}

func RunC(exec_path string, stdin string) (uint64, string, error) {
	//run docker container
	out, err := DockerRunElf(exec_path, stdin)
	if err != nil {
		return 0, "", err
	}
	return parseResult(out)
}

func RunCpp(exec_path string, stdin string) (uint64, string, error) {
	//run docker container
	out, err := DockerRunElf(exec_path, stdin)
	if err != nil {
		return 0, "", err
	}
	return parseResult(out)
}

func RunPython2(exec_path string, stdin string) (uint64, string, error) {
	//run docker container
	out, err := DockerRunElf("python2", stdin, exec_path)
	if err != nil {
		return 0, "", err
	}
	return parseResult(out)
}

func RunPython3(exec_path string, stdin string) (uint64, string, error) {
	//run docker container
	out, err := DockerRunElf("python3", stdin, exec_path)
	if err != nil {
		return 0, "", err
	}
	return parseResult(out)
}
