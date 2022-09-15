package chisato

import (
	"encoding/binary"
	"errors"
)

var work_path string = "/home/ctf/"

func parseResult(result string) (uint64, uint64, string, error) {
	//result has a header total 8*4 bytes
	if len(result) <= 20 {
		return 0, 0, "", errors.New("runner master returns a error format output")
	}
	//read header
	header_bytes := []byte(result[:20])
	execute_time := binary.LittleEndian.Uint64(header_bytes[0:8])
	execute_memory := binary.LittleEndian.Uint64(header_bytes[8:16])
	body_length := binary.LittleEndian.Uint64(header_bytes[16:24])
	has_error := binary.LittleEndian.Uint64(header_bytes[24:32])

	body := result[32:]
	if len(body) != int(body_length) {
		return 0, 0, "", errors.New("runner master returns a error format output")
	}

	if has_error != 0 {
		return execute_time, execute_memory, "", errors.New(body)
	}

	return execute_time, execute_memory, body, nil
}

func RunC(exec_path string, stdin string) (uint64, uint64, string, error) {
	//run docker container
	out, err := DockerRunElf(exec_path, stdin)
	if err != nil {
		return 0, 0, "", err
	}
	return parseResult(out)
}

func RunCpp(exec_path string, stdin string) (uint64, uint64, string, error) {
	//run docker container
	out, err := DockerRunElf(exec_path, stdin)
	if err != nil {
		return 0, 0, "", err
	}
	return parseResult(out)
}

func RunPython2(exec_path string, stdin string) (uint64, uint64, string, error) {
	//run docker container
	out, err := DockerRunElf("python2", stdin, exec_path)
	if err != nil {
		return 0, 0, "", err
	}
	return parseResult(out)
}

func RunPython3(exec_path string, stdin string) (uint64, uint64, string, error) {
	//run docker container
	out, err := DockerRunElf("python3", stdin, exec_path)
	if err != nil {
		return 0, 0, "", err
	}
	return parseResult(out)
}
