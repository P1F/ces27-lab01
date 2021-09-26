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
var ServConn *net.UDPConn           //conexão do meu servidor (onde recebo mensagens dos outros processos)
var myState string                  //define o estado do processo
const sharedResourceId int = 0      //define um id fixo para o SharedResource

const RELEASED string = "RELEASED"
const WANTED string = "WANTED"
const HELD string = "HELD"

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
		fmt.Printf("Received message '%s' from %s\n", message, addr)

		// TODO filtrar entrada WANT no input do usuário
		if strings.Contains(message, "WANT") { //yes, this can be a problem...
			//recebeu request que algum processo quer entrar na CS
			idxId := strings.Index(message, "[")
			idxClock := strings.Index(message, "Logical clock: ") + len("Logical clock: ")

			idStr := message[idxId+1 : idxId+2]
			logicalClockStr := message[idxClock:]

			fmt.Printf("Recebi pedido de entrada na CS -> id: %s, logical clock: %s\n", idStr, logicalClockStr)
		} else {
			//recebeu uma mensagem qualquer de um processo
			idx := strings.Index(message, "logical clock: ") + len("logical clock: ")
			msgLogicalClock, _ := strconv.ParseUint(message[idx:], 10, 64)
			myLogicalClock = Max(myLogicalClock, msgLogicalClock) + 1
			fmt.Printf("Logical clock updated to: %d\n", myLogicalClock)
		}

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func doClientJob(otherProcessId int) {
	// Enviar mensagem para outro processo contendo meu id e logical clock
	myIdStr := strconv.Itoa(myId)
	myLogicalClockStr := strconv.FormatUint(myLogicalClock, 10)
	msg := "Hello! Here's my info -> "
	msg += "id: " + myIdStr + " - logical clock: " + myLogicalClockStr
	buf := []byte(msg)
	_, err := CliConn[ports[otherProcessId]].Write(buf)
	if err != nil {
		fmt.Println(msg, err)
	}

	if otherProcessId == sharedResourceId {
		//avisar outros processos que quero acessar a CS
		myState = WANTED
		broadcastMsg := "Process [" + myIdStr + "] WANTs to enter CS! "
		broadcastMsg += "Logical clock: " + myLogicalClockStr
		buf2 := []byte(broadcastMsg)
		for port, Conn := range CliConn {
			if port != ports[sharedResourceId] && port != ports[myId] {
				_, err := Conn.Write(buf2)
				if err != nil {
					fmt.Println(broadcastMsg, err)
				}
			}
		}
	}

	time.Sleep(time.Second * 1)
}

func initConnections() {
	ports = map[int]string{
		0: ":10001", // porta fixa: utilizada para o SharedResource
		1: ":10002",
		2: ":10003",
		3: ":10004",
		4: ":10005",
		5: ":10006",
		6: ":10007",
		7: ":10008",
		8: ":10009",
		9: ":10010",
	}

	myId, _ = strconv.Atoi(os.Args[1])
	myPort = ports[myId]
	myState = RELEASED
	nServers = len(os.Args) - 2
	//Esse 2 tira o nome (no caso Process) e o meu id. As demais portas são dos outros processos

	CliConn = make(map[string]*net.UDPConn)

	/*Outros códigos para deixar ok a conexão do meu servidor (onde recebo msgs).
	O processo já deve ficar habilitado a receber msgs.*/
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1"+myPort)
	CheckError(err)
	ServConn, err = net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	/*Outros códigos para deixar ok a conexão com o servidor do SharedResource.
	Colocar tais conexões no map CliConn.*/
	SharedResourcesAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	CheckError(err)
	SharedResourcesConn, err := net.DialUDP("udp", nil, SharedResourcesAddr)
	CliConn[":10001"] = SharedResourcesConn
	CheckError(err)

	/*Outros códigos para deixar ok as conexões com os servidores dos outros processos.
	Colocar tais conexões no map CliConn.*/
	for servidores := 0; servidores < nServers; servidores++ {
		port := os.Args[2+servidores]
		if port != ":10001" {
			ServerAddr, err := net.ResolveUDPAddr("udp",
				"127.0.0.1"+port)
			CheckError(err)
			Conn, err := net.DialUDP("udp", nil, ServerAddr)
			CliConn[port] = Conn
			CheckError(err)
		}
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

	fmt.Printf("Process [%d] started at port %s\n", myId, myPort)

	ch := make(chan string) //canal que guarda itens lidos do teclado
	go readInput(ch)        //chamar rotina que ”escuta” o teclado

	go doServerJob()
	for {
		// Verificar (de forma não bloqueante) se tem algo no stdin (input do terminal)
		select {
		case x, valid := <-ch:
			if valid {
				fmt.Printf("Input received from keyboard: %s\n", x)
				id, erro := strconv.Atoi(x)
				if erro == nil {
					if id != myId { // chame rotina para envio de mensagens
						go doClientJob(id)
					} else {
						myLogicalClock++
						fmt.Printf("INTERNAL event! Logical clock updated to: %d\n", myLogicalClock)
					}
				} else {
					fmt.Printf("Input '%s' is not a number!\n", x)
				}
			} else {
				fmt.Println("Channel CLOSED!")
			}
		default:
			// Fazer nada... Mas não fica bloqueado esperando o teclado
			time.Sleep(time.Second * 1)
		}
		// Esperar um pouco
		time.Sleep(time.Second * 1)

	}
}
