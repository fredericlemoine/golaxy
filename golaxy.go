package golaxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type CreateHistoryResponse struct {
	Importable        bool         `json:"importable"`
	Create_time       string       `json:"create_time"`
	Contents_url      string       `json:"contente_url"`
	Id                string       `json:"id"`
	Size              int          `json:"size"`
	User_id           string       `json:"user_id"`
	Username_and_slug string       `json:"username_and_slug"`
	Annotation        string       `json:"annotation"`
	State_details     StateDetails `json:"state_details"`
	State             string       `json:"state"`
	empty             bool         `json:"empty"`
	Update_time       string       `json:"update_time"`
	Tags              []string     `json:"tags"`
	Deleted           bool         `json:"deleted"`
	Genome_build      string       `json:"genome_build"`
	Slug              string       `json:"slug"`
	Name              string       `json:"name"`
	Url               string       `json:"url"`
	State_ids         StateIds     `json:"state_ids"`
	Published         bool         `json:"published"`
	Model_class       string       `json:"model_class"`
	Purged            bool         `json:"purged"`
}

type StateDetails struct {
	Paused           int `json:"paused"`
	Ok               int `json:"ok"`
	Failed_metadata  int `json:"failed_metadata"`
	Upload           int `json:"upload"`
	Discarded        int `json:"discarded"`
	Running          int `json:"running"`
	Setting_metadata int `json:"setting_metadata"`
	Error            int `json:"error"`
	New              int `json:"new"`
	Queued           int `json:"queued"`
	Empty            int `json:"empty"`
}

type StateIds struct {
	Paused           []string `json:"paused"`
	Ok               []string `json:"ok"`
	Failed_metadata  []string `json:"failed_metadata"`
	Upload           []string `json:"upload"`
	Discarded        []string `json:"discarded"`
	Running          []string `json:"running"`
	Setting_metadata []string `json:"setting_metadata"`
	Error            []string `json:"error"`
	New              []string `json:"new"`
	Queued           []string `json:"queued"`
	Empty            []string `json:"empty"`
}

type FileUpload struct {
	File_type      string `json:"file_type"`
	Dbkey          string `json:"dbkey"`
	To_posix_lines bool   `json:"files0|to_posix_lines"`
	Space_to_tab   bool   `json:"files0|space_to_tab"`
	Filename       string `json:"files0|NAME"`
	Type           string `json:"files0|type"`
}

type ToolResponse struct {
	Outputs              []ToolOutput `json:"outputs"`
	Implicit_collections []string     `json:"implicit_collections"`
	Jobs                 []ToolJob    `json:"jobs"`
	Output_collections   []string     `json:"output_collections"`
}

type ToolOutput struct {
	Misc_blurb           string   `json:"misc_blurb"`
	Peek                 string   `json:"peek"`
	Update_time          string   `json:"update_time"`
	Data_type            string   `json:"data_type"`
	Tags                 []string `json:"tags"`
	Deleted              bool     `json:"deleted"`
	History_id           string   `json:"history_id"`
	Visible              bool     `json:"visible"`
	Genome_build         string   `json:"genome_build"`
	Create_time          string   `json:"create_time"`
	Hid                  int      `json:"hid"`
	File_size            int      `json:"file_size"`
	Metadata_data_lines  string   `json:"metadata_data_lines"`
	File_ext             string   `json:"file_ext"`
	Id                   string   `json:"id"`
	Misc_info            string   `json:"misc_info"`
	Hda_ldda             string   `json:"hda"`
	History_content_type string   `json:"dataset"`
	Name                 string   `json:"name"`
	Uuid                 string   `json:"uuid"`
	State                string   `json:"state"`
	Model_class          string   `json:"model_class"`
	Metadata_dbkey       string   `json:"metadata_dbkey"`
	Output_Name          string   `json:"output_name"`
	Purged               bool     `json:"purged"`
}

type ToolJob struct {
	Tool_id     string `json:"tool_id"`
	Update_time string `json:"update_time"`
	Exit_code   string `json:"exit_code"`
	State       string `json:"state"`
	Create_time string `json:"create_time"`
	Model_class string `json:"model_class"`
	Id          string `json:"id"`
}

type Job struct {
	Tool_id      string `json:"tool_id"`
	Update_time  string `json:"update_time"`
	Inputs       map[string]ToolInput
	Outputs      map[string]ToolInput
	Command_line string `json:"command_line"`
	Exit_code    int    `json:"exit_code"`
	State        string `json:"state"`
	Create_time  string `json:"create_time"`
	Params       map[string]string
	Model_class  string `json:"model_class"`
	External_id  string `json:"external_id"`
	Id           string `json:"id"`
}

type ToolLaunch struct {
	History_id string               `json:"history_id"`
	Tool_id    string               `json:"tool_id"`
	Infiles    map[string]ToolInput `json:"inputs"`
}

type ToolInput struct {
	Src  string `json:"src"`
	Id   string `json:"id"`
	UUid string `json:"uuid"`
}

type Galaxy struct {
	url    string
	apikey string
}

