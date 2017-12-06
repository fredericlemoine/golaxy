package golaxy

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Version of galaxy, returned by /api/version
type galaxyVersion struct {
	Extra         interface{} `json:"extra"`
	Version_major string      `json:"version_major"`
	Err_msg       string      `json:"err_msg"`  // In case of error, this field is !=""
	Err_code      int         `json:"err_code"` // In case of error, this field is !=0
}

// Response after the creation/deletion of a history
type HistoryFullInfo struct {
	Importable        bool                `json:"importable"`
	Create_time       string              `json:"create_time"`
	Contents_url      string              `json:"contente_url"`
	Id                string              `json:"id"`
	Size              int                 `json:"size"`
	User_id           string              `json:"user_id"`
	Username_and_slug string              `json:"username_and_slug"`
	Annotation        string              `json:"annotation"`
	State_details     HistoryStateDetails `json:"state_details"`
	State             string              `json:"state"`
	Empty             bool                `json:"empty"`
	Update_time       string              `json:"update_time"`
	Tags              []string            `json:"tags"`
	Deleted           bool                `json:"deleted"`
	Genome_build      string              `json:"genome_build"`
	Slug              string              `json:"slug"`
	Name              string              `json:"name"`
	Url               string              `json:"url"`
	State_ids         HistoryStateIds     `json:"state_ids"`
	Published         bool                `json:"published"`
	Model_class       string              `json:"model_class"`
	Purged            bool                `json:"purged"`
	Err_msg           string              `json:"err_msg"`  // In case of error, this field is !=""
	Err_code          int                 `json:"err_code"` // In case of error, this field is !=0
}

