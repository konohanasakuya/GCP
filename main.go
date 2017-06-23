package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/file"
	"google.golang.org/appengine/urlfetch"

	"cloud.google.com/go/storage"
)

// impoert "cloud.google.com/go/storage"
//   Note: Old version was compile error: cloud.google.com/go/storage/copy.go:81
//   dep: github.com/googleapis/gax-go
//   dep: google.golang.org/api/storage/v1

var localPc = false

// func main() {}

var (
	url             = "https://fcm.googleapis.com/fcm/send"
	token           = "fiJq5u36pH8:APA91bGIqyflcPyUuY03n1mL-HRiC5bjFrk4BSecIrFpuX1aqb41B1g9ppmTN8xVbdb8IULXN1dcCIIZjXkX-RtGc6YGovvCPoBiNd3jkIBrwne44nCRX9wXSVz9OcDyZSZfIPpAR1ZM"
	title           = "TITLE"
	api_key         = "AIzaSyBIwwzRJ_kHIqv0BankBrKclN8rvPHMX38"
	jsonStr         string
	fcmResp         string
	beforePath      = "index.html"
	test1Url        = "https://testproject2-168808.firebaseio.com/persons.json"
	deviceFile      = "https://testproject2-168808.firebaseio.com/devices.json"
	deviceFileLocal = "devices.json"
)

type Device struct {
	Id        string `json:"id"`
	Board     string `json:"board"`
	Os        string `json:"os"`
	OsVersion string `json:"version"`
	Time      int64  `json:"time"`
}

type Person struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Birthday string `json:"birthday"`
}

var msgNotFound = fmt.Sprintf(`<html><body><font size="5" color="#ff0000">Not Found : %d</font></body></html>`, http.StatusNotFound)

func getFirebaseDataStorage_local(w http.ResponseWriter, r *http.Request) {
	fmt.Println("run getFirebaseDataStorage_local")

	// --------------------------------------
	// ローカルPCでやる場合はこっち
	// --------------------------------------
	req, err := http.NewRequest("GET", test1Url, nil)
	if err != nil {
		panic(err)
	}
	client := &http.Client{Transport: http.DefaultTransport}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		panic(err)
	}

	byteArray, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(byteArray))

	var persons []Person
	if err := json.Unmarshal(byteArray, &persons); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	tpl, err1 := template.ParseFiles("test1.tmpl")
	if err1 != nil {
		panic(err1)
	}

	err2 := tpl.Execute(w, struct {
		List []Person
	}{
		List: persons,
	})
	if err2 != nil {
		fmt.Fprintf(w, "%s", err2.Error())
	}
}

func getFirebaseDataStorage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("run getFirebaseDataStorage")

	// --------------------------------------
	// AppEngine上でやる場合はこっち
	// --------------------------------------
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	resp, err := client.Get(test1Url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	byteArray, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(byteArray))

	var persons []Person
	if err := json.Unmarshal(byteArray, &persons); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	tpl, err1 := template.ParseFiles("test1.tmpl")
	if err1 != nil {
		panic(err1)
	}

	err2 := tpl.Execute(w, struct {
		List []Person
	}{
		List: persons,
	})
	if err2 != nil {
		fmt.Fprintf(w, "%s", err2.Error())
	}
}

func sendFCM_local(w http.ResponseWriter, r *http.Request, body string) (err error) {
	fmt.Println("run sendFCM_local")

	jsonStr = `{"to":"` + token + `","notification":{"title":"` + title + `","body":"` + body + `"}}`
	fmt.Printf("\n[[ jsonStr ]]\n%s\n", jsonStr)

	// --------------------------------------
	// ローカルPCでやる場合はこっち
	// --------------------------------------
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		return fmt.Errorf("Error NewRequest : %s", err.Error())
	}
	client := &http.Client{Transport: http.DefaultTransport}

	req.Header.Set("Authorization", "key="+api_key)
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")

	fmt.Println("\n[[ Http Header ]]")
	fmt.Println("Authorization:", req.Header.Get("Authorization"))
	fmt.Println("Content-Type:", req.Header.Get("Content-Type"))

	fmt.Println("\n[[ req ]]")
	fmt.Println(req)

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("Error Do : %s\n", err.Error())
	}

	fmt.Printf("\n[[ resp.Status ]]\n%s\n", resp.Status)

	return nil
}

