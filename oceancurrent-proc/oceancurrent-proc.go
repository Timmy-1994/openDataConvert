package main

/*
* 中央氣象局open data
* 海流模式-海流數值模式預報資料-第000小時
* https://opendata.cwb.gov.tw/dataset/climate/M-B0071-000
* 將2D經緯度資料轉為1D-array
* 經緯度 7, 119 >> 7, 126; 7.1, 119 >> 7.1, 126; .... ; 36, 126
*/

import (
	"flag"
	"log"
	"time"
	"fmt"

	"io"
	"os"
	"strings"
	"strconv"
	"errors"

	"encoding/json"
	"encoding/xml"
	"math"

	"bytes"
	"crypto/tls"
	"mime/multipart"
	"net"
	"net/http"
	"io/ioutil"
)

var (
	inFile = flag.String("i", "M-B0071-000.xml", "input XML file")
	outFile = flag.String("o", "M-B0071-000.grid.json", "output file")

	proxyAddr = flag.String("x", "", "socks5 proxy addr (127.0.0.1:5005)")
	connTimeout = flag.Int("timeout", 10, "connect timeout in Seconds")

	token = flag.String("auth", "CWB-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX", "token") // 氣象局open data的API授權碼
	url = flag.String("u", "https://opendata.cwb.gov.tw/fileapi/v1/opendataapi/M-B0071-000?Authorization=%v&downloadType=WEB&format=XML", "url")
	UA = flag.String("ua", "OAC bot", "User-Agent")

	verbosity = flag.Int("v", 3, "verbosity for app")

	hookUrl = flag.String("hook", "http://127.0.0.1:8080/api/push/89HuRzqCRlRGIrhSifYN", "web hook URL")
)

func main() {
	flag.Parse()

	if *token == "" {
		transFile(*inFile, *outFile)
		return
	}

	aurl := fmt.Sprintf(*url, *token)
	dialFunc := func(network, address string) (net.Conn, error) {
		return net.DialTimeout("tcp", address, time.Duration(*connTimeout) * time.Second)
	}
	if *proxyAddr != "" {
		dialFunc = func(network, address string) (net.Conn, error) {
			if network != "tcp" {
				return nil, errors.New("only support tcp")
			}
			return makeConnection(address, *proxyAddr, time.Duration(*connTimeout) * time.Second)
		}
	}

	fd, err := getUrlFd(aurl, dialFunc)
	if err != nil {
		Vln(2, "[get]err", aurl, err)
		return
	}
	defer fd.Close()

	grid, err := parseXML(fd)
	if err != nil {
		Vln(2, "[parse]err", err)
		return
	}
	Vln(3, "[grid]", grid.Nx, grid.Ny)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err = enc.Encode(grid)
	if err != nil {
		Vln(2, "[json]err", err)
	}
	Vln(3, "[json]ok")

	if *hookUrl != "" {
		postUrl(*hookUrl, *outFile, &buf)
		Vln(3, "[post]", *hookUrl)
	} else {
		of, err := os.OpenFile(*outFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			Vln(2, "[open]err", *outFile, err)
			return
		}
		defer of.Close()

		_, err = io.Copy(of, &buf)
		if err != nil {
			Vln(2, "[write]err", *outFile, err)
			return
		}
	}
}

func transFile(inFp string, outFp string) {
	fd, err := os.OpenFile(inFp, os.O_CREATE|os.O_RDONLY, 0400)
	if err != nil {
		Vln(2, "[open]err", *inFile, err)
		return
	}
	defer fd.Close()

	grid, err := parseXML(fd)
	if err != nil {
		Vln(2, "[parse]err", err)
		return
	}
	Vln(3, "[grid]", grid.Nx, grid.Ny)

	of, err := os.OpenFile(outFp, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		Vln(2, "[open]err", *outFile, err)
		return
	}
	defer of.Close()

	enc := json.NewEncoder(of)
	err = enc.Encode(grid)
	if err != nil {
		Vln(2, "[json]err", err)
	}
}