// Detail of the job state
type HistoryStateDetails struct {
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
type HistoryStateIds struct {
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

type HistoryShortInfo struct {
	Annotation  string   `json:"annotation"`
	Deleted     bool     `json:"deleted"`
	Id          string   `json:"id"`
	Model_class string   `json:"model_class"`
	Name        string   `json:"name"`
	Published   bool     `json:"published"`
	Purged      bool     `json:"purged"`
	Tags        []string `json:"tags"`
	Url         string   `json:"url"`
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
	Traceback    string               `json:"traceback"`
	Err_msg      string               `json:"err_msg"`  // In case of error, this field is !=""
	Err_code     int                  `json:"err_code"` // In case of error, this field is !=0
}

// Request to call a tool
type ToolLaunch struct {
	History_id string                 `json:"history_id"` // Id of history
	Tool_id    string                 `json:"tool_id"`    // Id of the tool
	Inputs     map[string]interface{} `json:"inputs"`     // Inputs: key name of the input, value dataset id
}

type toolInput struct {
	Src  string `json:"src"`  // "hda"
	Id   string `json:"id"`   // dataset id
	UUid string `json:"uuid"` // ?
}

// Informations about a specific tool
type ToolInfo struct {
	Description          string             `json:"description"` // Description of the tool
	Edam_operations      []string           `json:"edam_operations"`
	Edam_topics          []string           `json:"edam_topics"`
	Form_style           string             `json:"form_style"`
	Id                   string             `json:"id"`
	Labels               []string           `json:"labels"`
	Model_class          string             `json:"model_class"`
	Name                 string             `json:"name"`
	Panel_Section_Id     string             `json:"panel_section_id"`
	Panel_Section_Name   string             `json:"panel_section_name"`
	Tool_Shed_Repository ToolShedRepository `json:"tool_shed_repository"`
	Version              string             `json:"version"`
	Traceboack           string             `json:"traceback"` // Set only if the server returns an error
	Err_Msg              string             `json:"err_msg"`   // Err_Msg =="" if no error
	Err_Code             string             `json:"err_code"`  // Err_Code=="" if no error
}

// Informations about a toolshed
type ToolShedRepository struct {
	Changeset_Revision string `json:"changeset_revision"`
	Name               string `json:"name"`
	Owner              string `json:"owner"`
	Tool_shed          string `json:"tool_shed"`
}

type WorkflowInput struct {
	Label string `json:label`
	Uuid  string `json:uuid`
	Value string `json:uuid`
}

type WorkflowInputStep struct {
	Source_Step int    `json:"source_step"` // ID of the previous step serving as input here
	Step_Output string `json:"step_output"` // Name of the output  of the previous step serving as input here
}

type WorkflowStep struct {
	Annotation   string                       `json:"annotation"`
	Id           int                          `json:"id"`
	Input_Steps  map[string]WorkflowInputStep `json:"input_steps"`
	Tool_Id      string                       `json:"tool_id"`
	Tool_Inputs  map[string]string            `json:"tool_inputs"`
	Tool_Version string                       `json:"tool_version"`
	Type         string                       `json:"type"`
}

// Informations about a workflow
type WorkflowInfo struct {
	Annotation           string                   `json:"annotation"`
	Deleted              bool                     `json:"deleted"`
	Id                   string                   `json:"id"`
	Inputs               map[string]WorkflowInput `json:"inputs"`
	Latest_Workflow_UUID string                   `json:"latest_workflow_uuid"`
	Model_Class          string                   `json:"model_class"`
	Name                 string                   `json:"name"`
	Owner                string                   `json:"owner"`
	Published            bool                     `json:"published"`
	Steps                map[string]WorkflowStep  `json:"steps"`
	Tags                 []string                 `json:"tags"`
	Url                  string                   `json:"url"`
	Traceboack           string                   `json:"traceback"` // Set only if the server returns an error
	Err_Msg              string                   `json:"err_msg"`   // Err_Msg =="" if no error
	Err_Code             int                      `json:"err_code"`  // Err_Code==0 if no error
}

type WorkflowLaunch struct {
	History_id  string                    `json:"history_id"`  // Id of history
	Workflow_id string                    `json:"workflow_id"` // Id of the tool
	Inputs      map[string]toolInput      `json:"inputs"`      // Inputs: key name of the input, value dataset id
	Parameters  map[int]map[string]string `json:"parameters"`  // Parameters to the workflow : key: Step id, value: map of key:value parameters
}

// When a workflow is launched, it is returned by the server
type WorkflowInvocation struct {
	History     string                   `json:"history"`
	History_Id  string                   `json:"history_id"`
	Id          string                   `json:"id"`
	Inputs      map[string]toolInput     `json:"inputs"`
	Model_Class string                   `json:"model_class"`
	Outputs     []string                 `json:"outputs"`
	State       string                   `json:"state"`
	Steps       []WorkflowInvocationStep `json:"steps"`
	Update_Time string                   `json:"update_time"`
	Uuid        string                   `json:"uuid"`
	Workflow_Id string                   `json:"workflow_id"`
	Traceboack  string                   `json:"traceback"` // Set only if the server returns an error
	Err_Msg     string                   `json:"err_msg"`   // Err_Msg =="" if no error
	Err_Code    int                      `json:"err_code"`  // Err_Code=="" if no error
}

// One of the steps given after invocation of the workflow
type WorkflowInvocationStep struct {
	Action              string `json:"action"`
	Id                  string `json:"id"`
	Job_Id              string `json:"job_id"`
	Model_Class         string `json:"model_class"`
	Order_Index         int    `json:"order_index"`
	State               string `json:"state"`
	Update_Time         string `json:"update_time"`
	Workflow_Step_Id    string `json:"workflow_step_id"`
	Workflow_Step_Label string `json:"workflow_step_label"`
	Workflow_Step_Uuid  string `json:"workflow_step_uuid"`
}

// Error returned by galaxy
type genericError struct {
	Traceboack string `json:"traceback"` // Set only if the server returns an error
	Err_Msg    string `json:"err_msg"`   // Err_Msg =="" if no error
	Err_Code   int    `json:"err_code"`  // Err_Code==0 if no error
}

// Information about status of a workflow run
//	- General status
//	- Status of each step
//	- Output file ids of each steps
type WorkflowStatus struct {
	wfStatus   string                    // Workflow global status
	stepStatus map[int]string            // All step status per steprank
	outfiles   map[int]map[string]string // All output files (map[name]id) per steprank
}

type Galaxy struct {
	url              string // url of the galaxy instance: http(s)://ip:port/
	apikey           string // api key
	trustcertificate bool   // if we should trust galaxy certificate
	requestattempts  int    // Number of times requests (get post or delete) must be tried if an error occurs (timeout for example). Default 1
}

const (
	HISTORY   = "/api/histories"
	CHECK_JOB = "/api/jobs/"
	TOOLS     = "/api/tools"
	WORKFLOWS = "/api/workflows"
	VERSION   = "/api/version"
)

// Initializes a new Galaxy with given:
// 	- url of the form http(s)://ip:port
// 	- and an api key
func NewGalaxy(url, key string, trustcertificate bool) *Galaxy {
	return &Galaxy{
		url,
		key,
		trustcertificate,
		1,
	}
}

// Sets the number of times golaxy will try to execute a request on
// the galaxy server if an error occurs (like a timeout).
//
// Default: 1
//
// It only applies on get, post and delete http request errors (timeout,
// etc.), and not Galaxy server errors.
//
// If given attempts is <=0, will not change anything.
func (g *Galaxy) SetNbRequestAttempts(attempts int) {
	if attempts > 0 {
		g.requestattempts = attempts
	}
}

// Returns the Version of the Galaxy Server
func (g *Galaxy) Version() (version string, err error) {
	var url string = g.url + VERSION
	var answer galaxyVersion

	if err = g.galaxyGetRequestJSON(url, &answer); err != nil {
		return
	}

	if answer.Err_code != 0 || answer.Err_msg != "" {
		err = errors.New(answer.Err_msg)
		return
	}

	version = answer.Version_major
	return
}

// Creates an history with given name on the Galaxy instance
// and returns its id
func (g *Galaxy) CreateHistory(name string) (history HistoryFullInfo, err error) {
	var url string = g.url + HISTORY + "?key=" + g.apikey

	if err = g.galaxyPostRequestJSON(url, []byte("{\"name\":\""+name+"\"}"), &history); err != nil {
		return
	}

	if history.Err_code != 0 || history.Err_msg != "" {
		err = errors.New(history.Err_msg)
		return
	}
	return
}

func (g *Galaxy) ListHistories() (histories []HistoryShortInfo, err error) {
	var url string = g.url + HISTORY + "?key=" + g.apikey
	var body []byte
	var galaxyErr genericError

	if body, err = g.galaxyGetRequestBytes(url); err != nil {
		return
	}

	// If we cannot unmarshall the []HistoryShortInfo
	// The we try to unmarshall it as a galaxyError
	if err = json.Unmarshal(body, &histories); err != nil {
		if err = json.Unmarshal(body, &galaxyErr); err != nil {
			return
		}
		if galaxyErr.Err_Code != 0 || galaxyErr.Err_Msg != "" {
			err = errors.New(galaxyErr.Err_Msg)
		} else {
			err = errors.New("Error while listing histories")
		}
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
	var answer HistoryFullInfo

	if err = g.galaxyDeleteRequestJSON(url, []byte("{\"purge\":\"true\"}"), &answer); err != nil {
		return
	}

	if answer.Err_code != 0 || answer.Err_msg != "" {
		err = errors.New(answer.Err_msg)
		return
	}

	state = answer.State
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

	if historyid == "" {
		err = errors.New("UploadFile input history id is not valid")
		return
	}

	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)

	if part, err = writer.CreateFormFile("files_0|file_data", filepath.Base(path)); err != nil {
		err = errors.New("Error while creating upload file form: " + err.Error())
		return
	}

	if _, err = io.Copy(part, file); err != nil {
		err = errors.New("Error while copying file content to form: " + err.Error())
		return
	}

	if err = writer.WriteField("history_id", historyid); err != nil {
		err = errors.New("Error while writing history id to form: " + err.Error())
		return
	}
	if err = writer.WriteField("tool_id", "upload1"); err != nil {
		err = errors.New("Error while writing tool id to form: " + err.Error())
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
		err = errors.New("Error while marshaling fileinput: " + err.Error())
		return
	}

	if err = writer.WriteField("inputs", string(input)); err != nil {
		err = errors.New("Error while writing file inputs to form: " + err.Error())
		return
	}

	if err = writer.Close(); err != nil {
		err = errors.New("Error while closing form writer: " + err.Error())
		return
	}

	if postrequest, err = http.NewRequest("POST", url, body); err != nil {
		err = errors.New("Error while creating new POST request: " + err.Error())
		return
	}
	postrequest.Header.Set("Content-Type", writer.FormDataContentType())

	if postresponse, err = g.newClient().Do(postrequest); err != nil {
		err = errors.New("Error while POSTing form: " + err.Error())
		return
	}
	defer postresponse.Body.Close()

	if body2, err = ioutil.ReadAll(postresponse.Body); err != nil {
		err = errors.New("Error while reading server respone: " + err.Error())
		return
	}

	if err = json.Unmarshal(body2, &answer); err != nil {
		err = errors.New("Error while unmarsheling server respone: " + err.Error())
		return
	}
	if answer.Err_msg != "" {
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

// Downloads a file defined by its id from the given history of the galaxy instance
// Returns:
// 	- The content of the file in []byte
// 	- A potential error
func (g *Galaxy) DownloadFile(historyid, fileid string) (content []byte, err error) {
	var url string = g.url + "/api/histories/" + historyid + "/contents/" + fileid + "/display" + "?key=" + g.apikey
	content, err = g.galaxyGetRequestBytes(url)
	return
}

// Initializes a ToolLaunch that will be used to start a new job with the given tool
// on the given history
func (g *Galaxy) NewToolLauncher(historyid string, toolid string) (tl *ToolLaunch) {
	tl = &ToolLaunch{
		historyid,
		toolid,
		make(map[string]interface{}),
	}
	return
}

// Add new input file to the Tool Launcher
//
//	- inputIndex : index of this input in the workflow (see WorkflowInfo / GetWorkflowById)
//	- fielId: id of input file
//	- fielScr : one of  ["ldda", "ld", "hda", "hdca"]
func (tl *ToolLaunch) AddFileInput(paramname string, fileid string, filescr string) {
	tl.Inputs[paramname] = toolInput{filescr, fileid, ""}
}

// Add new parameter to Tool launcher
//	- paramname: name of the tool parameter
//	- paramvalue: value of the given parameter
func (tl *ToolLaunch) AddParameter(paramname, paramvalue string) {
	tl.Inputs[paramname] = paramvalue
}

// Launches a job at the given galaxy instance, with:
// 	- The tool given by its id (name)
// 	- Using the given history
// 	- Giving as input the files in the map : key: tool input name, value: dataset id
//
// Returns:
// 	- Tool outputs : map[out file name]=out file id
// 	- Jobs: array of job ids
func (g *Galaxy) LaunchTool(tl *ToolLaunch) (outfiles map[string]string, jobids []string, err error) {
	var url string = g.url + TOOLS + "?key=" + g.apikey
	var input []byte
	var answer toolResponse

	if input, err = json.Marshal(tl); err != nil {
		return
	}

	if err = g.galaxyPostRequestJSON(url, input, &answer); err != nil {
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
	var answer job

	if err = g.galaxyGetRequestJSON(url, &answer); err != nil {
		return
	}

	if answer.Err_code != 0 || answer.Err_msg != "" {
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

func (g *Galaxy) newClient() *http.Client {
	config := &tls.Config{InsecureSkipVerify: g.trustcertificate}
	tr := &http.Transport{
		TLSClientConfig:     config,
		Dial:                (&net.Dialer{Timeout: 40 * time.Second}).Dial,
		TLSHandshakeTimeout: 40 * time.Second,
	}
	return &http.Client{Transport: tr, Timeout: 60 * time.Second}
}

// This function returns ID of the tools corresponding to
// the name/ID in argument.
//
// It first queryies the
//	api/tools/<name>
// entry point. If The returned tool has its ID == given name
// then returns it.
// Otherwise, it queries the galaxy entry point
//	api/tools?q=<name>
func (g *Galaxy) SearchToolID(name string) (toolIds []string, err error) {
	var info ToolInfo
	// We first try to see if the tool with an ID
	// Corresponding to the given name exists
	info, err = g.GetToolById(name)
	if err == nil && info.Err_Code == "" && info.Id == name {
		toolIds = []string{info.Id}
	} else {
		// Otherwize we use the search entry point
		toolIds, err = g.searchToolIDsByName(name)
	}
	return
}

func (g *Galaxy) GetToolById(id string) (tool ToolInfo, err error) {
	var url string = g.url + TOOLS + "/" + id
	err = g.galaxyGetRequestJSON(url, &tool)
	return
}

func (g *Galaxy) searchToolIDsByName(name string) (ids []string, err error) {
	var url string = g.url + TOOLS + "?q=" + name
	err = g.galaxyGetRequestJSON(url, &ids)
	return
}

// Search a workflow on the glaaxy server.
//
// It will first search for a workflow with ID exactly equal to the given name.
// If it does not exist, then will search a workflow having a lower(name) matching
// containing the given lower(name).
func (g *Galaxy) SearchWorkflowIDs(name string, published bool) (ids []string, err error) {
	var wf WorkflowInfo
	if wf, err = g.GetWorkflowByID(name, published); err != nil {
		ids, err = g.SearchWorkflowIDsByName(name, published)
	} else {
		ids = []string{wf.Id}
	}
	return
}

// Call galaxy api to look for a workflow with the given ID
//
// If published: then search the workflow in the published+imported workflows
func (g *Galaxy) GetWorkflowByID(inputid string, published bool) (wf WorkflowInfo, err error) {
	var url string = g.url + WORKFLOWS + "/" + inputid + "?key=" + g.apikey
	if published {
		url += "&show_published=true"
	}

	if err = g.galaxyGetRequestJSON(url, &wf); err != nil {
		return
	}

	if wf.Err_Code != 0 || wf.Err_Msg != "" {
		err = errors.New(wf.Err_Msg)
	}
	return
}

// Searches workflow ids by name.
//
// If published: then search the workflow in the published+imported workflows
func (g *Galaxy) SearchWorkflowIDsByName(name string, published bool) (ids []string, err error) {
	var wfs []WorkflowInfo
	var r *regexp.Regexp

	ids = make([]string, 0)

	if wfs, err = g.ListWorkflows(published); err != nil {
		return
	}

	if r, err = regexp.Compile(".*" + strings.ToLower(name) + ".*"); err != nil {
		return
	}

	for _, wf := range wfs {
		if ok := r.MatchString(strings.ToLower(wf.Name)); ok {
			ids = append(ids, wf.Id)
		}
	}
	return
}

// Lists all the workflows imported in the user's account
//
// If published: then lists published+imported workflows
func (g *Galaxy) ListWorkflows(published bool) (workflows []WorkflowInfo, err error) {
	var url string = g.url + WORKFLOWS + "?key=" + g.apikey
	var body []byte
	var galaxyErr genericError

	if published {
		url += "&show_published=true"
	}

	if body, err = g.galaxyGetRequestBytes(url); err != nil {
		return
	}

	// If we cannot unmarshall the []WorkflowInfo
	// The we try to unmarshall it as a galaxyError
	if err = json.Unmarshal(body, &workflows); err != nil {
		if err = json.Unmarshal(body, &galaxyErr); err != nil {
			return
		}
		if galaxyErr.Err_Code != 0 || galaxyErr.Err_Msg != "" {
			err = errors.New(galaxyErr.Err_Msg)
		} else {
			err = errors.New("Error while listing workflows")
		}
		return
	}
	return
}

// Lists all the workflows imported in the user's account
func (g *Galaxy) ImportSharedWorkflow(sharedworkflowid string) (workflow WorkflowInfo, err error) {
	var url string = g.url + WORKFLOWS + "?key=" + g.apikey

	err = g.galaxyPostRequestJSON(url, []byte("{\"shared_workflow_id\":\""+sharedworkflowid+"\"}"), &workflow)

	if err == nil && workflow.Err_Msg != "" || workflow.Err_Code != 0 {
		err = errors.New(workflow.Err_Msg)
	}

	return
}

// Deletes a Workflow defined by its id
// 	- The state of the deletion ("Workflow '<name>' successfully deleted" for example)
// 	- A potential error if the workflow cannot be deleted (server response does not contain "successfully deleted")
func (g *Galaxy) DeleteWorkflow(workflowid string) (state string, err error) {
	var url string = g.url + WORKFLOWS + "/" + workflowid + "?key=" + g.apikey
	var answer []byte

	if answer, err = g.galaxyDeleteRequestBytes(url, []byte{}); err != nil {
		return
	}

	state = string(answer)

	// No json response from the server, just a message we must parse
	if !strings.Contains(state, "successfully deleted") {
		err = errors.New(state)
	}

	return
}

func (g *Galaxy) NewWorkflowLauncher(historyid string, workflowid string) (launch *WorkflowLaunch) {
	launch = &WorkflowLaunch{
		historyid,
		workflowid,
		make(map[string]toolInput),
		make(map[int]map[string]string),
	}
	return
}

// Add new input file to the workflow
//
//	- inputIndex : index of this input in the workflow (see WorkflowInfo / GetWorkflowById)
//	- fielId: id of input file
//	- fielScr : one of  ["ldda", "ld", "hda", "hdca"]
func (wl *WorkflowLaunch) AddFileInput(inputIndex string, fileId string, fileSrc string) {
	wl.Inputs[inputIndex] = toolInput{fileSrc, fileId, ""}
}

// Add new parameter to workflow launcher
//
//	- fielId: id of input file
//	- fielScr : one of  ["ldda", "ld", "hda", "hdca"]
func (wl *WorkflowLaunch) AddParameter(stepIndex int, paramName string, paramValue string) {
	if _, ok := wl.Parameters[stepIndex]; !ok {
		wl.Parameters[stepIndex] = make(map[string]string)
	}
	wl.Parameters[stepIndex][paramName] = paramValue
}

// Launches the given workflow (defined by its ID), with given inputs and params.
//
// Infiles are defined by their indexes on the workflow
//
// Inparams are defined as key: step id of the workflow, value: map of key:param name/value: param value
func (g *Galaxy) LaunchWorkflow(launch *WorkflowLaunch) (answer *WorkflowInvocation, err error) {
	var url string = g.url + WORKFLOWS + "?key=" + g.apikey
	var input []byte

	answer = &WorkflowInvocation{}

	if input, err = json.Marshal(launch); err != nil {
		return
	}

	if err = g.galaxyPostRequestJSON(url, input, answer); err != nil {
		return
	}

	if answer.Err_Code != 0 || answer.Err_Msg != "" {
		err = errors.New(answer.Err_Msg)
	}
	return
}

// This function Checks the status of each step of the workflow
//
// It returns workflowstatus: An indicator of the whole workflow status,
// its step status, and its step outputfiles. The whole status is computed as follows:
//		* If all steps are "ok": then  == "ok"
//              * Else if one step is "error": then == "error"
//		* Else if one step is "deleted": then == "deleted"
//		* Else if one step is "running": then == "running"
//		* Else if one step is "queued": then == "queued"
//		* Else if one step is "waiting": then == "waiting"
//		* Else if one is is "new": then == "new"
//		* Else : Unknown state
func (g *Galaxy) CheckWorkflow(wfi *WorkflowInvocation) (wfstatus *WorkflowStatus, err error) {
	var curstate string
	var curoutfiles map[string]string
	var cumstate map[string]int
	var jobstates map[int]string
	var outfiles map[int]map[string]string

	jobstates = make(map[int]string)
	outfiles = make(map[int]map[string]string)
	cumstate = make(map[string]int)

	wfstatus = &WorkflowStatus{
		curstate,
		jobstates,
		outfiles,
	}

	numjobsids := 0
	for _, step := range wfi.Steps {
		if step.Job_Id != "" {
			numjobsids++
			if curstate, curoutfiles, err = g.CheckJob(step.Job_Id); err != nil {
				return
			} else {
				jobstates[step.Order_Index] = curstate
				outfiles[step.Order_Index] = curoutfiles
				if cursum, ok := cumstate[curstate]; !ok {
					cumstate[curstate] = 1
				} else {
					cumstate[curstate] = cursum + 1
				}
			}
		}
	}

	var sum int
	var ok bool

	if sum, ok = cumstate["ok"]; ok && sum == numjobsids {
		wfstatus.wfStatus = "ok"
	} else if sum, ok = cumstate["error"]; ok && sum >= 1 {
		wfstatus.wfStatus = "error"
	} else if sum, ok = cumstate["deleted"]; ok && sum >= 1 {
		wfstatus.wfStatus = "deleted"
	} else if sum, ok = cumstate["running"]; ok && sum >= 1 {
		wfstatus.wfStatus = "running"
	} else if sum, ok = cumstate["queued"]; ok && sum >= 1 {
		wfstatus.wfStatus = "queued"
	} else if sum, ok = cumstate["waiting"]; ok && sum >= 1 {
		wfstatus.wfStatus = "waiting"
	} else if sum, ok = cumstate["new"]; ok && sum >= 1 {
		wfstatus.wfStatus = "new"
	} else {
		wfstatus.wfStatus = "unknown"
	}
	return
}

// Cancels a running workflow.
//
// TODO: handle json response from the server in case of success... nothing described.
func (g *Galaxy) DeleteWorkflowRun(wfi *WorkflowInvocation) (err error) {
	var url string = g.url + WORKFLOWS + "/" + wfi.Workflow_Id + "/invocations/" + wfi.Id + "?key=" + g.apikey
	var answer []byte
	var galaxyErr genericError

	if answer, err = g.galaxyDeleteRequestBytes(url, []byte{}); err != nil {
		return
	}

	// If we cannot unmarshall the []HistoryShortInfo
	// The we try to unmarshall it as a galaxyError
	if err = json.Unmarshal(answer, &galaxyErr); err == nil {
		if galaxyErr.Err_Code != 0 || galaxyErr.Err_Msg != "" {
			err = errors.New(galaxyErr.Err_Msg)
		}
	}

	return
}

// Returns the global status of the workflow. It is computed as follows:
//		* If all steps are "ok": then  == "ok"
//		* Else if one step is "deleted": then == "deleted"
//		* Else if one step is "running": then == "running"
//		* Else if one step is "queued": then == "queued"
//		* Else if one step is "waiting": then == "waiting"
//		* Else if one is is "new": then == "new"
//		* Else : Unknown state
func (ws *WorkflowStatus) Status() string {
	return ws.wfStatus
}

// Gets the output file id of the given step number and having the given name
//
// If the step does not exist or the file with the given name does not exist, returns an error.
func (ws *WorkflowStatus) StepOutputFileId(stepRank int, filename string) (fileId string, err error) {
	var ok bool
	var outfiles map[string]string
	if outfiles, ok = ws.outfiles[stepRank]; !ok {
		err = errors.New(fmt.Sprintf("No Step with rank %d exists", stepRank))
	}
	if fileId, ok = outfiles[filename]; !ok {
		err = errors.New(fmt.Sprintf("No file with name %s exists for the step with rank %d", filename, stepRank))
	}
	return
}

// Gets the output file names of the given step number and having the given name
//
// If the step does not exist, returns an error.
func (ws *WorkflowStatus) StepOutFileNames(stepRank int) (fileNames []string, err error) {
	var ok bool
	var outfiles map[string]string
	if outfiles, ok = ws.outfiles[stepRank]; !ok {
		err = errors.New(fmt.Sprintf("No Step with rank %d exists", stepRank))
	}
	fileNames = make([]string, 0, len(outfiles))
	for name, _ := range outfiles {
		fileNames = append(fileNames, name)
	}
	return
}

// Gets the status of the given step number
//
// If the step number does not exist, then returns an error
func (ws *WorkflowStatus) StepStatus(stepRank int) (status string, err error) {
	var ok bool
	if status, ok = ws.stepStatus[stepRank]; !ok {
		err = errors.New(fmt.Sprintf("No Step with rank %d exists", stepRank))
	}
	return
}

// Gets the list of all the step ranks/numbers
func (ws *WorkflowStatus) ListStepRanks() (stepRanks []int) {
	stepRanks = make([]int, 0, len(ws.stepStatus))
	for rank, _ := range ws.stepStatus {
		stepRanks = append(stepRanks, rank)
	}
	sort.Ints(stepRanks)
	return
}

// Requests the given url using GET.
// and returns the response in Bytes
func (g *Galaxy) galaxyGetRequestBytes(url string) (answer []byte, err error) {
	var req *http.Request
	var response *http.Response
	var client *http.Client

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		err = g.hideKeyFromError(err)
		return
	}

	for i := 0; i < g.requestattempts; i++ {
		client = g.newClient()
		if response, err = client.Do(req); err != nil {
			err = g.hideKeyFromError(err)
			//return
		} else {
			defer response.Body.Close()
			break
		}
	}
	if err != nil {
		return
	}

	answer, err = ioutil.ReadAll(response.Body)
	return
}

// Requests the given url using GET.
// and unmarshalls the resulting expected
// resulting json into the given structure
func (g *Galaxy) galaxyGetRequestJSON(url string, answer interface{}) (err error) {
	var req *http.Request
	var response *http.Response
	var body []byte
	var client *http.Client

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		err = g.hideKeyFromError(err)
		return
	}

	for i := 0; i < g.requestattempts; i++ {
		client = g.newClient()
		if response, err = client.Do(req); err != nil {
			err = g.hideKeyFromError(err)
			//return
		} else {
			defer response.Body.Close()
			break
		}
	}
	if err != nil {
		return
	}

	if body, err = ioutil.ReadAll(response.Body); err != nil {
		return
	}

	err = json.Unmarshal(body, answer)
	return
}

// Send data to the given url using POST,
// and unmarshalls the expected resulting json into the given structure.
func (g *Galaxy) galaxyPostRequestJSON(url string, data []byte, answer interface{}) (err error) {
	var req *http.Request
	var resp *http.Response
	var body []byte
	var client *http.Client

	if req, err = http.NewRequest("POST", url, bytes.NewBuffer(data)); err != nil {
		err = g.hideKeyFromError(err)
		return
	}

	for i := 0; i < g.requestattempts; i++ {
		client = g.newClient()
		if resp, err = client.Do(req); err != nil {
			err = g.hideKeyFromError(err)
			//return
		} else {
			defer resp.Body.Close()
			break
		}
	}
	if err != nil {
		return
	}

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	err = json.Unmarshal(body, answer)
	return
}

// Requests the given url using DELETE.
// and unmarshalls the expected resulting json into the given structure
func (g *Galaxy) galaxyDeleteRequestJSON(url string, data []byte, answer interface{}) (err error) {
	var req *http.Request
	var response *http.Response
	var body []byte
	var client *http.Client

	req, _ = http.NewRequest("DELETE", url, bytes.NewBuffer(data))

	for i := 0; i < g.requestattempts; i++ {
		client = g.newClient()
		if response, err = client.Do(req); err != nil {
			err = g.hideKeyFromError(err)
			//return
		} else {
			defer response.Body.Close()
			break
		}
	}
	if err != nil {
		return
	}

	if body, err = ioutil.ReadAll(response.Body); err != nil {
		return
	}

	err = json.Unmarshal(body, answer)
	return
}

// Requests the given url using DELETE.
// and returns the response byte content
func (g *Galaxy) galaxyDeleteRequestBytes(url string, data []byte) (answer []byte, err error) {
	var req *http.Request
	var response *http.Response
	var client *http.Client

	req, _ = http.NewRequest("DELETE", url, bytes.NewBuffer(data))

	// We do requestattempts attempts
	for i := 0; i < g.requestattempts; i++ {
		client = g.newClient()
		if response, err = client.Do(req); err != nil {
			err = g.hideKeyFromError(err)
			//return
		} else {
			defer response.Body.Close()
			break
		}
	}
	if err != nil {
		return
	}

	if answer, err = ioutil.ReadAll(response.Body); err != nil {
		return
	}
	return
}

// This function replaces the api key in url that might be written
// in the error message by XXXXXXXXXXXXXXXXXX
func (g *Galaxy) hideKeyFromError(inerr error) (outerr error) {
	newMessage := strings.Replace(inerr.Error(), g.apikey, "XXXXXXXXXXXXXXXXXX", -1)
	outerr = errors.New(newMessage)
	return
}