func sendFCM(w http.ResponseWriter, r *http.Request, body string) (err error) {
	fmt.Println("run sendFCM")
	jsonStr = `{"to":"` + token + `","notification":{"title":"` + title + `","body":"` + body + `"}}`
	fmt.Printf("\n[[ jsonStr ]]\n%s\n", jsonStr)

	// --------------------------------------
	// AppEngine上でやる場合はこっち
	// --------------------------------------
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "key="+api_key)
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")

	fmt.Println("\n[[ Http Header ]]")
	fmt.Println("Authorization:", req.Header.Get("Authorization"))
	fmt.Println("Content-Type:", req.Header.Get("Content-Type"))

	fmt.Println("\n[[ req ]]")
	fmt.Println(req)

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("Error Do : %s\n", err.Error())
	}

	fmt.Printf("\n[[ resp.Status ]]\n%s\n", resp.Status)

	return nil
}

func pageLoader(title string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("run pageLoader")

	defer func() {
		ret := recover()
		if ret != nil {
			w.Header().Set("Content-type", "text/html")
			w.Write([]byte(msgNotFound))
		}
	}()
	tpl := template.Must(template.ParseFiles(title))

	err := tpl.Execute(w, nil)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func testPageLoader(title string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("run viewTemplate")

	defer func() {
		ret := recover()
		if ret != nil {
			w.Write([]byte(msgNotFound))
		}
	}()
	tpl := template.Must(template.ParseFiles(title))

	err := tpl.Execute(w, struct {
		Resp string
	}{
		Resp: fcmResp,
	})
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func testGetStorage(w http.ResponseWriter, r *http.Request) {
	objname := "devicedata/main.go"

	// Requestから、Contextを取得
	ctx := appengine.NewContext(r)

	// Contextから、default bucket nameを取得
	bucketname, err := file.DefaultBucketName(ctx)
	if err != nil {
		return
	}

	// Contextから、Clientの取得
	client, err := storage.NewClient(ctx)
	if err != nil {
		return
	}

	// クライアントにバケット名を食わせてBucketHandleを取得し、
	// それにオブジェクト名を食わせてObjectHandleを取得し、
	// そこにContextを食わせてObjectへのReaderを取得
	reader, err := client.Bucket(bucketname).Object(objname).NewReader(ctx)
	if err != nil {
		return
	}
	defer reader.Close()

	// CloudStorage上のObjectの、コンテンツの読み込み
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	w.Header().Set("Content-type", "text/html")
	fmt.Fprintf(w, "<html><body><h1>%s</h1><pre><p>%s</p></pre></body></html>", objname, string(body))
}

// Google Cloud Storage への WRITE
//        ctx := appengine.NewContext(r)
//        bucketname, err := file.DefaultBucketName(ctx)
//        client, err := storage.NewClient(ctx)
//        body, err := ioutil.ReadAll(f)
//        writer := client.Bucket(bucketname).Object(objname).NewWriter(ctx)
//        writer.ContentType = "text/plain"
//        defer writer.Close()
//        if _, err := writer.Write(body); err != nil {

func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("run testHandler")

	path := r.URL.Path[1:]

	fmt.Println("r.Method:", r.Method, ", beforePath:", beforePath)
	fmt.Println("path:", path)

	if r.Method != "GET" {
		path = beforePath
	}

	switch path {
	case "test/1":
		if localPc == true {
			getFirebaseDataStorage_local(w, r)
		} else {
			getFirebaseDataStorage(w, r)
		}

	case "test/2":
		if r.Method == "POST" {
			input_value := r.FormValue("input_value")
			if input_value != "" {
				if localPc == true {
					err := sendFCM_local(w, r, input_value)
					if err != nil {
						fmt.Println(err.Error())
					}
				} else {
					err := sendFCM(w, r, input_value)
					if err != nil {
						fmt.Println(err.Error())
					}
				}
			}
		}
		testPageLoader("test2.tmpl", w, r)

	case "test/3":
		testGetStorage(w, r)
	}

	beforePath = path
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("run viewHandler")

	path := r.URL.Path[1:]

	if r.Method != "GET" {
		testHandler(w, r)
		return
	}

	fmt.Println("r.Method:", r.Method, ", beforePath:", beforePath)
	fmt.Println("path:", path)
	if path == "" {
		path = "index.html"
	}
	pageLoader(path, w, r)

	beforePath = path
}

type Logs struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type DeviceData struct {
	Info struct {
		Os        string `json:"os"`
		OsVersion string `json:"osversion"`
		Board     string `json:"board"`
	} `json:"info"`
	Time  string `json:"time,omitempty"`
	Times string `json:"times,omitempty"`
	Logs  []Logs `json:"logs,omitempty"`
}

func writeDataMap(w http.ResponseWriter, r *http.Request, path string, dataMap *map[string]DeviceData) error {
	if localPc {
		if len(*dataMap) == 0 {
			return os.Remove(path)
		}
		byteArray, _ := json.Marshal(*dataMap)
		return ioutil.WriteFile(path, byteArray, os.FileMode(0666))
	}

	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	byteArray, _ := json.Marshal(*dataMap)
	jsonData := string(byteArray)
	fmt.Fprintln(w, "jsonData:", jsonData)
	req, err := http.NewRequest("PATCH", path, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	if resp, _ := client.Do(req); resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("Error Do : %s\n", err.Error())
	}
	return nil
}

func readDataMap(w http.ResponseWriter, r *http.Request, path string) (dataMap map[string]DeviceData, err error) {
	if localPc {
		byteArray, err := ioutil.ReadFile(path)
		if err != nil {
			// do nothing
			return nil, err
		}
		err = json.Unmarshal(byteArray, &dataMap)
		return dataMap, err
	}

	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	resp, err := client.Get(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	byteArray, _ := ioutil.ReadAll(resp.Body)

	if err := json.Unmarshal(byteArray, &dataMap); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	defer resp.Body.Close()

	return
}

func devicePost(w http.ResponseWriter, r *http.Request, id string, deviceData *DeviceData) {
	var path string
	if localPc {
		path = deviceFileLocal
	} else {
		path = deviceFile
	}

	dataMap, err := readDataMap(w, r, path)
	if err != nil || dataMap == nil {
		/* Create Data */
		dataMap = map[string]DeviceData{}
	}

	if _, ok := dataMap[id]; ok {
		fmt.Println("Already exist (update)")
	}
	dataMap[id] = *deviceData
	if err := writeDataMap(w, r, path, &dataMap); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func deviceDelete(w http.ResponseWriter, r *http.Request, id string) {
	// ToDo
}

func deviceShow(w http.ResponseWriter, r *http.Request) {
	var path string
	if localPc {
		path = deviceFileLocal
	} else {
		path = deviceFile
	}

	DataMap, err := readDataMap(w, r, path)
	if err != nil || DataMap == nil {
		fmt.Fprintln(w, "Device Empty")
		return
	}

	tpl := template.Must(template.ParseFiles("devices.tmpl"))
	err = tpl.Execute(w, DataMap)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

// >>>>========================================================>>>>

// 「/save」用のハンドラ
func saveHandler(w http.ResponseWriter, r *http.Request) {
	token := ""
	mode := ""
	result := ""
	filename := ""
	// ResultMessage = ""

	filenum := 0
	var buffers [5]bytes.Buffer
	var filenames [5]string

	// MultipartReaderを用いて受け取ったファイルを読み込み
	reader, err := r.MultipartReader()
	// エラーが発生したら抜ける
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		switch part.FormName() {
		case "token":
			buf := make([]byte, 256)
			len, _ := part.Read(buf)
			token = string(buf[:len])
			fmt.Println("token:", token)
		case "mode":
			buf := make([]byte, 256)
			len, _ := part.Read(buf)
			mode = string(buf[:len])
			fmt.Println("mode:", mode)
		case "file":
			if part.FileName() == "" {
				continue
			}
			io.Copy(&buffers[filenum], part)
			filenames[filenum] = part.FileName()
			filenum++
		}
	}
	if token == "" || mode == "" {
		fmt.Fprintf(w,
			"<html><body>Upload Failure (invalid parameters) : token:%s<br/>mode:%s<br/></body></html>",
			token, mode)
		return
	}

	if mode == "command" {
		for i := 0; i < filenum; i++ {
			filename = filenames[i]
			times := time.Now().UTC().In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format("2006-01-02 15:04:05")
			path := "devices/" + token + "/" + mode + "/" + times + "/" + filename
			fmt.Printf("file name:%s\nsave path:\n", filename, path)
			result = putStorage(path, &buffers[i], r)
			fmt.Println("result:", result)
		}

	} else if mode == "command1" {
		// fmt.Fprintln(w, "@TEST@ command1")

		bytesReader := bytes.NewReader(buffers[0].Bytes())
		// rd, err := zip.NewReader(bytesReader, int64(len(buffers[0].Bytes())))
		rd, _ := zip.NewReader(bytesReader, int64(len(buffers[0].Bytes())))
		// defer r.Close() // ファイルクローズ

		// ファイルの中身を１つずつ処理
		for _, f := range rd.File {
			// ファイルオープン
			rc, err := f.Open()
			if err != nil { // エラー
				return
			}
			defer rc.Close() // ファイルクローズ

			// フォーカスしているのがディレクトリのとき
			if f.FileInfo().IsDir() {
				// ディレクトリを作成
				//	path := filepath.Join(dest, f.Name) // 作成するディレクトリのパスを作成
				//	os.MkdirAll(path, f.Mode())         // ディレクトリを作成
				// フォーカスしているのがファイルのとき
			} else {
				// ファイルの中身をメモリ（スライス）に読み出す
				buf := make([]byte, f.UncompressedSize) // メモリ確保
				_, err = io.ReadFull(rc, buf)           // メモリに書き込む
				if err != nil {                         // 書き込めなかったとき
					return
				}
				ResultMessage = "<< " + f.Name + ">>\n"
				ResultMessage = ResultMessage + string(buf) + "\n"
				// ファイルを作成
				//	path := filepath.Join(dest, f.Name)          // ファイルのパスを作成
				//	err := ioutil.WriteFile(path, buf, f.Mode()) // ファイルを作成して中身を書き込む
				//	if err != nil {                              // エラー
				//		return err
				//	}
			}
		}
		fmt.Fprintln(w, "@TEST@ ", ResultMessage)
		// http.Redirect(w, r, "/command1", 303)
	}
}

// 「/upload」用のハンドラ
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	var templatefile = template.Must(template.ParseFiles("upload.html"))
	templatefile.Execute(w, "upload.html")
}

// 「/errorPage」用のハンドラ
func errorPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", "<p>Internal Server Error</p>")
}

// errorが起こった時にエラーページに遷移する
func redirectToErrorPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/errorpage", http.StatusFound)
}

