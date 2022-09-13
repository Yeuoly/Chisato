package chisato

import (
	"encoding/json"
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
)

type ChisatoTestingServer struct {
	znet.BaseRouter
}

var global Chisato

func (root Chisato) Run() {
	server := znet.NewServer()
	server.AddRouter(MESSAGEID_TESTING, &ChisatoTestingServer{})
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

func RunTesting(testcase []ChisatoTestcase, callback func(string, string) (string, bool)) []ChisatoTestcaseResult {
	var testcase_result []ChisatoTestcaseResult
	for _, testcase := range testcase {
		result, pass := callback(testcase.Stdin, testcase.Stdout)
		testcase_result = append(testcase_result, ChisatoTestcaseResult{
			Result: result,
			Pass:   pass,
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
	tmp_path += strconv.FormatInt(time.Hour.Nanoseconds(), 16)

	//test language
	switch request.Language {
	case "c":
		//compile
		exec_path, err = CompileC(tmp_path, request.Code)
		if err != nil {
			return failed(err.Error())
		}
		//run
		testcase_result = RunTesting(request.Testcase, func(stdin string, stdout string) (string, bool) {
			result, err := RunC(exec_path, stdin)
			if err != nil {
				return err.Error(), false
			}
			return result, result == stdout
		})
	case "cpp":
		//compile
		exec_path, err = CompileCpp(tmp_path, request.Code)
		if err != nil {
			return failed(err.Error())
		}
		//run
		testcase_result = RunTesting(request.Testcase, func(stdin string, stdout string) (string, bool) {
			result, err := RunCpp(exec_path, stdin)
			if err != nil {
				return err.Error(), false
			}
			return result, result == stdout
		})
	case "go":
	case "python2":
		exec_path, err = CompilePython2(tmp_path, request.Code)
		if err != nil {
			return failed(err.Error())
		}
		//run
		testcase_result = RunTesting(request.Testcase, func(stdin string, stdout string) (string, bool) {
			result, err := RunPython2(exec_path, stdin)
			if err != nil {
				return err.Error(), false
			}
			return result, result == stdout
		})
	case "python3":
		exec_path, err = CompilePython3(tmp_path, request.Code)
		if err != nil {
			return failed(err.Error())
		}
		testcase_result = RunTesting(request.Testcase, func(stdin string, stdout string) (string, bool) {
			result, err := RunPython3(exec_path, stdin)
			if err != nil {
				return err.Error(), false
			}
			return result, result == stdout
		})
	case "java":
	case "node":
	}

	return success(testcase_result)
}
