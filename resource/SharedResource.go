package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Erro:", err)
		os.Exit(0)
	}
}

func PrintError(err error) {
	if err != nil {
		fmt.Println("Erro:", err)
	}
}

func main() {
	Address, err := net.ResolveUDPAddr("udp", ":10001")
	CheckError(err)
	Connection, err := net.ListenUDP("udp", Address)
	CheckError(err)
	defer Connection.Close()

	fmt.Println("SharedResource process started!")
	fmt.Println("Waiting for connections...")

	buf := make([]byte, 1024)
	for {
		/*Loop infinito para receber mensagem e escrever todo
		conteúdo (processo que enviou, relógio recebido e texto) na tela*/

		//Ler (uma vez somente) da conexão UDP a mensagem
		n, addr, err := Connection.ReadFromUDP(buf)
		//Escrever na tela a msg recebida (indicando o endereço de quem enviou)
		message := string(buf[0:n])
		fmt.Printf("Received message '%s' from %s\n", message, addr)

		idxId := strings.Index(message, "id: ") + len("id: ")
		idxClock := strings.Index(message, "logical clock: ") + len("logical clock: ")

		id, _ := strconv.Atoi(message[idxId:idxClock])
		logicalClock, _ := strconv.ParseUint(message[idxClock:], 10, 64)

		fmt.Printf("received -> id: %d - logical clock: %d\n", id, logicalClock)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
