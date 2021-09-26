package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

//Variáveis globais interessantes para o processo
var err string
var myPort string          //porta do meu servidor
var nServers int           //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores
//dos outros processos
var ServConn *net.UDPConn //conexão do meu servidor (onde recebo
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
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func doClientJob(otherProcess int, i int) {
	//Enviar uma mensagem (com valor i) para o servidor do processo otherServer
	msg := strconv.Itoa(i)
	buf := []byte(msg)
	_, err := CliConn[otherProcess].Write(buf)
	if err != nil {
		fmt.Println(msg, err)
	}

	time.Sleep(time.Second * 1)
}

func initConnections() {
	myPort = os.Args[1]
	nServers = len(os.Args) - 2
	/*Esse 2 tira o nome (no caso Process) e tira a primeira porta (que
	é a minha). As demais portas são dos outros processos*/
	CliConn = make([]*net.UDPConn, nServers)

	/*Outros códigos para deixar ok a conexão do meu servidor (onde recebo msgs).
	O processo já deve ficar habilitado a receber msgs.*/

	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1"+myPort)
	CheckError(err)
	ServConn, err = net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	/*Outros códigos para deixar ok as conexões com os servidores dos outros processos.
	Colocar tais conexões no vetor CliConn.*/

	for servidores := 0; servidores < nServers; servidores++ {
		ServerAddr, err := net.ResolveUDPAddr("udp",
			"127.0.0.1"+os.Args[2+servidores])
		CheckError(err)
		Conn, err := net.DialUDP("udp", nil, ServerAddr)
		CliConn[servidores] = Conn
		CheckError(err)
	}
}

func main() {
	initConnections()
	//O fechamento de conexões deve ficar aqui, assim só fecha
	//conexão quando a main morrer
	defer ServConn.Close()
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}

	ch := make(chan string) //canal que guarda itens lidos do teclado
	go readInput(ch)        //chamar rotina que ”escuta” o teclado

	go doServerJob()
	for {
		// Verificar (de forma não bloqueante) se tem algo no
		// stdin (input do terminal)
		select {
		case x, valid := <-ch:
			if valid {
				fmt.Printf("Recebi do teclado: %s \n", x)
				for j := 0; j < nServers; j++ {
					go doClientJob(j, 100)
				}
			} else {
				fmt.Println("Canal fechado!")
			}
		default:
			// Fazer nada...
			// Mas não fica bloqueado esperando o teclado
			time.Sleep(time.Second * 1)
		}
		// Esperar um pouco
		time.Sleep(time.Second * 1)

	}
}
