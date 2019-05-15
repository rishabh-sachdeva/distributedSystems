	package main

	import (
		"encoding/json"
		"fmt"
		"math"
		"strconv"
		"time"
		"sync"

		zmq "github.com/pebbe/zmq4"
	)

	//global variables
	const num_nodes_order = 6 // actual nodes 2^(num_nodes_order-1) = 32
	var node_addresses map[int]string
	var nodes_in_ring []string

	type Node struct {
		Key         int
		Succ        int                      //key of Succ
		Pre         int                      //key of Pre
		FingerTable [num_nodes_order - 1]int //map[int]int
		Address     string
		InRing      bool
		bucket      map[int]int
	}

	type DataStruct struct {
		Key   string
		Value string
	}

	type NodeState struct {
		Predecessor string
		Successor   string
		FingerTable [num_nodes_order - 1]int
	}

	type Command struct {
		Do             string
		SponsoringNode string
		Mode           string
		ReplyTo        string
		Data           DataStruct
		SenderNode     *Node
		NodeState      *NodeState
		NewPredecessor int
		NewSuccessor   int
	}

	// type MessageAcrossWorkers struct{
	// 	FingerTable [num_nodes_order - 1]int
	// }

	func worker(node *Node) {
		key := node.Key
		// succ := node.Succ
		// pre := node.Pre
		// finger_table := node.FingerTable
		my_add := node.Address
		// in_ring := node.InRing
		// //do socket stuff - pub and sub
		// fmt.Println(key, succ, pre, finger_table, my_add, in_ring)

		context, _ := zmq.NewContext()

		workerServer, _ := context.NewSocket(zmq.REP)
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
			// workerServer.Send("acknowledged", 0)
			if unMarshalledCommand.Do == "join-ring" {
				workerServer.Send("acknowledged", 0)
				sponsoringNode := unMarshalledCommand.SponsoringNode
				context, _ := zmq.NewContext()
				workerClient, _ := context.NewSocket(zmq.REQ) // worker client
				sponsoringNodeSucc := 0
				lastSucc := 1

				temp:=0
				for sponsoringNodeSucc < key && sponsoringNodeSucc != 1 {
					temp++
					if sponsoringNodeSucc != 0 {
						lastSucc = sponsoringNodeSucc
					}
					// lastSucc = sponsoringNodeSucc
					workerClient.Connect(sponsoringNode)
					fmt.Println("attempting to join", node.Key, "from", sponsoringNode)
					//how to do -- update buckets, pre and succ
					findSucc := &Command{
						Do:      "find-ring-successor",
						ReplyTo: my_add,
					}
					marshalled_joining, _ := json.Marshal(findSucc) //message packing into json
					workerClient.SendBytes(marshalled_joining, 0)

					recvSucc, _ := workerClient.Recv(0)
					fmt.Println("recvSucc -", recvSucc)
					sponsoringNodeSucc, _ = strconv.Atoi(recvSucc) //get acknowledgement
					sponsoringNode = node_addresses[sponsoringNodeSucc]
				
					node.Succ = sponsoringNodeSucc
					node.Pre = lastSucc
					node.InRing = true
			}
			if temp==1{
				//find predecessor for the current node
				findPre := &Command{
					Do:      "find-ring-predecessor",
					ReplyTo: my_add,
				}
				fmt.Println("Connecting to find pre",node_addresses[node.Succ])
				workerClient.Connect(node_addresses[node.Succ])
				marshalled_Precmd, _ := json.Marshal(findPre) //message packing into json
				workerClient.SendBytes(marshalled_Precmd, 0)
				recvPre, _ := workerClient.Recv(0)
				fmt.Println("recvPre -", recvPre)
				lastSucc, _ = strconv.Atoi(recvPre)
				node.Pre=lastSucc

			}
			// findPre := &Command{
			// 	Do:      "find-ring-predecessor",
			// 	ReplyTo: my_add,
			// }
			// fmt.Println("Connecting to find pre",node_addresses[node.Succ])
			// workerClient.Connect(node_addresses[node.Succ])
			// marshalled_Precmd, _ := json.Marshal(findPre) //message packing into json
			// workerClient.SendBytes(marshalled_Precmd, 0)
			// recvPre, _ := workerClient.Recv(0)
			// fmt.Println("recvPre -", recvPre)
			// lastSucc, _ = strconv.Atoi(recvPre)
			// node.Pre=lastSucc

			// node.Succ = sponsoringNodeSucc
			// node.Pre = lastSucc
			// node.InRing = true
			fmt.Println(">>>>Finalized Node Succ and Pre:",my_add, node.Succ, node.Pre)
			// updating concerned Pre and Succ
			updatePreCommand := &Command{
				Do : "update-predecessor",
				NewPredecessor : node.Key,
			}
			fmt.Println("executing updatePre for node key:",node.Succ)
			executeCommand(node_addresses[node.Succ],updatePreCommand)
			updateSuccCommand := &Command{
				Do : "update-successor",
				NewSuccessor : node.Key,
			}
			fmt.Println("executing updateSucc for node key:",node)
			executeCommand(node_addresses[node.Pre],updateSuccCommand)

			// fmt.Println(in_ring, finger_table, key, node.Succ, node.Pre)
			nodes_in_ring = append(nodes_in_ring, my_add)
			fmt.Println("nodes in ring -", nodes_in_ring)
			fmt.Println("Node details - ", node)
				
			} else if unMarshalledCommand.Do == "get-ring-fingers" {
				//return finger table
			} else if unMarshalledCommand.Do == "find-ring-successor" {
				// workerServer.Send(strconv.Itoa(succ), 0)
				workerServer.Send(strconv.Itoa(node.Succ), 0)
			} else if unMarshalledCommand.Do == "find-ring-predecessor" {
				 workerServer.Send(strconv.Itoa(node.Pre), 0)
			} else if unMarshalledCommand.Do == "update-predecessor"{
				node.Pre = unMarshalledCommand.NewPredecessor
				//send acknowledgement
				workerServer.Send("Updated Pre",0)
				fmt.Println("inside update pre")
				fmt.Println("Node details - Pre", node)

			} 	else if unMarshalledCommand.Do == "update-successor"{
				node.Succ = unMarshalledCommand.NewSuccessor
				workerServer.Send("Updated Succ",0)
				fmt.Println("Node details - Succ", node)

			}
		}
	}

	func main() {
		//publish coordinator thread for sending commands
		node_addresses = make(map[int]string)
		nodes_in_ring = append(nodes_in_ring, "tcp://127.0.0.1:5501")
		node_addresses[0] = "tcp://127.0.0.1:5500"
		node_addresses[1] = "tcp://127.0.0.1:5501"
		bucket_firstNode := make(map[int]int)
		bucket_firstNode[0] = 1
		bucket_firstNode[1] = 1
		bucket_firstNode[2] = 1
		bucket_firstNode[3] = 1
		bucket_firstNode[4] = 1
		node := &Node{
			Key:     1,
			InRing:  true,
			Address: "tcp://127.0.0.1:5501",
			bucket:  bucket_firstNode, //make(map[string]string),
			Pre:     1,
			Succ:    1,
		}
		go worker(node)

		for i := 1; i < int(math.Pow(2, num_nodes_order-1)); i++ {
			port := 5501 + i
			node_addresses[i+1] = "tcp://127.0.0.1:" + strconv.Itoa(port)
			node := &Node{
				Key:     i + 1,
				InRing:  false,
				Address: "tcp://127.0.0.1:" + strconv.Itoa(port),
				bucket:  make(map[int]int),
			}
			go worker(node)
		}

		//joining 8 command
		join_8 := &Command{
			Do:             "join-ring",
			SponsoringNode: "tcp://127.0.0.1:5501",
		}
		executeCommand("tcp://127.0.0.1:5508", join_8)

		time.Sleep(1 * time.Second)
		//joining node 14
		join_14 := &Command{
			Do:             "join-ring",
			SponsoringNode: "tcp://127.0.0.1:5501",
		}
		executeCommand("tcp://127.0.0.1:5514", join_14)

		time.Sleep(1 * time.Second)
		//joining node 9
		join_9 := &Command{
			Do:             "join-ring",
			SponsoringNode: "tcp://127.0.0.1:5508",
		}
		executeCommand("tcp://127.0.0.1:5509", join_9)

		// time.Sleep(1 * time.Second)
		// //joining node 12
		// join_12 := &Command{
		// 	Do:             "join-ring",
		// 	SponsoringNode: "tcp://127.0.0.1:5508",
		// }
		// executeCommand("tcp://127.0.0.1:5512", join_12)

		time.Sleep(1000 * time.Second)
	}

	// func executeCommand(client *zmq.Socket, command *Command) {
	// 	// Command marshalled
	// 	marshalledJson, _ := json.Marshal(command)
	// 	// Instructing Node to perform Command
	// 	client.SendBytes(marshalledJson, 0)
	// 	client.Recv(0) // get acknowledgement
	// }

	func executeCommand(address string, command *Command) {
		var mu sync.Mutex
		mu.Lock()
		context, _ := zmq.NewContext()
		coordinatorClient, _ := context.NewSocket(zmq.REQ)
		coordinatorClient.Connect(address)
		// Command marshalled
		marshalledJson, _ := json.Marshal(command)
		// Instructing Node to perform Command
		coordinatorClient.SendBytes(marshalledJson, 0)
		coordinatorClient.Recv(0) // get acknowledgement
		mu.Unlock()
	}
