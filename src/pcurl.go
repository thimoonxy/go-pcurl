package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func ckerr(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func getres(url string, rgstart, rgend int64) (res *http.Response) {

	// Init Request
	tr := &http.Transport{
		DisableCompression: true, // client will compress it by default
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	ckerr(err)
	// dump, _ := httputil.DumpRequest(req, true)
	// fmt.Printf("Request:\n%s\n", string(dump))

	// Headers
	req.Proto = "HTTP/1.1"
	req.Header.Add("Accept",
		"*/*")
	req.Header.Add("User-Agent",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36")
	req.Header.Del("Accept-Encoding")
	//	req.Header.Add("Accept",
	//		"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	//	req.Header.Add("Accept-Encoding",
	//	"gzip, deflate, sdch, br")

	// Range
	if rgstart >= 0 && rgend > 0 {
		req.Header.Add("Range",
			fmt.Sprintf("bytes=%d-%d", rgstart, rgend))
	}

	// Response
	res, err = client.Do(req)
	// dump, _ := httputil.DumpResponse(res, false)
	// fmt.Printf("Response:\n%s\n", string(dump))
	ckerr(err)
	return res
}

func createTMPdir(base, prefix string) string {
	name, err := ioutil.TempDir(base, prefix)
	ckerr(err)
	return name
}

func destroyTMPdir(dir, prefix string) {
	if strings.Contains(dir, prefix) != true {
		log.Fatal(dir + " Dir is not a temp dir.")
	}
	stat, err := os.Stat(dir)
	ckerr(err)
	if stat.IsDir() {
		err = os.RemoveAll(dir)
		ckerr(err)
		log.Printf("Removed tmpdir: %s", dir)
	}
}

type subtask struct {
	seq        int64
	rgstart    int64
	rgend      int64
	length     int64
	islast     bool
	url        string
	tmpfname   string
	tmpcreated bool
}

func (task *subtask) init(originlength, count int64, url, tmpdirname string) {
	task.url = url
	task.tmpcreated = false
	task.rgstart = task.seq*(originlength/count) + task.seq&task.seq
	if task.seq == count-1 {
		task.islast = true
		// task.rgstart = task.seq*(originlength/count) + originlength%count + task.seq&task.seq
		// task.length = originlength - originlength/count*count + originlength/count
		task.rgend = originlength
		task.length = task.rgend - task.rgstart
	} else {
		task.islast = false
		task.length = originlength / count
		task.rgend = task.rgstart + task.length
	}

	tmp, err := ioutil.TempFile(tmpdirname, strconv.FormatInt(task.seq, 10)+".")
	ckerr(err)
	task.tmpfname = tmp.Name()
}

func (task *subtask) get() {
	res := getres(task.url, task.rgstart, task.rgend)
	outputs, err := ioutil.ReadAll(res.Body)
	// ckerr(err)
	err = ioutil.WriteFile(task.tmpfname, outputs, 0644)
	ckerr(err)
	task.tmpcreated = true
	res.Body.Close()
}

func reassemble(tasks []subtask, dst string) (done bool) {
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer f.Close()
	ckerr(err)
	done = false
	for i := 0; i < len(tasks); i++ {
		if tasks[i].tmpcreated {
			tmp, err := os.Open(tasks[i].tmpfname)
			r := bufio.NewReader(tmp)
			n, err := r.WriteTo(f)
			ckerr(err)
			if n == 0 {
				return
			}
			tmp.Close()
			err = os.Remove(tasks[i].tmpfname)
			ckerr(err)
			log.Printf("Cleaned tmpfile: %s", tasks[i].tmpfname)
		}
	}
	done = true
	return
}

func main() {

	// Preparing  Vars
	if len(os.Args) != 3 {
		fmt.Println("Usage: $0 $src $dst")
		os.Exit(1)
	}
	var count int
	url := os.Args[1]
	dst := os.Args[2] // destination fpath, where file stores
	tmpbase := "/tmp/"
	tmpprefix := "gotemp"
	// url = "http://127.0.0.1:8080/centos.iso" // tmp testing
	lock := make(chan bool)
	res := getres(url, -1, -1) // -1 indicates no range specified
	originlength := res.ContentLength
	if originlength >= 10<<20 {
		count = 20
	} else {
		count = 4
	}
	tmpcreatedstat := make(chan bool, count)
	tasks := make([]subtask, count)

	// fmt.Println(originlength)
	res.Body.Close()
	tmpdirname := createTMPdir(tmpbase, tmpprefix)
	log.Printf("Created tmpdir: %s", tmpdirname)
	defer destroyTMPdir(tmpdirname, tmpprefix) // cleanning up after process

	// Processing
	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := 0; i < int(count); i++ {
		go func(i int) {
			tasks[i].seq = int64(i)
			tasks[i].init(originlength, int64(count), url, tmpdirname)
			tasks[i].get()
			// fmt.Printf("seq: %v , start: %v, end: %v, lenght: %v, islast: %t \n", tasks[i].seq, tasks[i].rgstart, tasks[i].rgend, tasks[i].length, tasks[i].islast)
			if tasks[i].tmpcreated {
				log.Printf("Created tmpfile: %s", tasks[i].tmpfname)
				tmpcreatedstat <- tasks[i].tmpcreated
			}
			lock <- true
		}(i)
	}
	// goroutine block
	for i := 0; i < int(count); i++ {
		<-tmpcreatedstat
		<-lock
	}
	// Outputting
	// fmt.Println(originlength)
	// fmt.Println(dst)
	// fmt.Printf("%#v\n", tasks)
	// []main.subtask{main.subtask{length:11131263, seq:0, pos:0, islast:false, tmpfname:"/tmp/gotemp594916354/0.a.txt.037813625", tmpcreated:true, tmpcleaned:false}, main.subtask{length:11131263, seq:1, pos:11131263, islast:false, tmpfname:"/tmp/gotemp594916354/1.a.txt.035369604", tmpcreated:true, tmpcleaned:false}, main.subtask{length:11131263, seq:2, pos:22262526, islast:false, tmpfname:"/tmp/gotemp594916354/2.a.txt.400843283", tmpcreated:true, tmpcleaned:false}, main.subtask{length:11131264, seq:3, pos:33393789, islast:true, tmpfname:"/tmp/gotemp594916354/3.a.txt.100040790", tmpcreated:true, tmpcleaned:false}}
	if reassemble(tasks, dst) {
		log.Printf("Downloaded: from %s to %s", url, dst)
	}

}
