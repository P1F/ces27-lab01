package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
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
var myRequestQueue []int            //define a fila para guardar requests
var myRepliesCount int              //define um contador de replies
var myRequestMessage string         //armazena a mensagem quando o processo entra em WANTED
var requestTimestamp uint64         //armazena o logical clock depois de ter enviado o broadcast
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

func accessCS(mutex *sync.Mutex) {
	fmt.Println("Entrei na CS")
	mutex.Lock()
	//entrar na CS -> trocar estado para HELD
	myState = HELD
	mutex.Unlock()
	fmt.Println("Agora estou em HELD")

	myIdStr := strconv.Itoa(myId)

	//enviar mensagem para Shared Resource
	msg := "Message received from id [" + myIdStr + "] -> "
	msg += "msg: " + myRequestMessage + " - logical clock: " + strconv.FormatUint(myLogicalClock, 10)
	buf := []byte(msg)
	_, err := CliConn[ports[sharedResourceId]].Write(buf)
	if err != nil {
		fmt.Println(msg, err)
	}

	//dormir por 2s (só para simular quando sair da CS)
	fmt.Println("Trabalhando...")
	time.Sleep(time.Second * 3)

	//sair da CS -> trocar estado para RELEASED
	mutex.Lock()
	myState = RELEASED
	mutex.Unlock()
	fmt.Println("Saí da CS")
	fmt.Println("Agora estou em RELEASED")

	if len(myRequestQueue) != 0 { //reply para todos os processos na fila
		//preparar mensagem de reply
		msg2 := "REPLY: reply from [" + myIdStr + "] - "
		msg2 += "logical clock: " + strconv.FormatUint(myLogicalClock, 10)
		buf2 := []byte(msg2)

		//dar reply para todos os processos com requests enfileirados
		for _, id := range myRequestQueue {
			_, err := CliConn[ports[id]].Write(buf2)
			if err != nil {
				fmt.Println(msg2, err)
			}
		}
		fmt.Println("REPLY enviado para ", myRequestQueue)
	} else {
		fmt.Println("Fila vazia! Não enviarei replies")
	}

	//talvez não seja necessário, mas foi colocado por precaução
	mutex.Lock()
	myRepliesCount = 0
	myRequestQueue = nil
	requestTimestamp = 0
	myRequestMessage = ""
	mutex.Unlock()
}

