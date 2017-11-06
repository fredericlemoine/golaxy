package golaxy

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Response after the creation/deletion of a history
type historyResponse struct {
	Importable        bool         `json:"importable"`
	Create_time       string       `json:"create_time"`
	Contents_url      string       `json:"contente_url"`
	Id                string       `json:"id"`
	Size              int          `json:"size"`
	User_id           string       `json:"user_id"`
	Username_and_slug string       `json:"username_and_slug"`
	Annotation        string       `json:"annotation"`
	State_details     stateDetails `json:"state_details"`
	State             string       `json:"state"`
	Empty             bool         `json:"empty"`
	Update_time       string       `json:"update_time"`
	Tags              []string     `json:"tags"`
	Deleted           bool         `json:"deleted"`
	Genome_build      string       `json:"genome_build"`
	Slug              string       `json:"slug"`
	Name              string       `json:"name"`
	Url               string       `json:"url"`
	State_ids         stateIds     `json:"state_ids"`
	Published         bool         `json:"published"`
	Model_class       string       `json:"model_class"`
	Purged            bool         `json:"purged"`
	Err_msg           string       `json:"err_msg"`  // In case of error, this field is !=""
	Err_code          int          `json:"err_code"` // In case of error, this field is !=0
}

// Detail of the job state
type stateDetails struct {
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

// Id of jobs in different states
type stateIds struct {
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

// Request for fileupload
type fileUpload struct {
	File_type      string `json:"file_type"`
	Dbkey          string `json:"dbkey"`
	To_posix_lines bool   `json:"files0|to_posix_lines"`
	Space_to_tab   bool   `json:"files0|space_to_tab"`
	Filename       string `json:"files0|NAME"`
	Type           string `json:"files0|type"`
}

// Response after calling a tool
type toolResponse struct {
	Outputs              []toolOutput `json:"outputs"`
	Implicit_collections []string     `json:"implicit_collections"`
	Jobs                 []toolJob    `json:"jobs"`
	Output_collections   []string     `json:"output_collections"`
	Err_msg              string       `json:"err_msg"`  // In case of error, this field is !=""
	Err_code             int          `json:"err_code"` // In case of error, this field is !=0
}

type toolOutput struct {
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

type toolJob struct {
	Tool_id     string `json:"tool_id"`     // id of the tool
	Update_time string `json:"update_time"` // time stamp
	Exit_code   string `json:"exit_code"`
	State       string `json:"state"` //
	Create_time string `json:"create_time"`
	Model_class string `json:"model_class"`
	Id          string `json:"id"`
}

// Response when checking job status
type job struct {
	Tool_id      string               `json:"tool_id"`      // id of the tool
	Update_time  string               `json:"update_time"`  // timestamp
	Inputs       map[string]toolInput `json:"inputs"`       // input datasets
	Outputs      map[string]toolInput `json:"outputs"`      // output datasets
	Command_line string               `json:"command_line"` // full commandline
	Exit_code    int                  `json:"exit_code"`    // Tool exit code
	State        string               `json:"state"`        // Job state
	Create_time  string               `json:"create_time"`  // Job creation time
	Params       map[string]string    `json:"params"`       // ?
	Model_class  string               `json:"model_class"`  // Kind of object: "job"
	External_id  string               `json:"external_id"`  // ?
	Id           string               `json:"id"`           // Id if the job
	Err_msg      string               `json:"err_msg"`      // In case of error, this field is !=""
	Err_code     int                  `json:"err_code"`     // In case of error, this field is !=0
}

// Request to call a tool
type toolLaunch struct {
	History_id string                 `json:"history_id"` // Id of history
	Tool_id    string                 `json:"tool_id"`    // Id of the tool
	Inputs     map[string]interface{} `json:"inputs"`     // Inputs: key name of the input, value dataset id
}

type toolInput struct {
	Src  string `json:"src"`  // "hda"
	Id   string `json:"id"`   // dataset id
	UUid string `json:"uuid"` // ?
}

type Galaxy struct {
	url              string // url of the galaxy instance: http(s)://ip:port/
	apikey           string // api key
	trustcertificate bool   // if we should trust galaxy certificate
}

const (
	HISTORY   = "/api/histories"
	CHECK_JOB = "/api/jobs/"
	TOOLS     = "/api/tools"
)

// Initializes a new Galaxy with given:
// 	- url of the form http(s)://ip:port
// 	- and an api key
func NewGalaxy(url, key string, trustcertificate bool) *Galaxy {
	return &Galaxy{
		url,
		key,
		trustcertificate,
	}
}

// Creates an history with given name on the Galaxy instance
// and returns its id
func (g *Galaxy) CreateHistory(name string) (historyid string, err error) {
	var url string = g.url + HISTORY + "?key=" + g.apikey
	var req *http.Request
	var resp *http.Response
	var answer historyResponse
	var body []byte

	if req, err = http.NewRequest("POST", url, bytes.NewBuffer([]byte("{\"name\":\""+name+"\"}"))); err != nil {
		return
	}

	if resp, err = g.newClient().Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &answer); err != nil {
		return
	}
	if answer.Err_code != 0 {
		err = errors.New(answer.Err_msg)
		return
	}
	historyid = answer.Id
	return
}

// Uploads the given file to the galaxy instance in the history defined by its id
// and the given type (auto/txt/nhx/etc.)
//
// Returns the file id, the job id and a potential error
func (g *Galaxy) UploadFile(historyid string, path string, ftype string) (fileid, jobid string, err error) {
	var url string = g.url + TOOLS + "?key=" + g.apikey
	var file *os.File
	var body *bytes.Buffer
	var body2 []byte
	var writer *multipart.Writer
	var part io.Writer
	var fileinput *fileUpload
	var input []byte
	var postrequest *http.Request
	var postresponse *http.Response
	var answer toolResponse

	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)

	if part, err = writer.CreateFormFile("files_0|file_data", filepath.Base(path)); err != nil {
		return
	}

	if _, err = io.Copy(part, file); err != nil {
		return
	}

	if err = writer.WriteField("history_id", historyid); err != nil {
		return
	}
	if err = writer.WriteField("tool_id", "upload1"); err != nil {
		return
	}

	fileinput = &fileUpload{
		ftype,
		"?",
		false,
		false,
		filepath.Base(path),
		"upload_dataset",
	}
	if input, err = json.Marshal(fileinput); err != nil {
		return
	}
	if err = writer.WriteField("inputs", string(input)); err != nil {
		return
	}

	if err = writer.Close(); err != nil {
		return
	}

	if postrequest, err = http.NewRequest("POST", url, body); err != nil {
		return
	}
	postrequest.Header.Set("Content-Type", writer.FormDataContentType())

	if postresponse, err = g.newClient().Do(postrequest); err != nil {
		return
	}
	defer postresponse.Body.Close()

	if body2, err = ioutil.ReadAll(postresponse.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body2, &answer); err != nil {
		return
	}
	if answer.Err_code != 0 {
		err = errors.New(answer.Err_msg)
		return
	}

	if len(answer.Outputs) != 1 {
		err = errors.New("Error while uploading the file : Number of Outputs")
		return
	}

	fileid = answer.Outputs[0].Id
	if len(answer.Jobs) != 1 {
		err = errors.New("Error while uploading the file : Number of Jobs")
		return
	}
	jobid = answer.Jobs[0].Id

	return
}

// Launches a job at the given galaxy instance, with:
// 	- The tool given by its id (name)
// 	- Using the given history
// 	- Giving as input the files in the map : key: tool input name, value: dataset id
//
// Returns:
// 	- Tool outputs : map[out file name]=out file id
// 	- Jobs: array of job ids
func (g *Galaxy) LaunchTool(historyid string, toolid string, infiles map[string]string, inparams map[string]string) (outfiles map[string]string, jobids []string, err error) {
	var url string = g.url + TOOLS + "?key=" + g.apikey
	var launch *toolLaunch
	var input []byte
	var req *http.Request
	var resp *http.Response
	var body []byte
	var answer toolResponse

	launch = &toolLaunch{
		historyid,
		toolid,
		make(map[string]interface{}),
	}

	for k, v := range infiles {
		launch.Inputs[k] = toolInput{"hda", v, ""}
	}

	for k, v := range inparams {
		launch.Inputs[k] = v
	}

	if input, err = json.Marshal(launch); err != nil {
		return
	}
	if req, err = http.NewRequest("POST", url, bytes.NewBuffer(input)); err != nil {
		return
	}

	if resp, err = g.newClient().Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	if err = json.Unmarshal(body, &answer); err != nil {
		return
	}
	if answer.Err_msg != "" {
		err = errors.New(answer.Err_msg)
		return
	}

	outfiles = make(map[string]string)
	for _, to := range answer.Outputs {
		outfiles[to.Name] = to.Id
	}
	jobids = make([]string, 0, 10)
	for _, j := range answer.Jobs {
		jobids = append(jobids, j.Id)
	}

	return
}

// Queries the galaxy instance to check the job defined by its Id
// Returns:
// 	- job State
// 	- Output files: map : key: out filename value: out file id
func (g *Galaxy) CheckJob(jobid string) (jobstate string, outfiles map[string]string, err error) {
	var url string = g.url + CHECK_JOB + "/" + jobid + "?key=" + g.apikey
	var req *http.Request
	var response *http.Response
	var body []byte
	var answer job

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}
	if response, err = g.newClient().Do(req); err != nil {
		return
	}
	defer response.Body.Close()

	if body, err = ioutil.ReadAll(response.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &answer); err != nil {
		return
	}
	if answer.Err_code != 0 {
		err = errors.New(answer.Err_msg)
		return
	}

	outfiles = make(map[string]string)
	for k, v := range answer.Outputs {
		outfiles[k] = v.Id
	}
	jobstate = answer.State
	return
}

