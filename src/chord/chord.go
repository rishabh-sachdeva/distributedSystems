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
				Bucket      map[int]int
			}

			type DataStruct struct {
				Key   int
				Value int
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
				Data           *DataStruct
				SenderNode     *Node
				NodeState      *NodeState
				NewPredecessor int
				NewSuccessor   int
			}
			type FingerTable struct{
				FingerTable [num_nodes_order-1]int
			}
			type BucketStruct struct{
				Bucket map[int]int
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
							//fmt.Println("attempting to join", node.Key, "from", sponsoringNode)
							//how to do -- update buckets, pre and succ
							findSucc := &Command{
								Do:      "find-ring-successor",
								ReplyTo: my_add,
							}
							marshalled_joining, _ := json.Marshal(findSucc) //message packing into json
							workerClient.SendBytes(marshalled_joining, 0)

							recvSucc, _ := workerClient.Recv(0)
							//fmt.Println("recvSucc -", recvSucc)
							sponsoringNodeSucc, _ = strconv.Atoi(recvSucc) 
							sponsoringNode = node_addresses[sponsoringNodeSucc]
						
							node.Succ = sponsoringNodeSucc
							node.Pre = lastSucc
							node.InRing = true
					}
					if temp==1{
						//issue - pre not updated
						//find predecessor for the current node
						workerClient.Connect(node_addresses[node.Succ])

						findPre := &Command{
							Do:      "find-ring-predecessor",
							ReplyTo: my_add,
						}
						//fmt.Println("****************Connecting to find pre",node_addresses[node.Succ])
						marshalled_Precmd, _ := json.Marshal(findPre) //message packing into json
						workerClient.SendBytes(marshalled_Precmd, 0)
						recvPre, _ := workerClient.Recv(0) // receiving predecessor details
						fmt.Println("recvPre -", recvPre)
						lastSucc, _ = strconv.Atoi(recvPre)
						node.Pre=lastSucc

					}
				//	fmt.Println(">>>>Finalized Node Succ and Pre:",my_add, node.Succ, node.Pre)
					// updating concerned Pre and Succ
					updatePreCommand := &Command{
						Do : "update-predecessor",
						NewPredecessor : node.Key,
					}
				//	fmt.Println("executing updatePre for node key:",node.Succ)
					executeCommand(node_addresses[node.Succ],updatePreCommand)
					updateSuccCommand := &Command{
						Do : "update-successor",
						NewSuccessor : node.Key,
					}
				//	fmt.Println("executing updateSucc for node key:",node)
					executeCommand(node_addresses[node.Pre],updateSuccCommand)

					// fmt.Println(in_ring, finger_table, key, node.Succ, node.Pre)
					nodes_in_ring = append(nodes_in_ring, my_add)
					fmt.Println("nodes in ring -", nodes_in_ring)
					fmt.Println("Node details - ", node)
						
					} else if unMarshalledCommand.Do == "find-ring-successor" {
						// workerServer.Send(strconv.Itoa(succ), 0)
						workerServer.Send(strconv.Itoa(node.Succ), 0)
					} else if unMarshalledCommand.Do == "find-ring-predecessor" {
						workerServer.Send(strconv.Itoa(node.Pre), 0)
					} else if unMarshalledCommand.Do == "update-predecessor"{
						node.Pre = unMarshalledCommand.NewPredecessor
						//send acknowledgement
						workerServer.Send("Updated Pre",0)
					//	fmt.Println("inside update pre")
					//	fmt.Println("Node details - Pre", node)

					} 	else if unMarshalledCommand.Do == "update-successor"{
						node.Succ = unMarshalledCommand.NewSuccessor
						workerServer.Send("Updated Succ",0) //send ack
					//	fmt.Println("Node details - Succ", node)
					}else if unMarshalledCommand.Do == "fix-ring-fingers" {
							workerServer.Send("acknowledged", 0)
							
							node.FingerTable[0] = node.Succ				
							current_succ:=node.Succ	
							currentAdd := node_addresses[node.Succ]
							for i:=1;i<num_nodes_order-1;i++{
								calculated_ele := node.Key + int(math.Pow(2,float64(i)))
								//fmt.Println("calculated_ele",calculated_ele)

								calculated_ele_mod := calculated_ele % int(math.Pow(2,num_nodes_order-1))//get mod of 32
								//address_calculated_ele := node_addresses[calculated_ele]
								//check if its in ring
								//	presentInRing := ifNodeAddressInRing(address_calculated_ele) 
								if current_succ == 1 && calculated_ele_mod == calculated_ele{
									node.FingerTable[i] = 1
								}else if calculated_ele <= current_succ{
									node.FingerTable[i]=current_succ

								}else{
									for {
										contextForFinger, _ := zmq.NewContext()
										workerClientForFinger, _ := contextForFinger.NewSocket(zmq.REQ) // client

										workerClientForFinger.Connect(currentAdd)
										findSucc := &Command{
												Do:      "find-ring-successor",
												ReplyTo: my_add,
										}
										marshalled_findSucc, _ := json.Marshal(findSucc) //message packing into json
										//fmt.Println("finding succ of",currentAdd)
										workerClientForFinger.SendBytes(marshalled_findSucc, 0)
										recvSucc, _ := workerClientForFinger.Recv(0) //get succ in form of acknowledgement
									//	fmt.Println("receiving succc of",currentAdd,"succ is",recvSucc)

										//fmt.Println("got succ ",recvSucc)

										current_succ, _ = strconv.Atoi(recvSucc) 
										currentAdd = node_addresses[current_succ]
										if current_succ>=calculated_ele || current_succ==1{
											node.FingerTable[i]=current_succ
											break
										}
									}
								}
							}
						//	fmt.Println("Finger table of node with address :",node.Address,node.FingerTable)

					}else if unMarshalledCommand.Do == "Put"{
					//	fmt.Println("inside PUT")
						workerServer.Send("acknowledged",0)
						data_key := unMarshalledCommand.Data.Key
						//data_val := unMarshalledCommand.Data.Value
						finger_table := node.FingerTable
						var current_node_key int
						//fmt.Println("Finger table of address:",my_add, finger_table, data_key)
						var concerned_node_address string
						for {
							current_node_key=finger_table[0]
							if finger_table[0] >= data_key{
								concerned_node_address = node_addresses[finger_table[0]]
								break	
							}
							for _,fingerKey := range finger_table{
								if fingerKey <= data_key && current_node_key<=fingerKey{
									current_node_key=fingerKey
								}else{
									break
								}
							}
							//	fmt.Println("current key",current_node_key)
							concerned_node_address = node_addresses[current_node_key]

							contextForDataPut, _ := zmq.NewContext()
							workerClientForDataPut, _ := contextForDataPut.NewSocket(zmq.REQ) // client
							workerClientForDataPut.Connect(concerned_node_address)

							getFingerCmd := &Command{
								Do: "get-ring-fingers",
								ReplyTo: my_add,
							}
							marshal_get_finger, _ := json.Marshal(getFingerCmd)
							workerClientForDataPut.SendBytes(marshal_get_finger,0)
							finger_table_json_ret, _ := workerClientForDataPut.RecvBytes(0)
							//fmt.Println("finger table returned :",ack)
							unMarshalledFingerTable := FingerTable{}
							json.Unmarshal(finger_table_json_ret,&unMarshalledFingerTable)
							fmt.Println(unMarshalledFingerTable.FingerTable)
							if unMarshalledFingerTable.FingerTable[0]>= data_key{
								//update in current node's successor 
								//break the loop here
								if unMarshalledFingerTable.FingerTable[0]>current_node_key{ 
									concerned_node_address = node_addresses[unMarshalledFingerTable.FingerTable[0]]
							}
								break
							}
						}

						contextForAdddata, _ := zmq.NewContext()
						workerClientAdddata, _ := contextForAdddata.NewSocket(zmq.REQ) // client
						workerClientAdddata.Connect(concerned_node_address)
						addData := &Command{
								Do:      "update-bucket",
								ReplyTo: my_add,
								Data: unMarshalledCommand.Data,
						}
						marshalled_addData, _ := json.Marshal(addData) //message packing into json
						//fmt.Println("send update bucket command of address:",concerned_node_address)
						workerClientAdddata.SendBytes(marshalled_addData, 0)
						workerClientAdddata.Recv(0)//get ack
						//fmt.Println(recvAck)
					}else if unMarshalledCommand.Do == "update-bucket"{
						fmt.Println("updating bucket in ",my_add, unMarshalledCommand.Data)
						node.Bucket[unMarshalledCommand.Data.Key] = unMarshalledCommand.Data.Value
						workerServer.Send("data updated in bucket "+my_add,0)
					}else if unMarshalledCommand.Do == "get-ring-fingers"{
						fingerTableJson := &FingerTable{node.FingerTable}
						marshalledJsonFingers,_ := json.Marshal(fingerTableJson)
						workerServer.SendBytes(marshalledJsonFingers,0)

					}else if unMarshalledCommand.Do =="Get"{

						data_key := unMarshalledCommand.Data.Key
						//fmt.Println("inside get", node.Bucket[data_key])
			//			finger_table := node.FingerTable
						workerServer.Send(strconv.Itoa(node.Bucket[data_key]),0)

					}else if unMarshalledCommand.Do =="init-ring-fingers"{

						node.FingerTable = [num_nodes_order-1]int{}
						workerServer.Send("init the fingers",0)

					}else if unMarshalledCommand.Do =="list-items"{
						bucketJson := BucketStruct{Bucket: node.Bucket}
						marshalledBucketJson,_ := json.Marshal(bucketJson)
						fmt.Println("Bucket:", node.Bucket)  
						workerServer.Send(string(marshalledBucketJson),0)
					}
				}
				
			}
			func ifNodeAddressInRing(address string) bool{
				for _,item := range nodes_in_ring{
					if item==address{
						return true
					}
				}
				return false
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
					Bucket:  bucket_firstNode, //make(map[string]string),
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
						Bucket:  make(map[int]int),
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
					SponsoringNode: "tcp://127.0.0.1:5501",
				}
				executeCommand("tcp://127.0.0.1:5509", join_9)

				time.Sleep(1 * time.Second)
				//joining node 9
				fix_finger_table_cmd := &Command{
					Do:     "fix-ring-fingers",
				}
				//command to fix the finger tables
				 for _,addr := range nodes_in_ring{
					time.Sleep(1*time.Second)
					fmt.Println("update finger table of:",addr )
				 	executeCommand(addr, fix_finger_table_cmd)
				 }
				 time.Sleep(1*time.Second)
				
				 // Put data commands
				 put_data_cmd := &Command{
					Do:     "Put",
					Data: 	&DataStruct{
						Key: 10,
						Value: 100,
					},
					ReplyTo: "tcp://127.0.0.1:5501",
				}
				executeCommand("tcp://127.0.0.1:5501", put_data_cmd)

				time.Sleep(1 * time.Second)

				 // Put data commands
				 put_data_cmd2 := &Command{
					Do:     "Put",
					Data: 	&DataStruct{
						Key: 5,
						Value: 500,
					},
					ReplyTo: "tcp://127.0.0.1:5501",
				}
				executeCommand("tcp://127.0.0.1:5501", put_data_cmd2)
				time.Sleep(1 * time.Second)

				put_data_cmd3 := &Command{
					Do:     "Put",
					Data: 	&DataStruct{
						Key:11,
						Value: 1100,
					},
					ReplyTo: "tcp://127.0.0.1:5501",
				}
				executeCommand("tcp://127.0.0.1:5501", put_data_cmd3)
				time.Sleep(1 * time.Second)

				get_cmd := &Command{
					Do:     "Get",
					Data: 	&DataStruct{
						Key:10,
					},
					ReplyTo: "tcp://127.0.0.1:5514",
				}
				executeCommand("tcp://127.0.0.1:5514", get_cmd)

				get_list_cmd := &Command{
					Do:     "list-items",
					ReplyTo: "tcp://127.0.0.1:5508",
				}
				executeCommand("tcp://127.0.0.1:5514", get_list_cmd)


				time.Sleep(1000 * time.Second)
			}
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
				msg,_:=coordinatorClient.Recv(0) // get acknowledgement
				fmt.Println(msg)
				mu.Unlock()
			}
			func executeBucketCommand(address string, command *Command) {
				var mu sync.Mutex
				mu.Lock()
				context, _ := zmq.NewContext()
				coordinatorClient, _ := context.NewSocket(zmq.REQ)
				coordinatorClient.Connect(address)
				// Command marshalled
				marshalledJson, _ := json.Marshal(command)
				// Instructing Node to perform Command
				coordinatorClient.SendBytes(marshalledJson, 0)
				msg,_:=coordinatorClient.RecvBytes(0) // get acknowledgement
				fmt.Println(msg)
				mu.Unlock()
			}