func getUrl(url string, dialFunc func(network, addr string) (net.Conn, error)) ([]byte, error) {
	resBody, err := getUrlFd(url, dialFunc)
	if err != nil {
		return nil, err
	}
	defer resBody.Close()

	data, err := ioutil.ReadAll(resBody)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getUrlFd(url string, dialFunc func(network, addr string) (net.Conn, error)) (io.ReadCloser, error) {
	var netTransport = &http.Transport{
		Dial: dialFunc,
		TLSHandshakeTimeout: time.Duration(*connTimeout) * time.Second,
	}

	var netClient = &http.Client{
		Timeout: time.Second * 180,
		Transport: netTransport,
	}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", *UA)
	req.Close = true
	res, err := netClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func postUrl(url string, fileName string, data io.Reader) ([]byte, error) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var netClient = &http.Client{
		Timeout: time.Second * 60,
		Transport: netTransport,
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fw, data)
	if err != nil {
		return nil, err
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	req, err := http.NewRequest("POST", url, nil)
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", *UA)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Close = true
	//req.Body = ioutil.NopCloser(data)
	req.Body = ioutil.NopCloser(&b)
	res, err := netClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	ret, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return ret, nil
}


type VectorGrid struct {
	// 原點 經度, 緯度
	Lo1 float32 `json:"lo1"`
	La1 float32 `json:"la1"`

	// 終點 經度, 緯度
	Lo2 float32 `json:"lo2"`
	La2 float32 `json:"la2"`

	Nx int `json:"nx"` // 經度格數
	Ny int `json:"ny"` // 緯度格數

	Time string `json:"time"` // just copy now
	Desc string `json:"Description"`  // just copy

	DataRange map[string][]jsonFloat `json:"drange"`

	Data map[string][]jsonFloat `json:"d"`
}

type jsonFloat float32
func (value jsonFloat) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(value)) {
		return []byte("\"\""), nil
	}
	return []byte(fmt.Sprintf("%v", value)), nil
}

func NewVectorGrid() *VectorGrid {
	vg := &VectorGrid{}
	vg.Data = make(map[string][]jsonFloat, 2)
	vg.DataRange = make(map[string][]jsonFloat, 2)
	return vg
}

func parseXML(r io.Reader) (*VectorGrid, error) {
	grid := NewVectorGrid()

	ps := &procState{}
	xs := NewXMLState()
	decoder := xml.NewDecoder(r)
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return grid, nil
			}
			return grid, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			stelm := xml.StartElement(t)
			//Vln(5, "start: ", stelm.Name.Local)
			xs.StartTag(stelm)

		case xml.EndElement:
			endelm := xml.EndElement(t)
			//Vln(5, "end: ", endelm.Name.Local)
			xs.EndTag(endelm)

		case xml.CharData:
			data := xml.CharData(t)
			ps.FillTag(xs, data, grid)

			//str := string(data)
			//Vln(5, "[val]", xs.GetPath(), str)
		}
	}

	return grid, nil
}

type procState struct {
	st int
	valName string

	lat float32
	lon float32
	lat0 float32
	lon0 float32
	lat1 float32
	lon1 float32
}
func (ps *procState) FillTag(xs *XmlState, data []byte, grid *VectorGrid) {
	switch ps.st {
	case 0:
		path := xs.GetPath()
		switch path {
		case "cwbopendata/dataset/datasetInfo/datasetDescription":
			str := string(data)
			Vln(3, "[desc]", path, str)
			grid.Desc = str
		case "cwbopendata/dataset/datasetInfo/parameterSet/parameter/parameterName":
			Vln(3, "[n?]", path, string(data))
		case "cwbopendata/dataset/datasetInfo/parameterSet/parameter/parameterValue":
			str := string(data)
			Vln(3, "[ny]", path, str)
			if v, err := strconv.ParseUint(str, 10, 32); err == nil {
				grid.Ny = int(v)
			}
		case "cwbopendata/dataset/time/datetime":
			str := string(data)
			Vln(3, "[time]", path, str)
			grid.Time = str
		case "cwbopendata/dataset/location":
			ps.st = 1 // start parse grids
			ps.lat0 = 9999
			ps.lon0 = 9999
			ps.lat1 = -9999
			ps.lon1 = -9999
		}
	case 1:
		tag := xs.LastPath()
		switch tag {
		case "lat": // 緯度
			if v, err := strconv.ParseFloat(string(data), 32); err == nil {
				ps.lat = float32(v)
				if ps.lat < ps.lat0 {
					ps.lat0 = ps.lat
				}
				if ps.lat > ps.lat1 {
					ps.lat1 = ps.lat
				}
			}
		case "lon": // 經度
			if v, err := strconv.ParseFloat(string(data), 32); err == nil {
				ps.lon = float32(v)
				if ps.lon < ps.lon0 {
					ps.lon0 = ps.lon
				}
				if ps.lon > ps.lon1 {
					ps.lon1 = ps.lon
				}
			}
		case "elementName":
			str := string(data)
			switch str {
			case "橫向流速":
				ps.valName = "X"
				//Vln(7, "[pos]", ps.lat, ps.lon)
			case "直向流速":
				ps.valName = "Y"
			case "海表溫度", "海高", "海表鹽度":
				ps.valName = str
			default:
				ps.valName = ""
			}
		case "value":
			if ps.valName == "" {
				break
			}
			arr, ok := grid.Data[ps.valName]
			if !ok {
				arr = make([]jsonFloat, 0, grid.Ny)
			}
			if v, err := strconv.ParseFloat(string(data), 32); err == nil {
				arr = append(arr, jsonFloat(v))
				grid.Data[ps.valName] = arr

				if !math.IsNaN(v) {
					// minimum and maximum
					minMax, ok := grid.DataRange[ps.valName]
					if !ok {
						minMax = []jsonFloat{jsonFloat(v), jsonFloat(v)}
						grid.DataRange[ps.valName] = minMax
					}
					if v < float64(minMax[0]) {
						minMax[0] = jsonFloat(v)
					}
					if v > float64(minMax[1]) {
						minMax[1] = jsonFloat(v)
					}
				}
			}
		case "cwbopendata": // end dataset
			ps.st = 2
			grid.Nx = len(grid.Data["X"]) / grid.Ny

			// min >> max
			grid.Lo1 = ps.lon0
			grid.Lo2 = ps.lon1

			// max >> min
			grid.La1 = ps.lat1
			grid.La2 = ps.lat0

			for k, arr := range grid.Data {
				grid.Data[k] = transT(arr, grid.Ny)
			}

			Vln(3, "[grid]", ps.lat, ps.lon, len(grid.Data["X"]), len(grid.Data["Y"]), len(grid.Data["海表溫度"]), len(grid.Data["海高"]), len(grid.Data["海表鹽度"]))
		}
	}
}

