package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
)

func checkError(err error) {
	if err != nil {
		fmt.Println("13")
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func Int32ToBytes(i int) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}

func main() {
	l, err := net.Listen("tcp", ":8080")
	checkError(err)
	cnt := 0
	for {
		cnt++
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Can't Accept")
			log.Panic(err)
			conn.Close()
			continue
		}
		fmt.Println(cnt)
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	//auth
	buf := make([]byte, 10240)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Network Error.Terminating..\n")
		return
	}
	if buf[0] != 0x05 || buf[1] != 0x01 || buf[2] != 0x00 {
		fmt.Printf("Protocol implementation error.Terminating...\n")
		conn.Write([]byte{0x05, 0xFF})
		return
	}

	_, err = conn.Write([]byte{0x05, 0x00})
	if err != nil {
		fmt.Printf("Network Error.Terminating..\n")
		return
	}
	//auth success

	//fmt.Println("auth success")
	var b []byte
	b = make([]byte, 10240)
	n, err := conn.Read(b)
	go func() {
		defer conn.Close()
		//TCP server
		servers := "127.0.0.1:1200"
		stcpaddr, _ := net.ResolveTCPAddr("tcp4", servers)
		server, cerr := net.DialTCP("tcp4", nil, stcpaddr)
		if cerr != nil {
			fmt.Printf("Network error.Terminating..\n")
			fmt.Println(cerr.Error())
			return
		}
		defer server.Close()
		b = b[:n]

		//send addr and port
		buf = Int32ToBytes(n)
		_, err = server.Write(buf)
		_, err = server.Write(b)
		if err != nil {
			fmt.Printf("83 Network Error.Terminating..\n")
			return
		}
		//send success

		//transport
		go func() {
			b = make([]byte, 10240)
			for {
				n, err := conn.Read(b)
				if err != nil {
					break
				}
				b = b[:n]
				if err != nil {
					break
				}
				//length
				var buf []byte
				buf = make([]byte, 4)
				buf = Int32ToBytes(n)
				_, err = server.Write(buf)
				_, err = server.Write(b)
				if err != nil {
					fmt.Printf("Network Error.Terminating..\n")
					break
				}
			}
		}()

		var buf []byte
		buf = make([]byte, 10240)
		for {
			n, err := server.Read(buf)
			buf = buf[:n]
			if err != nil {
				fmt.Printf("NetworkError.Cancelling...\n")
				fmt.Println(err.Error())
				break
			}
			_, err = conn.Write(buf)
			if err != nil {
				fmt.Printf("NetworkError.Cancelling...\n")
				fmt.Println(err.Error())
				break
			}
		}
	}()
	//send to client

	var tmpbuf []byte
	tmpbuf = make([]byte, n)
	for i := 0; i < n; i++ {
		tmpbuf[i] = b[i]
	}
	tmpbuf[1] = 0x00
	_, err = conn.Write(tmpbuf)
}
