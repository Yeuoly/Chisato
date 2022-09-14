package chisato

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var global_chisato_id string

func init() {
	//docker check if yeuoly/chisato:v1 image exists
	//if not, pull it
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	fmt.Println("checking if chisato image exists...")
	_, _, err = cli.ImageInspectWithRaw(context.Background(), "yeuoly/chisato:v1")
	if err != nil {
		out, err := cli.ImagePull(context.Background(), "yeuoly/chisato:v1", types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}
		io.Copy(os.Stdout, out)
	}
	//check if there is a container named chisato
	//if not, create it
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	var chisato_container_id string
	for _, c := range containers {
		if c.Image == "yeuoly/chisato:v1" {
			chisato_container_id = c.ID
			break
		}
	}
	if chisato_container_id == "" {
		//create container
		fmt.Println("creating chisato container...")
		resp, err := cli.ContainerCreate(context.Background(), &container.Config{
			Image: "yeuoly/chisato:v1",
			Tty:   true,
			Cmd:   []string{"/bin/bash"},
		}, nil, nil, nil, "chisato")
		if err != nil {
			panic(err)
		}
		if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}
		chisato_container_id = resp.ID
		fmt.Println("chisato container created, id:", chisato_container_id)
	}
	global_chisato_id = chisato_container_id

	//compile shell/shell.c and copy it to docker
	fmt.Println("compiling docker runner master to -> shell/shell")
	cmd := exec.Command("bash", "-c", "gcc ./shell/shell.c -o ./shell/shell")
	cmd.Run()
	fmt.Println("copying docker runner master to docker -> /home/ctf/shell")
	DockerCopyFileFromLocal("./shell/shell", "/home/ctf/shell")
}

func DockerCopyFileFromLocal(local_path string, container_path string) error {
	//copy file from local to container
	f, err := os.Open(local_path)
	if err != nil {
		return err
	}
	defer f.Close()

	//split container_path to dir and file
	var dir string
	for i := len(container_path) - 1; i >= 0; i-- {
		if container_path[i] == '/' {
			dir = container_path[:i]
			break
		}
	}

	//create dir in container
	_, err = DockerRunCommand("mkdir -p "+dir, "")
	if err != nil {
		return err
	}

	//clean
	container_path = path.Clean(container_path)
	command := "docker cp " + local_path + " " + global_chisato_id + ":" + container_path

	//copy file to container
	cmd := exec.Command("/bin/bash", "-c", command)
	_, err = cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func DockerRunCommand(command string, stdin string) (result string, err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	//run command in container
	exec_config := types.ExecConfig{
		Cmd:          []string{"/bin/bash", "-c", command},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}
	resp, err := cli.ContainerExecCreate(context.Background(), global_chisato_id, exec_config)
	if err != nil {
		return "", err
	}
	hijack, err := cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", err
	}
	defer hijack.Close()
	hijack.Conn.Write([]byte(stdin))
	hijack.CloseWrite()
	result_bytes, err := io.ReadAll(hijack.Reader)
	if err != nil {
		return
	}
	result = string(result_bytes)
	return
}

//run a elf file without DockerRunCommand
func DockerRunElf(elf_path string, stdin string, args ...string) (result string, err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	//run command in container
	// /home/ctf/shell elf_path args
	args = append([]string{"/home/ctf/shell", elf_path}, args...)
	exec_config := types.ExecConfig{
		Cmd:          args,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		User:         "ctf",
	}

	resp, err := cli.ContainerExecCreate(context.Background(), global_chisato_id, exec_config)
	if err != nil {
		return "", err
	}

	hijack, err := cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", err
	}
	//set timeout
	hijack.Conn.SetDeadline(time.Now().Add(time.Minute))

	//write stdin
	go func() {
		defer hijack.CloseWrite()
		hijack.Conn.Write([]byte(stdin))
	}()

	//read
	result = ""
	last := make([]byte, 0)
	for {
		//check if last is already contain a complete output
		if len(last) >= 8 {
			//read length
			length := binary.BigEndian.Uint32(last[4:8])
			if len(last) >= 8+int(length) {
				result += string(last[8 : 8+length])
				last = last[8+length:]
				continue
			}
		}

		buf := make([]byte, 1024)
		n, err := hijack.Conn.Read(buf)
		if err != nil {
			break
		}
		//append buf to last
		if len(last) == 0 {
			last = buf[:n]
		} else {
			last = append(last, buf[:n]...)
		}
	}
	return
}
