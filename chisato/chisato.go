package chisato

import (
	"encoding/json"
	"path"
	"strconv"
	"time"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

type Chisato struct {
	secret     string
	heart_beat time.Duration
}

type ChisatoAuth struct {
	secret string
}

type ChisatoTestcase struct {
	Stdin  string `json:"stdin"`
	Stdout string `json:"stdout"`
}

type ChisatoTestcaseResult struct {
	Result string `json:"result"`
	Pass   bool   `json:"pass"`
	Mem    int64  `json:"mem"`
	Time   int64  `json:"time"`
}

type ChisatoRequestTesting struct {
	Testcase []ChisatoTestcase `json:"testcase"`
	Code     string            `json:"code"`
	Language string            `json:"language"`
}

type ChisatoResponse struct {
	Res  int    `json:"res"`
	Err  string `json:"err"`
	Data any    `json:"data"`
}

func success(data any) ChisatoResponse {
	return ChisatoResponse{
		Res:  0,
		Err:  "",
		Data: data,
	}
}

func failed(err string) ChisatoResponse {
	return ChisatoResponse{
		Res: -1,
		Err: err,
	}
}

const (
	MESSAGEID_CONFIG  = 1
	MESSAGEID_TESTING = 2
	MESSAGEID_ADMIN   = 3
	MESSAGEID_PING    = 4
)

type ChisatoTestingServer struct {
	znet.BaseRouter
}

type ChisatoPingServer struct {
	znet.BaseRouter
}

var global Chisato
var docker_work_dir = "/home/ctf/"

func (root Chisato) Run() {
	server := znet.NewServer()
	server.AddRouter(MESSAGEID_TESTING, &ChisatoTestingServer{})
	server.AddRouter(MESSAGEID_PING, &ChisatoPingServer{})
	server.Serve()
}

func GetChisato() Chisato {
	return global
}

func (router *ChisatoTestingServer) Handle(req ziface.IRequest) {
	var request ChisatoRequestTesting
	//unmarshal
	err := json.Unmarshal(req.GetData(), &request)
	if err != nil {
		req.GetConnection().SendBuffMsg(MESSAGEID_TESTING, []byte("failed to unmarshal"))
	}

	response := GetChisato().Testing(request)
	//marshal
	text, _ := json.Marshal(response)
	req.GetConnection().SendBuffMsg(MESSAGEID_TESTING, text)
}

func (router *ChisatoPingServer) Handle(req ziface.IRequest) {
	req.GetConnection().SendBuffMsg(MESSAGEID_PING, []byte("pong"))
}

func RunTesting(testcase []ChisatoTestcase, callback func(string, string) (uint64, uint64, string, bool)) []ChisatoTestcaseResult {
	var testcase_result []ChisatoTestcaseResult

	for _, testcase := range testcase {
		execute_time, execute_memory, result, pass := callback(testcase.Stdin, testcase.Stdout)
		testcase_result = append(testcase_result, ChisatoTestcaseResult{
			Result: result,
			Pass:   pass,
			Time:   int64(execute_time),
			Mem:    int64(execute_memory),
		})
	}
	return testcase_result
}

func (root Chisato) Testing(request ChisatoRequestTesting) ChisatoResponse {
	var exec_path string
	var err error
	var tmp_path string

	var testcase_result []ChisatoTestcaseResult

	//create tmp path
	tmp_path = "./tmp/" + strconv.FormatInt(time.Now().UnixMilli(), 16)
	tmp_path += strconv.FormatInt(time.Now().UnixNano(), 16)

	//test language
	switch request.Language {
	case "c":
		//compile
		exec_path, err = CompileC(tmp_path, request.Code)
		if err != nil {
			return failed(err.Error())
		}
		//copy file to docker
		docker_path := docker_work_dir + exec_path
		docker_path = path.Clean(docker_path)
		err = DockerCopyFileFromLocal(exec_path, docker_path)
		if err != nil {
			return failed(err.Error())
		}
		//run
		testcase_result = RunTesting(request.Testcase, func(stdin string, stdout string) (uint64, uint64, string, bool) {
			execute_time, execute_memory, result, err := RunC(docker_path, stdin)
			if err != nil {
				return 0, 0, err.Error(), false
			}
			return execute_time, execute_memory, result, result == stdout
		})
	case "cpp":
		//compile
		exec_path, err = CompileCpp(tmp_path, request.Code)
		if err != nil {
			return failed(err.Error())
		}
		//copy file to docker
		docker_path := docker_work_dir + exec_path
		docker_path = path.Clean(docker_path)
		err = DockerCopyFileFromLocal(exec_path, docker_path)
		if err != nil {
			return failed(err.Error())
		}
		//run
		testcase_result = RunTesting(request.Testcase, func(stdin string, stdout string) (uint64, uint64, string, bool) {
			execute_time, execute_memory, result, err := RunCpp(docker_path, stdin)
			if err != nil {
				return 0, 0, err.Error(), false
			}
			return execute_time, execute_memory, result, result == stdout
		})
	case "go":
		//compile
		exec_path, err = CompileGolang(tmp_path, request.Code)
		if err != nil {
			return failed(err.Error())
		}
		//copy file to docker
		docker_path := docker_work_dir + exec_path
		docker_path = path.Clean(docker_path)
		err = DockerCopyFileFromLocal(exec_path, docker_path)
		if err != nil {
			return failed(err.Error())
		}
		//run
		testcase_result = RunTesting(request.Testcase, func(stdin string, stdout string) (uint64, uint64, string, bool) {
			execute_time, execute_memory, result, err := RunGolang(docker_path, stdin)
			if err != nil {
				return 0, 0, err.Error(), false
			}
			return execute_time, execute_memory, result, result == stdout
		})
	case "python2":
		exec_path, err := CompilePython2(tmp_path, request.Code)
		if err != nil {
			return failed(err.Error())
		}
		//copy file to docker
		docker_path := docker_work_dir + exec_path
		docker_path = path.Clean(docker_path)
		err = DockerCopyFileFromLocal(exec_path, docker_path)
		if err != nil {
			return failed(err.Error())
		}
		//run
		testcase_result = RunTesting(request.Testcase, func(stdin string, stdout string) (uint64, uint64, string, bool) {
			execute_time, execute_memory, result, err := RunPython2(docker_path, stdin)
			if err != nil {
				return 0, 0, err.Error(), false
			}
			return execute_time, execute_memory, result, result == stdout
		})
	case "python3":
		exec_path, err := CompilePython3(tmp_path, request.Code)
		if err != nil {
			return failed(err.Error())
		}
		//copy file to docker
		docker_path := docker_work_dir + exec_path
		docker_path = path.Clean(docker_path)
		err = DockerCopyFileFromLocal(exec_path, docker_path)
		if err != nil {
			return failed(err.Error())
		}
		//run
		testcase_result = RunTesting(request.Testcase, func(stdin string, stdout string) (uint64, uint64, string, bool) {
			execute_time, execute_memory, result, err := RunPython3(docker_path, stdin)
			if err != nil {
				return 0, 0, err.Error(), false
			}
			return execute_time, execute_memory, result, result == stdout
		})
	case "java":
		exec_path, err := CompileJava(tmp_path, request.Code)
		if err != nil {
			return failed(err.Error())
		}
		//copy file to docker
		docker_path := docker_work_dir + exec_path
		docker_path = path.Clean(docker_path)
		err = DockerCopyFileFromLocal(exec_path, docker_path)
		if err != nil {
			return failed(err.Error())
		}
		//run
		testcase_result = RunTesting(request.Testcase, func(stdin string, stdout string) (uint64, uint64, string, bool) {
			execute_time, execute_memory, result, err := RunJava(docker_path, stdin)
			if err != nil {
				return 0, 0, err.Error(), false
			}
			return execute_time, execute_memory, result, result == stdout
		})
	case "node":
	}

	return success(testcase_result)
}
