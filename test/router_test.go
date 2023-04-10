package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

const (
	httpURL     = "http://localhost:9090"
	contentType = "application/x-www-form-urlencoded"
)

// 注册用户
func TestRegister(t *testing.T) {
	// 创建一个 http.Client
	client := &http.Client{}

	// 准备要发送的表单数据
	data := url.Values{}
	data.Set("phone_number", "55555")
	data.Set("nickname", "test")
	data.Set("password", "123")

	// 创建一个 POST 请求
	req, err := http.NewRequest("POST", httpURL+"/register", strings.NewReader(data.Encode()))
	if err != nil {
		// 处理错误
		return
	}
	req.Header.Set("Content-Type", contentType)

	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		// 处理错误
		return
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// 处理读取错误
		panic(err)
	}

	var respData struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Token string `json:"token"`
			Id    string `json:"id"`
		} `json:"data"`
	}
	err = json.Unmarshal(responseBody, &respData)
	if err != nil {
		panic(err)
	}
	t.Log(respData)
}

// 登录
func TestLogin(t *testing.T) {
	// 创建一个 http.Client
	client := &http.Client{}

	// 准备要发送的表单数据
	data := url.Values{}
	data.Set("phone_number", "123456789")
	data.Set("password", "123")

	// 创建一个 POST 请求
	req, err := http.NewRequest("POST", httpURL+"/login", strings.NewReader(data.Encode()))
	if err != nil {
		// 处理错误
		return
	}
	req.Header.Set("Content-Type", contentType)

	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		// 处理错误
		return
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// 处理读取错误
		panic(err)
	}
	var respData struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Token  string `json:"token"`
			UserId string `json:"user_id"`
		} `json:"data"`
	}
	err = json.Unmarshal(responseBody, &respData)
	if err != nil {
		panic(err)
	}
	t.Log(respData)
}

func TestAddFriend(t *testing.T) {
	// 创建一个 http.Client
	client := &http.Client{}

	// 准备要发送的表单数据
	data := url.Values{}
	data.Set("friend_id", "6")

	// 创建一个 POST 请求
	req, err := http.NewRequest("POST", httpURL+"/friend/add", strings.NewReader(data.Encode()))
	if err != nil {
		// 处理错误
		return
	}
	req.Header.Set("Content-Type", contentType)
	// 设置 token
	//req.Header.Set("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MywiZXhwIjoxNjgyNTcyNTI5fQ.uHn7XSVb2T4cBUC6gBE8-iQbnI_pqB0bWFPAkOtQmPk")
	req.Header.Set("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6NCwiZXhwIjoxNjgyNTczODcxfQ.Ksw5J8vfVkUPPmM-2EJeiMFr9THqKhvlRKIR_W4H3SE")

	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		// 处理错误
		return
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// 处理读取错误
		panic(err)
	}
	fmt.Println(string(responseBody))
	var respData struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	err = json.Unmarshal(responseBody, &respData)
	if err != nil {
		panic(err)
	}
	t.Log(respData)
}

func TestCreateGroup(t *testing.T) {
	// 创建一个 http.Client
	client := &http.Client{}

	// 准备要发送的表单数据
	data := url.Values{}
	data.Set("name", "6")
	ids := []string{"1", "2", "3", "4", "5", "6", "7"}
	for _, id := range ids {
		data.Add("ids", id)
	}

	// 创建一个 POST 请求
	req, err := http.NewRequest("POST", httpURL+"/group/create", strings.NewReader(data.Encode()))
	if err != nil {
		// 处理错误
		return
	}
	req.Header.Set("Content-Type", contentType)
	// 设置 token
	req.Header.Set("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6NCwiZXhwIjoxNjgyNTczODcxfQ.Ksw5J8vfVkUPPmM-2EJeiMFr9THqKhvlRKIR_W4H3SE")

	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		// 处理错误
		return
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// 处理读取错误
		panic(err)
	}
	fmt.Println(string(responseBody))
	var respData struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Id string `json:"id"`
		} `json:"data"`
	}
	err = json.Unmarshal(responseBody, &respData)
	if err != nil {
		panic(err)
	}
	t.Log(respData)
}

func TestGroupUserList(t *testing.T) {
	// 创建一个 http.Client
	client := &http.Client{}

	// 准备要发送的表单数据
	data := url.Values{}
	data.Set("group_id", "1")

	// 创建一个 POST 请求
	req, err := http.NewRequest("GET", httpURL+"/group_user/list?"+data.Encode(), nil)
	if err != nil {
		// 处理错误
		return
	}
	// 设置 token
	req.Header.Set("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6NCwiZXhwIjoxNjgyNTczODcxfQ.Ksw5J8vfVkUPPmM-2EJeiMFr9THqKhvlRKIR_W4H3SE")

	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		// 处理错误
		return
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// 处理读取错误
		panic(err)
	}
	fmt.Println(string(responseBody))
	var respData struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Ids []string `json:"ids"`
		} `json:"data"`
	}
	err = json.Unmarshal(responseBody, &respData)
	if err != nil {
		panic(err)
	}
	t.Log(respData)
}