// Downloads a file defined by its id from the given history of the galaxy instance
// Returns:
// 	- The content of the file in []byte
// 	- A potential error
func (g *Galaxy) DownloadFile(historyid, fileid string) (content []byte, err error) {
	var url string = g.url + "/api/histories/" + historyid + "/contents/" + fileid + "/display" + "?key=" + g.apikey
	var req *http.Request
	var response *http.Response

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}

	if response, err = g.newClient().Do(req); err != nil {
		return
	}
	defer response.Body.Close()

	if content, err = ioutil.ReadAll(response.Body); err != nil {
		return
	}
	return
}

// Deletes and purges an history defined by its id
// Returns:
// 	- The state of the deletion ("ok")
// 	- A potential error
func (g *Galaxy) DeleteHistory(historyid string) (state string, err error) {
	var url string = g.url + HISTORY + "/" + historyid + "?key=" + g.apikey
	var req *http.Request
	var response *http.Response
	var body []byte
	var answer historyResponse

	req, _ = http.NewRequest("DELETE", url, bytes.NewBuffer([]byte("{\"purge\":\"true\"}")))

	if response, err = g.newClient().Do(req); err != nil {
		return
	}
	defer response.Body.Close()

	if body, err = ioutil.ReadAll(response.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &answer); err != nil {
		return
	}
	if answer.Err_code != 0 {
		err = errors.New(answer.Err_msg)
		return
	}

	state = answer.State
	return
}

func (g *Galaxy) newClient() *http.Client {
	config := &tls.Config{InsecureSkipVerify: g.trustcertificate} // this line here
	tr := &http.Transport{TLSClientConfig: config}
	return &http.Client{Transport: tr}
}

// This function returns ID of the tools corresponding to
// the name in argument
//
// It queries the galaxy entry point api/tools?q=<name>
func (g *Galaxy) SearchTool(name string) (answer []string, err error) {
	var url string = g.url + TOOLS + "?q=" + name
	var req *http.Request
	var resp *http.Response
	var body []byte

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}

	if resp, err = g.newClient().Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	err = json.Unmarshal(body, &answer)
	return
}