func transT(in []jsonFloat, stride int) []jsonFloat {
	sz := len(in)
	stride2 := sz / stride
	out := make([]jsonFloat, sz, sz)
	for i, v := range in {
		a := i / stride
		b := i % stride
		idx := a + b * stride2
		out[idx] = v
	}
	return out
}

type XmlState struct {
	Path []string
}

func NewXMLState() *XmlState {
	xs := &XmlState{}
	xs.Path = make([]string, 0, 32)
	return xs
}

func (xs *XmlState) StartTag(t xml.StartElement) {
	xs.Path = append(xs.Path, string(t.Name.Local))
}

func (xs *XmlState) EndTag(t xml.EndElement) {
	sz := len(xs.Path)
	xs.Path = xs.Path[:sz-1]
}

func (xs *XmlState) GetPath() string {
	return strings.Join(xs.Path, "/")
}

func (xs *XmlState) LastPath() string {
	sz := len(xs.Path) - 1
	if sz < 0 {
		return ""
	}
	return xs.Path[sz]
}

func (xs *XmlState) PathLevel() int {
	return len(xs.Path)
}


// ==== proxy ====
func makeConnection(targetAddr string, socksAddr string, timeout time.Duration) (net.Conn, error) {

	host, portStr, err := net.SplitHostPort(targetAddr)
	if err != nil {
		Vln(2, "SplitHostPort err:", targetAddr, err)
		return nil, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		Vln(2, "failed to parse port number:", portStr, err)
		return nil, err
	}
	if port < 1 || port > 0xffff {
		Vln(2, "port number out of range:", portStr, err)
		return nil, err
	}

	socksReq := []byte{0x05, 0x01, 0x00, 0x03}
	socksReq = append(socksReq, byte(len(host)))
	socksReq = append(socksReq, host...)
	socksReq = append(socksReq, byte(port>>8), byte(port))


	conn, err := net.DialTimeout("tcp", socksAddr, timeout)
	if err != nil {
		Vln(2, "connect to ", socksAddr, err)
		return nil, err
	}

	var b [10]byte

	// send request
	conn.Write([]byte{0x05, 0x01, 0x00})

	// read reply
	_, err = conn.Read(b[:2])
	if err != nil {
		return nil, err
	}

	// send server addr
	conn.Write(socksReq)

	// read reply
	n, err := conn.Read(b[:10])
	if n < 10 {
		Vln(2, "Dial err replay:", targetAddr, "via", socksAddr, n)
		return nil, err
	}
	if err != nil || b[1] != 0x00 {
		Vln(2, "Dial err:", targetAddr, "via", socksAddr, n, b[1], err)
		return nil, err
	}

	return conn, nil
}

// ==== log ====
func Vf(level int, format string, v ...interface{}) {
	if level <= *verbosity {
		log.Printf(format, v...)
	}
}
func Vln(level int, v ...interface{}) {
	if level <= *verbosity {
		log.Println(v...)
	}
}

