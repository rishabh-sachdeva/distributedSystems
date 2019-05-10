package main

import (
	"fmt"
     zmq "github.com/zeromq/goczmq"
    "time"
)
type Message_worker{
	SocketAddress string
	Value int
	Neighbour1Socket string
	neighbour2Socket string 
	
}

func worker(Message_worker message){
	value := message.Value
	my_address := message.SocketAddress
	neighbour1 := message.Neighbour1Socket
	neighbour2 := message.Neighbour2Socket
	context,_ := zmq.NewContext()
	zmq_socket = context.NewSocket(zmq.REQ)
	zmq_socket.bind(my_address)
	
	if neighbour1 != nil{
		
	}
	
	
}

func main(){
	num_nodes := 4
	value_map = make(map[int]int)
	value_map[1]=10
	value_map[2]=20
	value_map[3]=30
	value_map[4]=40
	
	var arr_nodes [num_nodes]int
	arr_nodes[0] = 1
	arr_nodes[1] = 2
	arr_nodes[2] = 3
	arr_nodes[3] = 4
	
	socket_address_map = make(map[int]string)
	for i:=1; i<=num_nodes;i++{		
		//assign addresses to node workers.
		socket_address_map[i] = "tcp://127.0.0.1:555"+i
		
					
	}
	for i:=1; i<=num_nodes;i++{
		neigh1, ok := socket_address_map[i-1]
		neigh2, ok := socket_address_map[i+1]
		msg_worker := &Message_worker{socket_address_map[i],value_map[i],neigh1, neigh2}
		go worker(msg_worker)
	}
}