func doServerJob(mutex *sync.Mutex) {
	buf := make([]byte, 1024)
	for {
		//Ler (uma vez somente) da conexão UDP a mensagem
		n, _, err := ServConn.ReadFromUDP(buf)
		//Escrever na tela a msg recebida (indicando o endereço de quem enviou)
		message := string(buf[0:n])

		// TODO filtrar entrada WANT no input do usuário
		if strings.Contains(message, "REQUEST:") && strings.Contains(message, "WANT") { //yes, this can be a problem...
			//recebeu request que algum processo quer entrar na CS
			idxId := strings.Index(message, "[")
			idxClock := strings.Index(message, "Logical clock: ") + len("Logical clock: ")

			otherIdStr := message[idxId+1 : idxId+2]
			otherLogicalClockStr := message[idxClock:]
			otherId, _ := strconv.Atoi(otherIdStr)
			otherLogicalClock, _ := strconv.ParseUint(otherLogicalClockStr, 10, 64)

			mutex.Lock()
			myLogicalClock = Max(myLogicalClock, otherLogicalClock) + 1
			mutex.Unlock()
			fmt.Printf("REQUEST RECEBIDO de [%d]! Logical clock updated to: %d\n", otherId, myLogicalClock)

			isMyPreference := false
			if myState == WANTED {
				if (requestTimestamp < otherLogicalClock) ||
					(requestTimestamp == otherLogicalClock && myId < otherId) {
					isMyPreference = true
				}
			}

			if myState == HELD || isMyPreference {
				//enfileirar o request de otherId sem dar reply
				fmt.Printf("Enfileirando %d...\n", otherId)
				mutex.Lock()
				myRequestQueue = append(myRequestQueue, otherId)
				mutex.Unlock()
				fmt.Println("Status da fila:", myRequestQueue)
			} else {
				//dar reply para otherId
				fmt.Println("Não tenho preferência...")
				msg := "REPLY: reply from [" + strconv.Itoa(myId) + "] - "
				msg += "logical clock: " + strconv.FormatUint(myLogicalClock, 10)
				buf := []byte(msg)
				_, err := CliConn[ports[otherId]].Write(buf)
				if err != nil {
					fmt.Println(msg, err)
				}

				fmt.Printf("REPLY enviado para %d\n", otherId)
			}
		} else if strings.Contains(message, "REPLY:") {
			//obter id do outro processo
			idxId := strings.Index(message, "[")
			otherIdStr := message[idxId+1 : idxId+2]
			otherId, _ := strconv.Atoi(otherIdStr)

			//obter logical clock do outro processo
			idxClock := strings.Index(message, "logical clock: ") + len("logical clock: ")
			msgLogicalClockStr := message[idxClock:]
			msgLogicalClock, _ := strconv.ParseUint(msgLogicalClockStr, 10, 64)
			mutex.Lock()
			myLogicalClock = Max(myLogicalClock, msgLogicalClock) + 1
			myRepliesCount++
			mutex.Unlock()
			fmt.Printf("REPLY %d/%d RECEBIDO do id [%d]! Logical clock updated to: %d\n", myRepliesCount, nServers-1, otherId, myLogicalClock)

			if myRepliesCount == nServers-1 {
				go accessCS(mutex)
			}
		} else {
			//recebeu uma mensagem qualquer de um processo
			idx := strings.Index(message, "logical clock: ") + len("logical clock: ")
			msgLogicalClock, _ := strconv.ParseUint(message[idx:], 10, 64)
			mutex.Lock()
			myLogicalClock = Max(myLogicalClock, msgLogicalClock) + 1
			mutex.Unlock()
			fmt.Printf("MENSAGEM RECEBIDA! Logical clock updated to: %d\n", myLogicalClock)
		}

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func doClientJob(otherProcessId int, message string, mutex *sync.Mutex) {
	// Enviar mensagem para outro processo contendo meu id e logical clock

	if otherProcessId == sharedResourceId && myState == RELEASED {
		//avisar outros processos que quero acessar a CS
		myIdStr := strconv.Itoa(myId)
		mutex.Lock()
		myRequestMessage = message
		myState = WANTED
		myRepliesCount = 0
		myRequestQueue = nil
		myLogicalClock++
		mutex.Unlock()
		fmt.Println("Agora estou em WANTED")
		broadcastMsg := "REQUEST: Process [" + myIdStr + "] WANTs to enter CS! "
		broadcastMsg += "Logical clock: " + strconv.FormatUint(myLogicalClock, 10)
		buf2 := []byte(broadcastMsg)
		for port, Conn := range CliConn {
			if port != ports[sharedResourceId] && port != ports[myId] {
				_, err := Conn.Write(buf2)
				if err != nil {
					fmt.Println(broadcastMsg, err)
				}
			}
		}
		mutex.Lock()
		requestTimestamp = myLogicalClock
		mutex.Unlock()
		fmt.Printf("REQUEST EM BROADCAST ENVIADO! Logical clock updated to: %d\n", myLogicalClock)
	} else {
		fmt.Printf("'%s' ignorado\n", message)
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
	myLogicalClock = 0
	myRepliesCount = 0
	myRequestQueue = nil
	myState = RELEASED
	myRequestMessage = ""
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
	var mutex sync.Mutex
	initConnections()
	//O fechamento de conexões deve ficar aqui, assim só fecha conexão quando a main morrer
	defer ServConn.Close()
	for _, Conn := range CliConn {
		defer Conn.Close()
	}

	fmt.Printf("Process [%d] started at port %s\n", myId, myPort)

	ch := make(chan string) //canal que guarda itens lidos do teclado
	go readInput(ch)        //chamar rotina que ”escuta” o teclado

	go doServerJob(&mutex)
	for {
		// Verificar (de forma não bloqueante) se tem algo no stdin (input do terminal)
		select {
		case x, valid := <-ch:
			if valid {
				fmt.Printf("Input received from keyboard: %s\n", x)
				id, erro := strconv.Atoi(x)
				if erro == nil && id == myId {
					myLogicalClock++
					fmt.Printf("INTERNAL event! Logical clock updated to: %d\n", myLogicalClock)
				} else {
					go doClientJob(sharedResourceId, x, &mutex)
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
