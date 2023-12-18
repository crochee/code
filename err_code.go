package code

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type innerErrCode struct {
	serviceName    string
	httpStatusCode int
	// 3(service)+4(error)
	code    string
	message string
	result  interface{}
}

func (e *innerErrCode) Error() string {
	return fmt.Sprintf("service_name:%s,http_status_code:%d,code:%s,message:%s,result:%v",
		e.serviceName, e.httpStatusCode, e.code, e.message, e.result)
}

func (e *innerErrCode) ServiceName() string {
	return e.serviceName
}

func (e *innerErrCode) StatusCode() int {
	return e.httpStatusCode
}

func (e *innerErrCode) Code() string {
	return e.code
}

func (e *innerErrCode) Message() string {
	return e.message
}

func (e *innerErrCode) Result() interface{} {
	return e.result
}

func (e *innerErrCode) WithStatusCode(statusCode int) ErrorCode {
	ec := *e
	ec.httpStatusCode = statusCode
	return &ec
}

func (e *innerErrCode) WithCode(code string) ErrorCode {
	ec := *e
	ec.code = code
	return &ec
}

func (e *innerErrCode) WithMessage(msg string) ErrorCode {
	ec := *e
	ec.message = msg
	return &ec
}

func (e *innerErrCode) WithResult(result interface{}) ErrorCode {
	ec := *e
	ec.result = result
	return &ec
}

func (e *innerErrCode) Is(v error) bool {
	err, ok := v.(ErrorCode)
	if !ok {
		return false
	}
	return err.Code() == e.Code()
}

func (e *innerErrCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Result  interface{} `json:"result"`
	}{
		Code:    fmt.Sprintf("%s.%3d%s", e.ServiceName(), e.StatusCode(), e.Code()),
		Message: e.Message(),
		Result:  e.Result(),
	})
}

func (e *innerErrCode) UnmarshalJSON(bytes []byte) error {
	var result struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Result  interface{} `json:"result"`
	}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	_ = e.Froze(result.Code, result.Message, result.Result)
	return nil
}

func (e *innerErrCode) Froze(code, message string, result interface{}) ErrorCode {
	// 默认 ErrInternalServerError
	e.serviceName = "undefined"
	e.httpStatusCode = http.StatusInternalServerError
	e.code = "0000001"
	e.message = message

	multiErrCode := code
	index := strings.Index(multiErrCode, ".")
	if index > 0 {
		e.serviceName = multiErrCode[:index]
		if index >= len(multiErrCode)-1 {
			return e.WithResult(code + ";" + message)
		}
		multiErrCode = multiErrCode[index+1:]
	}
	if len(multiErrCode) <= 3 {
		return e.WithResult(code + ";" + message)
	}
	httpStatusCode, err := strconv.Atoi(multiErrCode[:3])
	if err != nil {
		return e.WithResult(
			fmt.Sprintf("code:%s,message:%s;%e",
				code, message, err))
	}
	if httpStatusCode < 100 || httpStatusCode > 599 {
		return e.WithResult(code + ";" + message)
	}
	e.httpStatusCode = httpStatusCode
	e.code = multiErrCode[3:]
	e.result = result
	return e
}

// Froze defines ErrorCode
func Froze(code, message string) ErrorCode {
	return (&innerErrCode{}).Froze(code, message, nil)
}