func putStorage(path string, buff *bytes.Buffer, r *http.Request) string {
	ctx := appengine.NewContext(r)

	bucketname, err := file.DefaultBucketName(ctx)
	if err != nil {
		return "file.DefaultBucketName(ctx) : " + err.Error()
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return "storage.NewClient(ctx) : " + err.Error()
	}

	// Writer取得
	writer := client.Bucket(bucketname).Object(path).NewWriter(ctx)
	writer.ContentType = "application/zip"
	defer writer.Close()

	// コンテンツを書き込む
	if _, err := writer.Write(buff.Bytes()); err != nil {
		return "writer.Write(buff.Bytes()) : " + err.Error()
	}

	return "Success"
}

// <<<<========================================================<<<<

func deviceRegisterHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("run deviceRegisterHandler")

	var dataMap map[string]DeviceData
	byteArray, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(byteArray, &dataMap)
	var id string
	var deviceData DeviceData
	for id = range dataMap {
		deviceData = dataMap[id]
	}

	switch r.Method {
	case "POST", "PUT", "PATCH":
		deviceData.Time = strconv.FormatInt(time.Now().Unix(), 10)
		deviceData.Times = time.Now().UTC().In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format("2006-01-02 15:04:05")
		devicePost(w, r, id, &deviceData)
	case "DELETE":
		deviceDelete(w, r, id)
	case "GET":
		deviceShow(w, r)
	}
}

