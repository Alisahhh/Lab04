package main

import (
	"fmt"
	"net"
	"os"
	"log"
	"strconv"
	"encoding/binary"
)

type IOpair struct{
	in chan []byte
	out chan []byte
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}


func BytesToInt32(buf []byte) int {
	return int(binary.BigEndian.Uint32(buf))
}

func main() {
	l, err := net.Listen("tcp", ":1200")
	checkError(err)
	for{
		conn,err:=l.Accept()

		if err != nil {
			log.Panic(err)
			continue
		}
		go handleClientRequest(conn)
	}
}

func handleClientRequest(conn net.Conn){
		defer conn.Close()
		var b []byte
		b=make([]byte,4)
		_,err := conn.Read(b)
		n:=BytesToInt32(b)
		//fmt.Println(n)
		b=make([]byte,n)
		_, err = conn.Read(b)
		//b=b[:n]
		fmt.Println("--------------52------------------")
		fmt.Println(b)
		if err != nil {
			log.Println(err)
			return
		}

		//DNS
		var zgr getRequest
		if b[0] == 0x05 {
			b=b[:n]
			switch b[3] {
			case 0x01: //IP V4
				zgr.ADDR = net.IPv4(b[4], b[5], b[6], b[7]).String()
			case 0x03:                       //域名
				zgr.ADDR = string(b[5: n-2]) //b[4]表示域名的长度
			case 0x04: //IP V6
				zgr.ADDR = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
			}
			zgr.PORT = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
		} else {
			fmt.Printf("Protocol implementation error.Terminating...\n")
			return
		}
		//mt.Println(zgr.ADDR)
		//fmt.Println(zgr.PORT)
		sconn, err := net.Dial("tcp", net.JoinHostPort(zgr.ADDR, zgr.PORT))
		if err != nil {
			//conn.Write([]byte{0x00})
			log.Println(err)
			return
		}
		defer sconn.Close()
		fmt.Println(net.JoinHostPort(zgr.ADDR, zgr.PORT))
		//sconn, err := net.Dial("tcp", "127.0.0.1:80")
		if err != nil {
			fmt.Printf("NetworkError.Cancelling...\n 81")
			fmt.Println(err.Error())
			//delete(conns, addrxport2id(zgr.ADDR, zgr.PORT))
			return
		}
		//send to server
		go func(){
			for {
				var b []byte
				b = make([]byte, 4)
				_, err := conn.Read(b)
				n := BytesToInt32(b)
				//fmt.Println(n)
				b = make([]byte, n)
				_, err = conn.Read(b)
				//b=b[:n]
				fmt.Println("--------------97------------------")
				fmt.Println(b)
				if err != nil {
					log.Println(err)
					return
				}
				_, err = sconn.Write(b)
				if err != nil {
					fmt.Printf("NetworkError.Cancelling...103\n")
					fmt.Println(err.Error())
					//delete(conns, addrxport2id(zgr.ADDR, zgr.PORT))
					return
				}
			}
		}()
		//sent to client
		for{
			//fmt.Println("prepare to read")
			buf:=make([]byte,10240)
			n,err=sconn.Read(buf)
			buf=buf[:n]
			fmt.Println("--------------117------------------")
			fmt.Println(buf)
			//fmt.Println("Readok")
			if err!=nil {
				fmt.Printf("NetworkError.Cancelling...117\n")
				fmt.Println(err.Error())
				//delete(conns, addrxport2id(zgr.ADDR, zgr.PORT))
				return
			}
			//buf=buf[:n]
			//fmt.Println(buf)
			n,err=conn.Write(buf)
			if err!=nil {
				fmt.Printf("NetworkError.Cancelling...126\n")
				fmt.Println(err.Error())
				//delete(conns, addrxport2id(zgr.ADDR, zgr.PORT))
				return
			}
			//fmt.Println("Writeok")
		}
}

