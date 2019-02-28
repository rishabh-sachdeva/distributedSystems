package main

import (
	"fmt"
	"sync"
    "encoding/json"
    "os"
    "math"
    "strings"
    "strconv"
)
//var wg sync.WaitGroup 	

//JSON message struct
type Message struct{
	Start int
	End int
	Fname string
}

type Response struct{
	Psum int
	Pcount int
	Prefix string
	Suffix string
	Start int
	End int
}

var wg sync.WaitGroup

func coordinator(fileName string, M int){

	startPos := 0
	//fragmentSize := 3
	file, err := os.Open("F:/OperatingSystems/workspace/distributedSystems/src/computeAvg/data.txt")

         if err != nil {
                 fmt.Println(err)
                 os.Exit(1)
         }
         defer file.Close()

    fileInfo, _ := file.Stat()

	var file_size int64 = fileInfo.Size()
	fmt.Println("file size",file_size)

	//dividing file into M chunks
	fragment_size := math.Floor(float64(file_size-1) / float64(M))
	//lastFragment := int64(file_size) - (int64(M)* int64(fragments))
	fmt.Println(fragment_size)
	
	//creating M workers
	for i:=0; i< M; i++ {
		wg.Add(1)
		msg := &Message{startPos,startPos+int(fragment_size),fileName}
		marshalledMsg, _ := json.Marshal(msg) // message packing
		//fmt.Println("msg",marshalledMsg) 
		channel := make(chan []byte,1)
      // posting request message in channel
     go func(){
		channel <- marshalledMsg
		close(channel)
     }()
     startPos = startPos + int(fragment_size)+1;

	  //initiate job for a worker
	  go worker(fileName, channel)
	}
}

func worker(fileName string,channel chan []byte){
	//create a place "msg" where the decoded data will be stored
	defer wg.Done()
	 marshalledMsg := <-channel
	 unmarshalledMsg := Message{}
	 json.Unmarshal(marshalledMsg, &unmarshalledMsg)// unpacking request message
     response := calculateSum(unmarshalledMsg.Start,unmarshalledMsg.End,fileName)
     
     //pack or marshal the respose
     // sent res to coordinator via channel
	 fmt.Println("Psum",response)
}

func calculateSum(start int, end int, fileName string) *Response{
	sum:=0
	file, err := os.Open(fileName)
	checkIfAnyError(err)
    defer file.Close()
    _,e :=file.Seek(int64(start), 0) //check
	checkIfAnyError(e)
	
	fileContentForThisWorker := make([]byte, byte(end-start+1))
	_,e1 := file.Read(fileContentForThisWorker)
	
	checkIfAnyError(e1)
	chunk:=string(fileContentForThisWorker)
	fmt.Println("this chunk",chunk)
	
	nos := strings.Fields(chunk)
	fmt.Println(nos,len(nos))
	nums := []int{}
	
	prefix,suffix:=processChunkString(chunk)
	
	//initialize prefix, suffix and error variables
	//pre :=0
	//suf :=0
	//var ep,es error
	if len(prefix)>0{
		prefix = strings.TrimSpace(prefix)
		//fmt.Println("inside len", prefix)
		if len(prefix)<1{
			prefix=" "
		}
		//pre,ep =strconv.Atoi(prefix)
		//checkIfAnyError(ep)
	}
	if(len(suffix)>0){
		suffix=strings.TrimSpace(suffix)
		//fmt.Println("inside len s", suffix)

		if len(suffix)==0{
			suffix=" "
		}
		//suf,es = strconv.Atoi(suffix)
		//checkIfAnyError(es)
	}
	for _,i := range nos{
		i = strings.TrimSpace(i)
		j, err := strconv.Atoi(strings.Trim(i,"\x00")) // remove any null characters from string version of integer
		checkIfAnyError(err)
		nums = append(nums,j)
	}
	st:=1
	en:=len(nums)-1
	if prefix == " "{
		st =0
		fmt.Println("stta",st)
	}
	if suffix == " "{
		en=len(nums)
		fmt.Println("end",en)

	}
	for i:=st;i<en;i++{
		sum += nums[i]
	}
	res := &Response{Psum:sum, Pcount: len(nums)-2, Prefix:prefix , 
		Suffix: suffix,Start: start, End: end}
	//fmt.Println("pcount",res.Pcount)
	return res
}

func processChunkString(chunk string) (string,string){
	tokens := strings.Split(chunk, " ")
	prefix := tokens[0]
	suffix := tokens[len(tokens)-1]
	//nums := []string{}
	return prefix,suffix
	
}
func checkIfAnyError(err error){
	if err != nil {
                 fmt.Println(err)
                 os.Exit(1)
         }
}
func main(){
	fileName := "F:/OperatingSystems/workspace/distributedSystems/src/computeAvg/data.txt"
	//fileName := "data.txt"
	//a := []int{7, 2, 8, -9, 4, 0}
	coordinator(fileName,4)
    wg.Wait()
}