// >>>>========================================================>>>>
var ResultMessage string = "シングル送信の場合、結果がここに表示されます"

func commandHandler1(w http.ResponseWriter, r *http.Request) {
	fmt.Println("run commandHandler1")

	if r.Method == "POST" {
		// ResultMessage = ""
		token := r.FormValue("token")
		command := r.FormValue("command")

		// resp, err := sendCommand(w, r, []string{token}, mode)
		_, err := sendCommand(w, r, []string{token}, command, true)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	defer func() {
		ret := recover()
		if ret != nil {
			w.Write([]byte(msgNotFound))
		}
	}()
	//	tpl := template.Must(template.ParseFiles("command.tmpl"))
	tpl, _ := template.ParseFiles("command.tmpl")
	//err := tpl.Execute(w, ResultMessage)
	err := tpl.Execute(w, struct {
		Result string
	}{
		Result: ResultMessage,
	})
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func commandHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("run commandHandler")

	if r.Method == "POST" {
		// ResultMessage = ""
		token := r.FormValue("token")
		command := r.FormValue("command")

		// resp, err := sendCommand(w, r, []string{token}, mode)
		_, err := sendCommand(w, r, []string{token}, command, false)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	defer func() {
		ret := recover()
		if ret != nil {
			w.Write([]byte(msgNotFound))
		}
	}()
	//	tpl := template.Must(template.ParseFiles("command.tmpl"))
	tpl, _ := template.ParseFiles("command.tmpl")
	//err := tpl.Execute(w, ResultMessage)
	err := tpl.Execute(w, struct {
		Result string
	}{
		Result: ResultMessage,
	})
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func deviceShow2(w http.ResponseWriter, r *http.Request) {
	var path string
	if localPc {
		path = deviceFileLocal
	} else {
		path = deviceFile
	}

	DataMap, err := readDataMap(w, r, path)
	if err != nil || DataMap == nil {
		fmt.Fprintln(w, "Device Empty")
		return
	}

	tpl := template.Must(template.ParseFiles("devices.tmpl"))
	err = tpl.Execute(w, DataMap)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func sendCommand(w http.ResponseWriter, r *http.Request, tokens []string, command string, bool oneshot) (resp *http.Response, err error) {
	fmt.Println("run sendCommand")
	jsonStr := ""

	if len(tokens) == 1 {
		if oneshot {
			jsonStr = `{"to":"` + tokens[0] + `","data":{"mode":"command1","action":"` + command + `"}}`
		} else {
			jsonStr = `{"to":"` + tokens[0] + `","data":{"mode":"command","action":"` + command + `"}}`
		}
	} else {
		token := string(tokens[0])
		jsonStr = `{"registration_ids":["` + token + `"`
		ids := tokens[1:]
		for _, id := range ids {
			jsonStr = jsonStr + `,"` + id + `"`
		}
		jsonStr = `],"data":{"mode":"command","action":"` + command + `"}}`
	}
	fmt.Printf("\n[[ jsonStr ]]\n%s\n", jsonStr)

	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		//		http.Error(w, err.Error(), http.StatusInternalServerError)
		return resp, err
	}

	req.Header.Set("Authorization", "key="+api_key)
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")

	fmt.Println("\n[[ req ]]")
	fmt.Println(req)

	resp, err = client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return resp, fmt.Errorf("Error Do : %s\n", err.Error())
	}

	fmt.Printf("\n[[ resp.Status ]]\n%s\n", resp.Status)

	return resp, nil
}

// <<<<========================================================<<<<

func init() {
	// ハンドラ登録
	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/test/", testHandler)
	http.HandleFunc("/devices", deviceRegisterHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/save", saveHandler)
	http.HandleFunc("/errorpage", errorPageHandler)
	http.HandleFunc("/command1", commandHandler1)
	http.HandleFunc("/command", commandHandler)

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css/"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("fonts/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images/"))))

	// HTTPサーバ起動
	fmt.Println("http server start")
	http.ListenAndServe(":8080", nil)
}
