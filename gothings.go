package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pepemontana7/gothings/godevice"
	"github.com/pepemontana7/osin"
)

//Trace logger
var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

//InitLogger configures log params.
func InitLogger(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

type myResponse struct {
	code     string
	tokenUrl string
}

func (mr *myResponse) MarshalJSON() ([]byte, error) {
	m := make(map[string]string)
	m["auth_code"] = mr.code
	m["token_url"] = mr.tokenUrl
	return json.Marshal(m)
}

type device_action struct {
	Action string
}

// InfoRequest is a request for information about some AccessData
type ResourceRequest struct {
	Code       string           // Code to look up
	AccessData *osin.AccessData // AccessData associated with Code
}

// HandleInfoRequest is an http.HandlerFunc for server information
// NOT an RFC specification.
func HandleResourceRequest(s *osin.Server, w *osin.Response, r *http.Request) *ResourceRequest {
	r.ParseForm()
	bearer := osin.CheckBearerAuth(r)
	if bearer == nil {
		w.SetError(osin.E_INVALID_REQUEST, "")
		return nil
	}

	// generate info request
	ret := &ResourceRequest{
		Code: bearer.Code,
	}

	if ret.Code == "" {
		w.SetError(osin.E_INVALID_REQUEST, "")
		return nil
	}

	var err error

	// load access data
	ret.AccessData, err = w.Storage.LoadAccess(ret.Code)
	if err != nil {
		w.SetError(osin.E_INVALID_REQUEST, "")
		w.InternalError = err
		return nil
	}
	if ret.AccessData == nil {
		w.SetError(osin.E_INVALID_REQUEST, "")
		return nil
	}
	if ret.AccessData.Client == nil {
		w.SetError(osin.E_UNAUTHORIZED_CLIENT, "")
		return nil
	}
	if ret.AccessData.Client.GetRedirectUri() == "" {
		w.SetError(osin.E_UNAUTHORIZED_CLIENT, "")
		return nil
	}
	if ret.AccessData.IsExpiredAt(s.Now()) {
		w.SetError(osin.E_INVALID_GRANT, "")
		return nil
	}

	return ret
}

// FinishInfoRequest finalizes the request handled by HandleInfoRequest
func FinishResourceRequest(w *osin.Response, r *http.Request, rs *ResourceRequest) {
	// don't process if is alrseady an error
	if w.IsError {
		return
	}

	//idStr := strings.TrimPrefix(r.URL.Path, "/devices/")
	//id, err := strconv.Atoi(idStr)
	fmt.Println("id", r.Header.Get("id"))
	switch r.Method {
	case "GET":
		// Retrieve that device from the DB.
		if r.Header.Get("id") != "" {
			id, err := strconv.Atoi(r.Header.Get("id"))
			d, err := godevice.Find(id)
			if err != nil {
				w.SetError(err.Error(), "")
				return
			}
			w.Output["data"] = d

		} else {
			devices := godevice.All()
			if devices == nil {
				w.SetError("No Devices", "")
			}
			w.Output["data"] = devices

		}
		w.Output["status"] = http.StatusOK
		w.Output["scuccess"] = "yes"

	case "POST":
		if r.Header.Get("id") == "" {
			w.SetError("Post not allowed temporarily in /devices endpoint", "")
			w.StatusCode = 405
			return
		}
		id, err := strconv.Atoi(r.Header.Get("id"))
		d, err := godevice.Find(id)
		if err != nil {
			w.SetError(err.Error(), "")
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(body))
		var da device_action
		err = json.Unmarshal(body, &da)
		if err != nil {
			panic(err)
		}
		allowedAction := false
		for _, b := range d.Actions {
			if b == da.Action {
				allowedAction = true
			}
		}
		if allowedAction == false {
			w.SetError("Action not allowed "+da.Action+" for device "+r.Header.Get("id"), "")
			w.Output["scuccess"] = "no"
			w.StatusCode = 405
		} else {
			w.Output["id"] = r.Header.Get("id")
			w.Output["action"] = da.Action
			w.Output["scuccess"] = "yes"
		}

	}

}

func main() {

	hostPtr := flag.String("endpoint", "http://localhost:14000", "Ngrok endpoint https://ngr.ngrok.com or http://localhost:14000. (Required)")
	redirectPtr := flag.String("redirect", "http://localhost:14000/appauth", "Redirect uri. (Required)")
	defIDPtr := flag.String("client-id", "1234", "Default client id  (Required)")
	defSecretPtr := flag.String("client-secret", "aabbccdd", "Default client secret. (Required)")

	flag.Parse()

	// Open a file for logs
	gothings, err := os.OpenFile("gothings.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Error.Fatalln("Failed to open gothings log file")
	}
	defer gothings.Close()
	// Logging all option to  send to ioutil.Discard
	InitLogger(gothings, gothings, gothings, gothings)

	rtr := mux.NewRouter()
	cfg := osin.NewServerConfig()
	cfg.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE, osin.REFRESH_TOKEN}
	cfg.AllowGetAccessRequest = true
	cfg.AllowClientSecretInParams = true
	server := osin.NewServer(cfg, NewTestStorage(*defIDPtr, *defSecretPtr, *redirectPtr))

	//Temp workaround for client, until user mgmt is avail
	//client := server.Storage.GetClient("1234")

	// Authorization code endpoint
	rtr.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		Info.Println("Authorize Endpoint called")

		resp := server.NewResponse()
		defer resp.Close()

		if ar := server.HandleAuthorizeRequest(resp, r); ar != nil {
			if !HandleLoginPage(ar, w, r) {
				return
			}
			ar.Authorized = true
			server.FinishAuthorizeRequest(resp, r, ar)
		}
		if resp.IsError && resp.InternalError != nil {
			Error.Printf("Internal Error: %s", resp.InternalError)
		}
		osin.OutputJSON(resp, w, r)
		Info.Println("Resp, req, respwriter: ", resp, r, w)
	})

	// Access token endpoint
	rtr.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		Info.Println("Reached Token endpoint")
		resp := server.NewResponse()
		defer resp.Close()

		if ar := server.HandleAccessRequest(resp, r); ar != nil {
			ar.Authorized = true
			server.FinishAccessRequest(resp, r, ar)
		}
		if resp.IsError && resp.InternalError != nil {
			Error.Printf("Internal Error: %s", resp.InternalError)
		}
		osin.OutputJSON(resp, w, r)
	})

	// Resource Device endpoint
	rtr.HandleFunc("/devices/{id}", func(w http.ResponseWriter, r *http.Request) {
		resp := server.NewResponse()
		defer resp.Close()
		vars := mux.Vars(r)
		Info.Println("Device id ", vars["id"])
		r.Header.Set("id", vars["id"])
		if ir := HandleResourceRequest(server, resp, r); ir != nil {
			FinishResourceRequest(resp, r, ir)
		}
		osin.OutputJSON(resp, w, r)
	})

	rtr.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		resp := server.NewResponse()
		defer resp.Close()
		if ir := HandleResourceRequest(server, resp, r); ir != nil {
			FinishResourceRequest(resp, r, ir)
		}
		osin.OutputJSON(resp, w, r)
	})

	// Application home endpoint
	rtr.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		u := strings.Join([]string{*hostPtr, "appauth/code"}, "/")
		w.Write([]byte("<html><body>"))
		w.Write([]byte(fmt.Sprintf("<a href=\"/authorize?response_type=code&client_id=%s&state=xyz&scope=everything&redirect_uri=%s\">Login</a><br/>", *defIDPtr, url.QueryEscape(u))))
		w.Write([]byte("</body></html>"))
	})

	// Application destination - CODE
	rtr.HandleFunc("/appauth/code", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		code := r.Form.Get("code")

		if code == "" {
			Error.Println("Error no code")
			return
		}
		var mr myResponse
		u := strings.Join([]string{*hostPtr, "appauth/code"}, "/")
		// build access code url
		aurl := fmt.Sprintf("/token?grant_type=authorization_code&client_id=1234&client_secret=aabbccdd&state=xyz&redirect_uri=%s&code=%s",
			url.QueryEscape(u), url.QueryEscape(code))
		mr.tokenUrl = aurl
		mr.code = url.QueryEscape(code)
		Info.Println("TokenURL, Code: ", mr.tokenUrl, mr.code)

		data, err := mr.MarshalJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return

	})

	http.ListenAndServe(":14000", rtr)
}
