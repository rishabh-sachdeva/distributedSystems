package main

import(
	"fmt"
	zmq "github.com/pebbe/zmq4"
    "time"
    "strconv"
    "encoding/json"
	"math"
)
//global variables
const num_nodes_order = 6 // actual nodes 2^(num_nodes_order-1) = 32
var node_addresses []string
var nodes_in_ring []string
type Node struct{
	Key int
	Succ int //key of Succ
	Pre int //key of Pre
	FingerTable [num_nodes_order - 1]int//map[int]int
	Address string
	InRing bool
	bucket map[int]int
}

type DataStruct struct{
	Key string
	Value string
}

type Command struct{
	Do string
	SponsoringNode string
	Mode string
	ReplyTo string
	Data DataStruct	
	SenderNode *Node
}

// type MessageAcrossWorkers struct{
// 	FingerTable [num_nodes_order - 1]int
// }


	func worker(node *Node){
	    key := node.Key
		// succ := node.Succ
		// pre := node.Pre
		 finger_table := node.FingerTable
		 my_add := node.Address
		 in_ring := node.InRing
		//do socket stuff - pub and sub
		//fmt.Println(key,succ,pre,finger_table,my_add,in_ring)

		context,_ := zmq.NewContext()

		workerServer,_ := context.NewSocket(zmq.REP)
		workerServer.Bind(my_add)

		// workerServerAcross,_ := context.NewSocket(zmq.REP)
		// workerServerAcross.Bind(my_add)
		// go ListenToWorkers(workerServerAcross,my_add)

		for {
			recv_msg, err := workerServer.RecvBytes(0)
			if err != nil {
				fmt.Println(err)
				continue
			}
			unMarshalledCommand := Command{}
			json.Unmarshal(recv_msg, &unMarshalledCommand)
			workerServer.Send("acknowledged",0)
			fmt.Println(unMarshalledCommand,my_add)
			if unMarshalledCommand.Do == "join-ring"{
				in_ring=true
				sponsoringNode := unMarshalledCommand.SponsoringNode
				context,_ := zmq.NewContext()
				workerClient,_ := context.NewSocket(zmq.REQ) // worker client
				workerClient.Connect(sponsoringNode)
				fmt.Println("attempt to join",sponsoringNode)
				//how to do -- update buckets, pre and succ
				to_join := &Command{
					Do: "joining",
					SenderNode: node,
				}
				marshalled_joining, _ := json.Marshal(to_join) //message packing into json

				workerClient.SendBytes(marshalled_joining,0)

				workerClient.Recv(0)//get acknowledgement
				fmt.Println(in_ring,finger_table,key)
				nodes_in_ring = append(nodes_in_ring,my_add)

			}else if unMarshalledCommand.Do == "get-ring-fingers"{
				//return finger table
				
			}else if unMarshalledCommand.Do == "joining"{
				in_node := unMarshalledCommand.SenderNode
				fmt.Println(my_add,"joining ring",in_node)
				fmt.Println("joining ring with fingers	",in_node.FingerTable)

			}
		}
	}

	func main(){
		//publish coordinator thread for sending commands
		nodes_in_ring = append(nodes_in_ring,"tcp://127.0.0.1:5501")
		node_addresses = append(node_addresses,"tcp://127.0.0.1:5500")
		node_addresses = append(node_addresses,"tcp://127.0.0.1:5501")
		bucket_firstNode := make(map[int]int)
		bucket_firstNode[0]=1
		bucket_firstNode[1]=1
		bucket_firstNode[2]=1
		bucket_firstNode[3]=1
		bucket_firstNode[4]=1
		node := &Node{
			Key:1,
			InRing: true,
			Address:"tcp://127.0.0.1:5501",
			bucket:	bucket_firstNode,//make(map[string]string),

		}
		go worker(node)

		for i:=1;i<int(math.Pow(2, num_nodes_order-1));i++{
			port:=5501+i
			node_addresses = append(node_addresses,"tcp://127.0.0.1:"+strconv.Itoa(port))
			node := &Node{
				Key:i,
				InRing: false,
				Address:"tcp://127.0.0.1:"+strconv.Itoa(port),
				bucket:	make(map[int]int),
	
			}
			go worker(node)
		}
		context,_ := zmq.NewContext()
		coordinatorClient,_ := context.NewSocket(zmq.REQ) // coordinator client
		//joining
		coordinatorClient.Connect("tcp://127.0.0.1:5508")

		//joining command
		join_2 := &Command{
			Do: "join-ring",
			SponsoringNode : "tcp://127.0.0.1:5501",
		}
		marshalled_join_2, _ := json.Marshal(join_2) //message packing into json
		coordinatorClient.SendBytes(marshalled_join_2,0) // send msg to worker to join the ring
		coordinatorClient.Recv(0) 		// get acknowledgement
		time.Sleep(10*time.Second)
	}