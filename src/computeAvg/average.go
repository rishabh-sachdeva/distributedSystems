package main

/**
* Author: Rishabh Sachdeva and Arshita Jain
**/
import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)


//JSON message struct
type Message struct {
	Start int
	End   int
	Fname string
}

//JSON response struct
type Response struct {
	Psum   int64
	Pcount int
	Prefix string
	Suffix string
	Start  int
	End    int
}

var wg sync.WaitGroup
var wg_res sync.WaitGroup

func coordinator(fileName string, M int) float64{

	startPos := 0
	//fragmentSize := 3
	file, err := os.Open(fileName)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()

	var file_size int64 = fileInfo.Size()
	fmt.Println("file size", file_size)

	//dividing file into M chunks
	fragment_size := math.Floor(float64(file_size-1) / float64(M))
	responseChannel := make(chan []byte, M)
	wg_res.Add(M)
	//creating M workers
	for i := 0; i < M; i++ {
		wg.Add(1)
		endPos := 0
		if i == M-1 {
			endPos = int(file_size)
		} else {
			endPos = startPos + int(fragment_size)
		}
		msg := &Message{startPos, endPos, fileName}
		marshalledMsg, _ := json.Marshal(msg) // message packing
		channel := make(chan []byte, 1)
		// posting request message in channel
		go func() {
			channel <- marshalledMsg
		}()
		startPos = startPos + int(fragment_size) + 1
		//initiate job for a worker
		go worker(fileName, channel, responseChannel)
	}
	wg_res.Wait()
	responseArr := []Response{}

	responseArr = getResponseInArray(responseChannel, responseArr)
	//fmt.Println(responseArr)

	sum_ret, count_ret := getAvgOfReturned(responseArr)
	sum_rem, count_rem := getAvgRemaining(responseArr)
	sum_final := sum_ret + sum_rem
	count_final := count_ret + count_rem
	fmt.Println("Total Sum",sum_final)
	fmt.Println("Total count",count_final)
	average := float64(sum_final) / float64(count_final)
	fmt.Println("*******AVERAGE********",average)
	return average
}

func getAvgRemaining(responseArr []Response) (int64, int) {
	sum := int64(0)
	count := 0 // one for prefix of first response
	//sort array on the basis of "Start" byte
	sort.Slice(responseArr[:], func(i, j int) bool { return responseArr[i].Start < responseArr[j].Start })
	//fmt.Println("sorted array on the basis of Response.Start", responseArr)
	pre_suff := "" //responseArr[0].Suffix
	for i := 0; i < len(responseArr); i++ {
		if responseArr[i].Pcount == 0 && responseArr[i].Prefix == responseArr[i].Suffix {
			//single number in a fragment
			pre_suff += responseArr[i].Prefix
			//fmt.Println("pre_suff",pre_suff)
		} else {
			concat_num, err := strconv.ParseInt(strings.TrimSpace(strings.Trim(pre_suff+responseArr[i].Prefix, "\x00")), 10, 64)
			checkIfAnyError(err, "getAvgRemaining-inside for:part 2")
			//fmt.Println("concat", concat_num)
			sum += concat_num
			pre_suff = responseArr[i].Suffix
			count++
		}
	}
	last_num, err := strconv.ParseInt(strings.TrimSpace(strings.Trim(pre_suff, "\x00")), 10, 64)
	checkIfAnyError(err, "getAvgRemaining-end")
	return sum + last_num, count + 1

}

func getAvgOfReturned(responseArr []Response) (int64, int) {
	sum := int64(0)
	count := 0
	for _, res := range responseArr {
		sum += res.Psum
		count += res.Pcount
	}
	return sum, count
}
func getResponseInArray(responseChannel chan []byte, responseArr []Response) []Response {
	var flag_closed = false
	for elem := range responseChannel {
		//marshalledRes := <-responseChannel
		unmarshalledRes := Response{}
		json.Unmarshal(elem, &unmarshalledRes) // unpacking response
		responseArr = append(responseArr, unmarshalledRes)
		if !flag_closed {
			close(responseChannel)
			flag_closed = true
		}
	}
	return responseArr
}

func worker(fileName string, channel chan []byte, responseChannel chan []byte) {
	//create a place "msg" where the decoded data will be stored
	defer wg.Done()
	marshalledMsg := <-channel
	unmarshalledMsg := Message{}
	json.Unmarshal(marshalledMsg, &unmarshalledMsg) // unpacking request message
	response := calculateSum(unmarshalledMsg.Start, unmarshalledMsg.End, fileName)
	marshalledResponse, _ := json.Marshal(response) // response packing
	// sent res to coordinator via channel
	//responseChannel := make(chan []byte,1)
	go func() {
		responseChannel <- marshalledResponse
		defer wg_res.Done()
	}()
}

func calculateSum(start int, end int, fileName string) *Response {
	sum := int64(0)
	file, err := os.Open(fileName)
	checkIfAnyError(err, "calculateSum")
	defer file.Close()
	_, e := file.Seek(int64(start), 0) //check
	checkIfAnyError(e, "calculateSum")

	fileContentForThisWorker := make([]byte, byte(end-start+1))
	_, e1 := file.Read(fileContentForThisWorker)

	checkIfAnyError(e1, "calculateSum")
	chunk := string(fileContentForThisWorker)
	nos := strings.Fields(chunk)
	nums := []int64{}

	prefix, suffix := processChunkString(chunk)
	if len(prefix) > 0 {
		prefix = strings.TrimSpace(prefix)
		if len(prefix) == 0 {
			prefix = " "
		}
	}
	if len(suffix) > 0 {
		suffix = strings.TrimSpace(suffix)
		if len(suffix) == 0 {
			suffix = " "
		}
	}
	for _, i := range nos {
		i = strings.TrimSpace(i)
		j, err := strconv.ParseInt(strings.Trim(i, "\x00"), 10, 64) // remove any null characters from string version of integer
		checkIfAnyError(err, "calculateSum "+i)
		nums = append(nums, j)
	}
	count_deduct := 2
	st := 1
	en := len(nums) - 1
	if prefix == " " {
		st = 0
		count_deduct -= 1
	}
	if suffix == " " {
		en = len(nums)
		count_deduct -= 1
	}
	for i := st; i < en; i++ {
		sum += nums[i]
	}
	cnt := len(nums) - count_deduct
	if cnt < 0 {
		cnt = 0
	}
	res := &Response{Psum: sum, Pcount: cnt, Prefix: prefix,
		Suffix: suffix, Start: start, End: end}
	return res
}

func processChunkString(chunk string) (string, string) {
	prefix := ""
	suffix := ""
	if strings.HasPrefix(chunk, " ") {
		prefix = " "
	}
	if strings.HasSuffix(chunk, " ") {
		suffix = " "
	}
	tokens := strings.Split(chunk, " ")
	if prefix == "" {
		prefix = tokens[0]
	}
	if suffix == "" {
		suffix = tokens[len(tokens)-1]
	}
	//nums := []string{}
	return prefix, suffix
}

func checkIfAnyError(err error, func_name string) {
	if err != nil {
		fmt.Println(func_name+":", err)
		os.Exit(1)
	}
}
func main() {
	fileNamePtr := flag.String("fName", "F:/OperatingSystems/workspace/distributedSystems/src/computeAvg/data.txt", "Path to the file")
	MPtr := flag.Int("M",3, "number of fragments")
	flag.Parse()

	fmt.Println(*fileNamePtr, *MPtr)
	coordinator(*fileNamePtr, *MPtr)
	wg.Wait()
}
