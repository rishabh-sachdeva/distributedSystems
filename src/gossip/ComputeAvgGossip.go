package main

import (
 	"fmt"
     zmq "github.com/pebbe/zmq4"
    "time"
  	"strconv"

)

type Message_worker struct{
	SocketAddress string
	Value int
	Neighbour1Socket string
	Neighbour2Socket string 
	
}

func worker(message *Message_worker){
	//value := message.Value
	//fmt.Println("Connecting")

	my_address := message.SocketAddress
	neighbour1 := message.Neighbour1Socket
	neighbour2 := message.Neighbour2Socket
	context,_ := zmq.NewContext()
	sub,_ := context.NewSocket(zmq.SUB)

	sub.SetLinger(0)
	defer sub.Close()
	
	sub.Connect(neighbour1)
	sub.Connect(neighbour2)
	sub.SetSubscribe("")


	go Listen(sub)

	pub,_ := context.NewSocket(zmq.PUB)
	pub.Bind(my_address)

	for _ = range time.Tick(time.Second) {
        pub.Send("test", 0)
        fmt.Println("send", "test")
    }
	defer pub.Close()
}

func Listen(subscriber *zmq.Socket) {
    for {
        s, err := subscriber.Recv(0)
        if err != nil {
            fmt.Println(err)
            continue
        }
        fmt.Println("receiving", s)
    }
}

func main(){
	num_nodes := 4
	value_map := make(map[int]int)
	value_map[1]=10
	value_map[2]=20
	value_map[3]=30
	value_map[4]=40
	
	var arr_nodes [4]int
	arr_nodes[0] = 1
	arr_nodes[1] = 2
	arr_nodes[2] = 3
	arr_nodes[3] = 4
	
	socket_address_map := make(map[int]string)
	for i:=1; i<=num_nodes;i++{		
		//assign addresses to node workers.
		port := 5550 + i
		socket_address_map[i] = "tcp://127.0.0.1:"+strconv.Itoa(port)
		fmt.Println(socket_address_map[i])
					
	}
	for j:=1; j<=num_nodes;j++{
		neigh1,_ := socket_address_map[j-1]
		neigh2,_ := socket_address_map[j+1]
		fmt.Println("calling worker")
		msg_worker := &Message_worker{socket_address_map[j],value_map[j],neigh1, neigh2}
		fmt.Println(*msg_worker)
		go worker(msg_worker)
		
	}
	time.Sleep(7 * time.Second)

}