package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

//Variáveis globais interessantes para o processo
var err string
var myId int                        //id do meu servidor
var myPort string                   //porta do meu servidor
var myLogicalClock uint64           //logical clock do meu servidor
var nServers int                    //qtde de outros processo
var ports map[int]string            //map com portas de cada id
var CliConn map[string]*net.UDPConn //map com conexões para os servidores dos outros processos por porta
var ServConn *net.UDPConn           //conexão do meu servidor (onde recebo
//mensagens dos outros processos)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
		os.Exit(0)
	}
}

func PrintError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
	}
}

func Max(a uint64, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func readInput(ch chan string) {
	// Rotina não-bloqueante que “escuta” o stdin
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}

func doServerJob() {
	buf := make([]byte, 1024)
	for {
		//Ler (uma vez somente) da conexão UDP a mensagem
		n, addr, err := ServConn.ReadFromUDP(buf)
		//Escrever na tela a msg recebida (indicando o endereço de quem enviou)
		message := string(buf[0:n])
		fmt.Println("Received ", message, " from ", addr)
		idx := strings.Index(message, "logical clock: ") + len("logical clock: ")
		msgLogicalClock, _ := strconv.ParseUint(message[idx:], 10, 64)
		myLogicalClock = Max(myLogicalClock, msgLogicalClock) + 1
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func doClientJob(otherProcessId int) {
	// Enviar mensagem para outro processo contendo meu id e logical clock
	myIdStr := strconv.Itoa(myId)
	myLogicalClockStr := strconv.FormatUint(myLogicalClock, 10)
	msg := "id: " + myIdStr + " - logical clock: " + myLogicalClockStr
	buf := []byte(msg)
	_, err := CliConn[ports[otherProcessId]].Write(buf)
	if err != nil {
		fmt.Println(msg, err)
	}

	time.Sleep(time.Second * 1)
}

func initConnections() {
	ports = map[int]string{
		1: ":10001",
		2: ":10002",
		3: ":10003",
		4: ":10004",
		5: ":10005",
		6: ":10006",
		7: ":10007",
		8: ":10008",
		9: ":10009",
	}

	myId, _ = strconv.Atoi(os.Args[1])
	myPort = ports[myId]
	nServers = len(os.Args) - 2
	//Esse 2 tira o nome (no caso Process) e o meu id. As demais portas são dos outros processos

	CliConn = make(map[string]*net.UDPConn)

	/*Outros códigos para deixar ok a conexão do meu servidor (onde recebo msgs).
	O processo já deve ficar habilitado a receber msgs.*/
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1"+myPort)
	CheckError(err)
	ServConn, err = net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	/*Outros códigos para deixar ok as conexões com os servidores dos outros processos.
	Colocar tais conexões no map CliConn.*/
	for servidores := 0; servidores < nServers; servidores++ {
		port := os.Args[2+servidores]
		ServerAddr, err := net.ResolveUDPAddr("udp",
			"127.0.0.1"+port)
		CheckError(err)
		Conn, err := net.DialUDP("udp", nil, ServerAddr)
		CliConn[port] = Conn
		CheckError(err)
	}
}

func main() {
	initConnections()
	myLogicalClock = 0
	//O fechamento de conexões deve ficar aqui, assim só fecha conexão quando a main morrer
	defer ServConn.Close()
	for _, Conn := range CliConn {
		defer Conn.Close()
	}

	fmt.Println("Meu id:", myId, "- Minha porta:", myPort)

	ch := make(chan string) //canal que guarda itens lidos do teclado
	go readInput(ch)        //chamar rotina que ”escuta” o teclado

	go doServerJob()
	for {
		// Verificar (de forma não bloqueante) se tem algo no stdin (input do terminal)
		select {
		case x, valid := <-ch:
			if valid {
				fmt.Printf("Recebi do teclado: %s \n", x)
				id, erro := strconv.Atoi(x)
				if erro == nil {
					if id != myId {
						go doClientJob(id)
					} else {
						myLogicalClock++
					}
				} else {
					fmt.Println("Entrada inválida!")
				}

			} else {
				fmt.Println("Canal fechado!")
			}
		default:
			// Fazer nada... Mas não fica bloqueado esperando o teclado
			time.Sleep(time.Second * 1)
		}
		// Esperar um pouco
		time.Sleep(time.Second * 1)

	}
}
