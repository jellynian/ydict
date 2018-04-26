package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

const (
	appKey     = "6316ef77f6cbeed3"
	privateKey = "eLChc3YZNUaJz8UT23e9EmcYKlbMiWsN"
)

type JTSData struct {
	Web         []Web    `json:"web"`
	Translation []string `json:"translation"`
	ErrorCode   string   `json:"errorCode"`
	Webdict     Webdict  `json:"webdict"`
	SpeakURL    string   `json:"speakUrl"`
	TSpeakURL   string   `json:"tSpeakUrl"`
	Query       string   `json:"query"`
	Dict        Dict     `json:"dict"`
	Basic       Basic    `json:"basic"`
	L           string   `json:"l"`
}

type Webdict struct {
	URL string `json:"url"`
}

type Dict struct {
	URL string `json:"url"`
}

type Web struct {
	Value []string `json:"value"`
	Key   string   `json:"key"`
}

type Basic struct {
	UsPhonetic string   `json:"us-phonetic"`
	Phonetic   string   `json:"phonetic"`
	UkPhonetic string   `json:"uk-phonetic"`
	UkSpeech   string   `json:"uk-speech"`
	Explains   []string `json:"explains"`
	UsSpeech   string   `json:"us-speech"`
}

func main() {

	q := os.Args[1]

	salt := strconv.Itoa(int(time.Now().UnixNano()))
	signStr := appKey + q + salt + privateKey
	h := md5.New()
	h.Write([]byte(signStr))
	sign := h.Sum(nil)
	queryString := fmt.Sprintf("?q=%s&from=%s&to=%s&appKey=%s&salt=%s&sign=%x&ext=%s", q, "auto", "auto", appKey, salt, sign, "mp3")

	exit := make(chan error)
	myresp := make(chan JTSData)

	go func() {
		resp, err := http.Get("https://openapi.youdao.com/api" + queryString)
		if err != nil {
			exit <- err
			runtime.Goexit()
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			exit <- err
			runtime.Goexit()
		}

		result := new(JTSData)

		err = json.Unmarshal(data, result)
		if err != nil {
			exit <- err
			runtime.Goexit()
		}

		myresp <- *result

	}()

	var result JTSData

	flag := false

	fmt.Printf("\033[0;32m")

	for {
		select {
		case result = <-myresp:
			flag = true
			break
		case err := <-exit:
			log.Fatal(err.Error())
			os.Exit(1)
		case <-time.After(time.Second * 3):
			log.Fatal("time out")
		default:
			for i := 0; i < 100; i++ {
				fmt.Printf("\r%s", bar(i))
			}
		}
		if flag {
			fmt.Printf("\r")
			break
		}
	}

	re, err := strconv.Atoi(result.ErrorCode)

	if err != nil {
		log.Fatal(err.Error())
	}

	if re != 0 {
		os.Exit(1)
	}

	fmt.Println("  \033[0;32m美:", "[", result.Basic.Phonetic, "]", "英:", "[", result.Basic.UkPhonetic, "]")

	fmt.Printf("\n")

	for _, v := range result.Basic.Explains {
		fmt.Println("  ", v)
	}
	fmt.Printf("\033[0m\n")

	for k, v := range result.Web {
		fmt.Println("\033[0;32m ", k+1, ".", v.Key)
		out := ""
		for _, b := range v.Value {
			out += b
			out += "  "
		}
		fmt.Println("\033[0;31m ", out)
	}
}

func bar(vl int) string {
	global := []rune("▒▒▒▒▒▒▒▒▒▒")
	for i := 0; i < vl/10; i++ {
		global[i] = '█'
	}
	return string(global)
}
