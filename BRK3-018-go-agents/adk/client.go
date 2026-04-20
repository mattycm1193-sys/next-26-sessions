package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/genai"
)

type ADKClient struct {
	BaseURL string
	App     string
	User    string
	Session string
	c       http.Client
}

type AdkRunReq struct {
	AppName string        `json:"appName"`
	User    string        `json:"userId"`
	Session string        `json:"sessionId"`
	Message genai.Content `json:"newMessage"`
}
type AdkRunResp struct {
	ID      string        `json:"id"`
	Content genai.Content `json:"content"`
}

type AdkSessionResp struct {
	ID      string
	AppName string
	UserId  string
}

func (c *ADKClient) NewSession() error {
	url, err := url.JoinPath(c.BaseURL, fmt.Sprintf("/apps/%s/users/%s/sessions", c.App, c.User))
	if err != nil {
		return err
	}
	res, err := http.Post(url, "text/plain", strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("failed to create session %w", err)
	}
	var resp AdkSessionResp
	jd := json.NewDecoder(res.Body)
	defer res.Body.Close()
	err = jd.Decode(&resp)
	if err != nil {
		return err
	}
	c.Session = resp.ID
	return nil
}
func (c *ADKClient) Run(content genai.Content) ([]AdkRunResp, error) {
	req := &AdkRunReq{
		AppName: c.App,
		User:    c.User,
		Session: c.Session,
		Message: content,
	}
	reqbytes, err := json.Marshal(req)

	res, err := http.Post(c.BaseURL+"/run",
		"application/json", bytes.NewReader(reqbytes))
	if err != nil {
		return []AdkRunResp{}, err
	}
	defer res.Body.Close()
	response, _ := io.ReadAll(res.Body)
	respEvents := make([]AdkRunResp, 0)
	err = json.Unmarshal(response, &respEvents)
	return respEvents, err
}