const (
	CREATE_HISTORY = "/api/histories"
	CHECK_JOB      = "/api/jobs/"
	TOOLS          = "/api/tools"
)

func NewGalaxy(url, key string) *Galaxy {
	return &Galaxy{
		url,
		key,
	}
}

func (g *Galaxy) CreateHistory(name string) (string, error) {
	url := g.url + CREATE_HISTORY + "?key=" + g.apikey
	fmt.Println(url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("{\"name\":\""+name+"\"}")))
	client := &http.Client{}
	resp, err2 := client.Do(req)
	if err2 != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err3 := ioutil.ReadAll(resp.Body)
	if err3 != nil {
		return "", err3
	}
	var answer CreateHistoryResponse
	err4 := json.Unmarshal(body, &answer)
	if err4 != nil {
		return "", err4
	}

	return answer.Id, nil
}

/* Returns File id, job id and error*/
func (g *Galaxy) UploadFile(historyid string, path string) (string, string, error) {
	url := g.url + TOOLS + "?key=" + g.apikey
	fmt.Println(url)

	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err2 := writer.CreateFormFile("files_0|file_data", filepath.Base(path))
	if err2 != nil {
		return "", "", err2
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", "", err
	}

	if err = writer.WriteField("history_id", historyid); err != nil {
		return "", "", err
	}
	if err = writer.WriteField("tool_id", "upload1"); err != nil {
		return "", "", err
	}

	fileinput := &FileUpload{
		"auto",
		"?",
		false,
		false,
		filepath.Base(path),
		"upload_dataset",
	}
	input, err4 := json.Marshal(fileinput)
	if err4 != nil {
		return "", "", err
	}
	if err = writer.WriteField("inputs", string(input)); err != nil {
		return "", "", err
	}

	if err = writer.Close(); err != nil {
		return "", "", err
	}

	postrequest, err5 := http.NewRequest("POST", url, body)
	if err5 != nil {
		return "", "", err5
	}

	postrequest.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	postresponse, err6 := client.Do(postrequest)
	if err6 != nil {
		return "", "", err6
	}

	defer postresponse.Body.Close()

	body2, err7 := ioutil.ReadAll(postresponse.Body)
	if err7 != nil {
		return "", "", err7
	}

	var answer ToolResponse
	err8 := json.Unmarshal(body2, &answer)
	if err8 != nil {
		return "", "", err8
	}

	if len(answer.Outputs) != 1 {
		return "", "", errors.New("Error while uploading the file : Number of Outputs")
	}
	fileid := answer.Outputs[0].Id
	if len(answer.Jobs) != 1 {
		return "", "", errors.New("Error while uploading the file : Number of Jobs")
	}
	jobid := answer.Jobs[0].Id

	return fileid, jobid, nil
}

/* Returns job id and error
infiles: key: input name, value: file id
Output: map[out file name]=out file id
Output: array of out job ids
*/
func (g *Galaxy) LaunchTool(historyid string, toolid string, infiles map[string]string) (map[string]string, []string, error) {
	url := g.url + TOOLS + "?key=" + g.apikey
	fmt.Println(url)

	launch := &ToolLaunch{
		historyid,
		toolid,
		make(map[string]ToolInput),
	}
	for k, v := range infiles {
		launch.Infiles[k] = ToolInput{"hda", v, ""}
	}
	input, err := json.Marshal(launch)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(input))
	client := &http.Client{}
	resp, err2 := client.Do(req)
	if err2 != nil {
		return nil, nil, err2
	}
	defer resp.Body.Close()

	body, err3 := ioutil.ReadAll(resp.Body)
	if err3 != nil {
		return nil, nil, err3
	}
	var answer ToolResponse
	err4 := json.Unmarshal(body, &answer)
	if err4 != nil {
		return nil, nil, err4
	}

	outfiles := make(map[string]string)
	for _, to := range answer.Outputs {
		outfiles[to.Name] = to.Id
	}
	outjobs := make([]string, 0, 10)
	for _, j := range answer.Jobs {
		outjobs = append(outjobs, j.Id)
	}

	return outfiles, outjobs, nil
}

/*
  int job status
  map : key: out filename value: out file id
*/
func (g *Galaxy) CheckJob(jobid string) (string, map[string]string, error) {
	url := g.url + CHECK_JOB + "/" + jobid + "?key=" + g.apikey
	fmt.Println(url)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	response, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer response.Body.Close()

	body, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		return "", nil, err2
	}

	var answer Job
	err3 := json.Unmarshal(body, &answer)
	if err3 != nil {
		return "", nil, err3
	}

	outfiles := make(map[string]string)
	for k, v := range answer.Outputs {
		outfiles[k] = v.Id
	}
	return answer.State, outfiles, nil
}
func (g *Galaxy) DownloadFile(historyid, fileid string) ([]byte, error) {
	url := g.url + "/api/histories/" + historyid + "/contents/" + fileid + "/display" + "?key=" + g.apikey
	fmt.Println(url)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		return nil, err2
	}
	return body, nil
}
