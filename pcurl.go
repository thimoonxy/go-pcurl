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
	"time"
)

var Clientvar *http.Client

func ckerr(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func getres(client *http.Client, url string, rgstart, rgend int64) (res *http.Response) {

	// Init Request
	req, err := http.NewRequest("GET", url, nil)
	ckerr(err)

	// Headers
	req.Proto = "HTTP/1.1"
	req.Header.Add("Accept",
		"*/*")
	req.Header.Add("User-Agent",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36")
	req.Header.Del("Accept-Encoding")

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

/* 	Byte Ranges
Reference to https://tools.ietf.org/html/rfc7233#section-2.1
The first-byte-pos value in a byte-range-spec gives the byte-offset
of the first byte in a range.  The last-byte-pos value gives the
byte-offset of the last byte in the range; that is, the byte
positions specified are inclusive.  Byte offsets start at zero.

Examples of byte-ranges-specifier values:

o  The first 500 bytes (byte offsets 0-499, inclusive):
		bytes=0-499
o  The second 500 bytes (byte offsets 500-999, inclusive):
		bytes=500-999 */
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
		task.length = originlength/count + 1
		task.rgend = task.rgstart + task.length - 1
	}

	tmp, err := ioutil.TempFile(tmpdirname, strconv.FormatInt(task.seq, 10)+".")
	ckerr(err)
	task.tmpfname = tmp.Name()
}

func (task *subtask) get(client *http.Client) {
	// create bar with appropriate decorators

	res := getres(client, task.url, task.rgstart, task.rgend)

	// read from proxy reader

	outputs, err := ioutil.ReadAll(res.Body)
	// ckerr(err)

	// write to tmp file
	err = ioutil.WriteFile(task.tmpfname, outputs, 0600)
	ckerr(err)
	task.tmpcreated = true

	// io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()
}

func reassemble(tasks []subtask, dst string) (done bool) {
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer f.Close()
	ckerr(err)
	done = false
	for i := 0; i < len(tasks); i++ {
		if tasks[i].tmpcreated {
			// appending to destination fpath
			tmp, err := os.Open(tasks[i].tmpfname)
			r := bufio.NewReader(tmp)
			n, err := r.WriteTo(f)
			ckerr(err)
			if n == 0 {
				return
			}

			// verifying tmpfile length vs range len in this task
			s, err := tmp.Stat()
			ckerr(err)
			if s.Size() != tasks[i].length {
				tmp.Close()
				log.Fatal(fmt.Sprintf("TmpFile %s size %d != range length %d", tasks[i].tmpfname, s.Size(), tasks[i].length))
			}

			// cleanning up
			tmp.Close()
			err = os.Remove(tasks[i].tmpfname)
			ckerr(err)
			log.Printf("Cleaned tmpfile: %s", tasks[i].tmpfname)
		}
	}
	done = true
	return
}

func b2s(data int64, n int) string {
	var byteUnits = []string{"B", "KB", "MB", "GB", "TB", "PB"}
	if data < 1<<10 {
		return fmt.Sprintf("%d %s", data, byteUnits[n])
	}
	return b2s(data/1024, n+1)
}

func precount(originlength, bufsize int64) (onetime, tmpfcount int) {
	if originlength >= bufsize {
		tmpfcount = int(originlength / bufsize)
		if tmpfcount >= 400 { // in case too many openfiles
			tmpfcount = 200
		} else if tmpfcount <= 4 {
			tmpfcount = 4
		}
		log.Printf("TmpFile amount: %d", tmpfcount)
		onetime = 4
	} else {
		tmpfcount = 4
		onetime = tmpfcount
	}
	return
}

func acceptRange(res *http.Response, dst string) {
	accept := res.Header.Get("Accept-Ranges")
	if strings.TrimSpace(accept) == "" {
		log.Println("Range not accepted, downloading directly ...")
		f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		defer f.Close()
		ckerr(err)
		r := bufio.NewReader(res.Body)
		n, err := r.WriteTo(f)
		ckerr(err)
		if res.ContentLength != n {
			log.Fatal(fmt.Sprintf("ContentLength %d != file size %d", res.ContentLength, n))
		}
		log.Println("Finshied:", dst)
		os.Exit(0)
	}
}

func main() {
	// Constants
	tmpbase := "/tmp"
	f, _ := os.Stat(tmpbase)
	if !f.IsDir() {
		tmpbase = os.TempDir()
	}
	tmpprefix := "gotemp"
	bufsize := int64(10 << 20) // 10MB
	MaxIdleConnections := 20
	RequestTimeout := 0 // no timeout

	// Preparing args
	if len(os.Args) != 3 {
		fmt.Println("Usage: $0 $src $dst")
		os.Exit(1)
	}
	url := os.Args[1] // url = "http://127.0.0.1:8080/centos.iso" // tmp testing
	dst := os.Args[2] // destination fpath, where file stores
	_, err := os.Stat(dst)
	if err == nil {
		fmt.Printf("%s found, Override?: [y/n]", dst)
		var decide string
		fmt.Scan(&decide)
		decide = strings.ToLower(decide)
		if strings.TrimSpace(decide) == "y" {
			os.Remove(dst)
		} else {
			log.Fatalf("Expected [y/n], but got %v", decide)
		}
	}

	// Preparing  Vars
	var b2sn int            // init n=0, used for b2s()
	var count int           // tmp file count
	var onetime int         // how many goroutines run at a time
	lock := make(chan bool) // main-routine waits till goroutines finish
	Clientvar = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: MaxIdleConnections,
			DisableCompression:  true, // client will compress it by default
		},
		Timeout: time.Duration(RequestTimeout) * time.Second,
	}
	res := getres(Clientvar, url, -1, -1) // -1 indicates no range specified
	acceptRange(res, dst)
	originlength := res.ContentLength
	onetime, count = precount(originlength, bufsize)
	tmpcreatedstat := make(chan bool, onetime) // ch with buff, to control batch scale
	tasks := make([]subtask, count)
	res.Body.Close()

	// Creating tmp dir
	tmpdirname := createTMPdir(tmpbase, tmpprefix)
	log.Printf("Originlength: %v", b2s(originlength, b2sn))
	log.Printf("Created tmpdir: %s", tmpdirname)
	defer destroyTMPdir(tmpdirname, tmpprefix) // cleanning up after process

	// Multi-processing
	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := 0; i < int(count); i++ {
		go func(i int, tmpcreatedstat chan bool) {
			tasks[i].seq = int64(i)
			tasks[i].init(originlength, int64(count), url, tmpdirname)
			tmpcreatedstat <- tasks[i].tmpcreated // write to ch before tasks[i].get() starts to 'buffer' the routines
			// log.Printf("Started getting tmpfile: %s", tasks[i].tmpfname)
			tasks[i].get(Clientvar)
			// fmt.Printf("seq: %v , start: %v, end: %v, lenght: %v, islast: %t \n", tasks[i].seq, tasks[i].rgstart, tasks[i].rgend, tasks[i].length, tasks[i].islast)
			if tasks[i].tmpcreated {
				// log.Printf("Got tmpfile: %s", tasks[i].tmpfname)
			}
			lock <- true
		}(i, tmpcreatedstat)
	}

	// Sticking goroutines onto main
	for i := 0; i < int(count); i++ {
		<-lock
		<-tmpcreatedstat
	}

	// Outputting

	if reassemble(tasks, dst) {
		log.Printf("Downloaded: from %s to %s", url, dst)
	}

}
