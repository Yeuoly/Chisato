package chisato

import (
	"fmt"
	"os/exec"
	"strings"
)

func RunC(exec_path string, stdin string) (string, error) {
	cmd := exec.Command(exec_path)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// erease the last \n if exist
	if out[len(out)-1] == 10 {
		out = out[:len(out)-1]
	}
	return string(out), nil
}

func RunCpp(exec_path string, stdin string) (string, error) {
	cmd := exec.Command(exec_path)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// erease the last \n if exist
	if out[len(out)-1] == 10 {
		out = out[:len(out)-1]
	}
	return string(out), nil
}

func RunPython2(exec_path string, stdin string) (string, error) {
	fmt.Println(exec_path)
	cmd := exec.Command("python2", exec_path)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// erease the last \n if exist
	if out[len(out)-1] == 10 {
		out = out[:len(out)-1]
	}

	return string(out), nil
}

func RunPython3(exec_path string, stdin string) (string, error) {
	cmd := exec.Command("python3", exec_path)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// erease the last \n if exist
	if out[len(out)-1] == 10 {
		out = out[:len(out)-1]
	}
	return string(out), nil
}
