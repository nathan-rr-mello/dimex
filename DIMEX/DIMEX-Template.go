/*  Construido como parte da disciplina: FPPD - PUCRS - Escola Politecnica
    Professor: Fernando Dotti  (https://fldotti.github.io/)
    Modulo representando Algoritmo de Exclusão Mútua Distribuída:
    Semestre 2023/1
	Aspectos a observar:
	   mapeamento de módulo para estrutura
	   inicializacao
	   semantica de concorrência: cada evento é atômico
	   							  módulo trata 1 por vez
	Q U E S T A O
	   Além de obviamente entender a estrutura ...
	   Implementar o núcleo do algoritmo ja descrito, ou seja, o corpo das
	   funcoes reativas a cada entrada possível:
	   			handleUponReqEntry()  // recebe do nivel de cima (app)
				handleUponReqExit()   // recebe do nivel de cima (app)
				handleUponDeliverRespOk(msgOutro)   // recebe do nivel de baixo
				handleUponDeliverReqEntry(msgOutro) // recebe do nivel de baixo
*/

package DIMEX

import (
	"SD/PP2PLink"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ------------------------------------------------------------------------------------
// ------- principais tipos
// ------------------------------------------------------------------------------------

type State int // enumeracao dos estados possiveis de um processo
const (
	noMX State = iota
	wantMX
	inMX
)

func (s State) String() string {
	switch s {
	case noMX:
		return "noMX"
	case wantMX:
		return "wantMX"
	case inMX:
		return "inMX"
	default:
		return "Unknown"
	}
}

type dmxReq int // enumeracao dos estados possiveis de um processo
const (
	ENTER dmxReq = iota
	EXIT
	SNAPSHOT
)

type dmxResp struct { // mensagem do módulo DIMEX infrmando que pode acessar - pode ser somente um sinal (vazio)
	// mensagem para aplicacao indicando que pode prosseguir
}

type DIMEX_Module struct {
	Req              chan dmxReq  // canal para receber pedidos da aplicacao (REQ e EXIT)
	Ind              chan dmxResp // canal para informar aplicacao que pode acessar
	addresses        []string     // endereco de todos, na mesma ordem
	id               int          // identificador do processo - é o indice no array de enderecos acima
	st               State        // estado deste processo na exclusao mutua distribuida
	waiting          []bool       // processos aguardando tem flag true
	lcl              int          // relogio logico local
	reqTs            int          // timestamp local da ultima requisicao deste processo
	nbrResps         int
	dbg              bool
	takingSnapshot   bool
	snapshotReceived []bool
	nextSnapshotId   int
	snapshot         Snapshot

	Pp2plink *PP2PLink.PP2PLink // acesso aa comunicacao enviar por PP2PLinq.Req  e receber por PP2PLinq.Ind
}

type Snapshot struct {
	snapshotId             int
	lcl                    int
	st                     State
	waiting                []bool
	messagesDuringSnapshot []string
}

// ------------------------------------------------------------------------------------
// ------- inicializacao
// ------------------------------------------------------------------------------------

func NewDIMEX(_addresses []string, _id int, _dbg bool) *DIMEX_Module {

	p2p := PP2PLink.NewPP2PLink(_addresses[_id], _dbg)

	dmx := &DIMEX_Module{
		Req: make(chan dmxReq, 1),
		Ind: make(chan dmxResp, 1),

		addresses:        _addresses,
		id:               _id,
		st:               noMX,
		waiting:          make([]bool, len(_addresses)),
		lcl:              0,
		reqTs:            0,
		dbg:              _dbg,
		takingSnapshot:   false,
		snapshotReceived: make([]bool, len(_addresses)),
		nextSnapshotId:   0,

		Pp2plink: p2p}

	for i := 0; i < len(dmx.waiting); i++ {
		dmx.waiting[i] = false
	}
	dmx.Start()
	dmx.outDbg("Init DIMEX!")
	return dmx
}

// ------------------------------------------------------------------------------------
// ------- nucleo do funcionamento
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) Start() {

	go func() {
		for {
			select {
			case dmxR := <-module.Req: // vindo da  aplicação
				if dmxR == ENTER {
					module.outDbg("app pede mx")
					module.handleUponReqEntry() // ENTRADA DO ALGORITMO

				} else if dmxR == EXIT {
					module.outDbg("app libera mx")
					module.handleUponReqExit() // ENTRADA DO ALGORITMO
				} else if dmxR == SNAPSHOT {
					if !module.takingSnapshot && module.id == 0 {
						module.handleUponReqSnapshot() // ENTRADA DO ALGORITMO
					}
				}

			case msgOutro := <-module.Pp2plink.Ind: // vindo de outro processo
				//fmt.Printf("dimex recebe da rede: ", msgOutro)
				if strings.Contains(msgOutro.Message, "respOK") {
					module.outDbg("         <<<---- responde! " + msgOutro.Message)
					module.handleUponDeliverRespOk(msgOutro) // ENTRADA DO ALGORITMO

				} else if strings.Contains(msgOutro.Message, "reqEntry") {
					module.outDbg("          <<<---- pede??  " + msgOutro.Message)
					module.handleUponDeliverReqEntry(msgOutro) // ENTRADA DO ALGORITMO
				} else if strings.Contains(msgOutro.Message, "takeSnapshot") {
					module.outDbg("          <<<---- takeSnapshot??  " + msgOutro.Message)
					module.handleUponDeliveryReqSnapshot(msgOutro) // ENTRADA DO ALGORITMO
				}

				if module.takingSnapshot {
					if !strings.Contains(msgOutro.Message, "takeSnapshot") {
						var fromId int
						fmt.Sscanf(msgOutro.Message, "from:%d msgType:%s timestamp:%d", &fromId, new(string), new(string))
						if !module.snapshotReceived[fromId] {
							module.snapshot.messagesDuringSnapshot[fromId] = module.snapshot.messagesDuringSnapshot[fromId] + ", " + msgOutro.Message
						}
					}
				}
			}
		}
	}()
}

// ------------------------------------------------------------------------------------
// ------- tratamento de pedidos vindos da aplicacao
// ------- UPON ENTRY
// ------- UPON EXIT
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) handleUponReqEntry() {
	/*
					upon event [ dmx, Entry  |  r ]  do
		    			lts.ts++
		    			myTs := lts
		    			resps := 0
		    			para todo processo p
							trigger [ pl , Send | [ reqEntry, r, myTs ]
		    			estado := queroSC
	*/

	module.lcl++
	module.reqTs = module.lcl
	module.nbrResps = 0

	message := fmt.Sprintf("from:%d msgType:%s timestamp:%d", module.id, "reqEntry", module.reqTs)

	for index, address := range module.addresses {
		if index != module.id {
			module.sendToLink(address, message, module.addresses[module.id])
		}
	}
	module.st = wantMX
}

func (module *DIMEX_Module) handleUponReqExit() {
	/*
						upon event [ dmx, Exit  |  r  ]  do
		       				para todo [p, r, ts ] em waiting
		          				trigger [ pl, Send | p , [ respOk, r ]  ]
		    				estado := naoQueroSC
							waiting := {}
	*/

	for index, address := range module.addresses {
		if module.waiting[index] {
			module.sendToLink(address, "respOK", module.addresses[module.id])
		}
	}

	clear(module.waiting)
	module.st = noMX
}

func (module *DIMEX_Module) handleUponReqSnapshot() {
	module.takingSnapshot = true
	module.snapshot = Snapshot{
		snapshotId:             module.nextSnapshotId,
		lcl:                    module.lcl,
		st:                     module.st,
		waiting:                module.waiting,
		messagesDuringSnapshot: make([]string, len(module.addresses)),
	}
	for index, address := range module.addresses {
		if module.id != index {
			message := fmt.Sprintf("from:%d msgType:%s timestamp:%d", module.id, "takeSnapshot", module.reqTs)
			module.sendToLink(address, message, module.addresses[module.id])
		}
	}
	module.snapshotReceived[module.id] = true
}

// ------------------------------------------------------------------------------------
// ------- tratamento de mensagens de outros processos
// ------- UPON respOK
// ------- UPON reqEntry
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) handleUponDeliverRespOk(msgOutro PP2PLink.PP2PLink_Ind_Message) {
	/*
						upon event [ pl, Deliver | p, [ respOk, r ] ]
		      				resps++
		      				se resps = N
		    				então trigger [ dmx, Deliver | free2Access ]
		  					    estado := estouNaSC

	*/

	module.nbrResps++
	if module.nbrResps == len(module.addresses)-1 { //todos os outros processos responderam
		module.Ind <- dmxResp{}
		module.st = inMX
	}
}

func (module *DIMEX_Module) handleUponDeliverReqEntry(msgOutro PP2PLink.PP2PLink_Ind_Message) {
	// outro processo quer entrar na SC
	/*
						upon event [ pl, Deliver | p, [ reqEntry, r, rts ]  do
		     				se (estado == naoQueroSC)   OR
		        				 (estado == QueroSC AND  myTs >  ts)
							então  trigger [ pl, Send | p , [ respOk, r ]  ]
		 					senão
		        				se (estado == estouNaSC) OR
		           					 (estado == QueroSC AND  myTs < ts)
		        				então  postergados := postergados + [p, r ]
		     					lts.ts := max(lts.ts, rts.ts)
	*/

	var otherTs int
	var otherId int

	_, err := fmt.Sscanf(msgOutro.Message, "from:%d msgType:%s timestamp:%d", &otherId, new(string), &otherTs)
	if err != nil {
		fmt.Println("Error parsing malformed message: received ->", msgOutro.Message, err)
		return
	}

	if module.st == noMX || (module.st == wantMX && (module.reqTs > otherTs || (module.reqTs == otherTs && module.id > otherId))) { //se o processo nao quer a mx ou o outro tem prioridade
		module.sendToLink(module.addresses[otherId], "respOK", module.addresses[module.id])
	} else if module.st == inMX || (module.st == wantMX && (module.reqTs < otherTs || (module.reqTs == otherTs && module.id < otherId))) { //se o processo está na mx ou se tem prioridade
		module.waiting[otherId] = true
	}

	module.lcl = max(module.lcl, otherTs)
}

func (module *DIMEX_Module) handleUponDeliveryReqSnapshot(msgOutro PP2PLink.PP2PLink_Ind_Message) {
	var fromId int
	var fromTs int

	fmt.Sscanf(msgOutro.Message, "from:%d msgType:%s timestamp:%d", &fromId, new(string), &fromTs)

	if module.takingSnapshot {
		module.snapshotReceived[fromId] = true
		allReceived := true
		for i := 0; i < len(module.snapshotReceived); i++ {
			if !module.snapshotReceived[i] {
				allReceived = false
				break
			}
		}
		if allReceived {
			snapshot := module.snapshot
			file, err := os.OpenFile("snapshot"+strconv.Itoa(module.id)+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()
			file.WriteString("-------Snapshot-------\n")
			file.WriteString("Snapshot ID: " + strconv.Itoa(snapshot.snapshotId) + "\n")
			file.WriteString("Process ID: " + strconv.Itoa(module.id) + "\n")
			file.WriteString("lcl: " + strconv.Itoa(snapshot.lcl) + "\n")
			file.WriteString("Status: " + snapshot.st.String() + "\n")
			file.WriteString("Waiting: " + fmt.Sprint(snapshot.waiting) + "\n")
			file.WriteString("Messages received during Sanpshot\n")
			for i := 0; i < len(snapshot.messagesDuringSnapshot); i++ {
				file.WriteString("  Process " + strconv.Itoa(i) + ": " + snapshot.messagesDuringSnapshot[i] + "\n")
			}
			file.WriteString("-------End Snapshot-------\n\n")
			module.takingSnapshot = false
			module.snapshotReceived = make([]bool, len(module.addresses))
			module.nextSnapshotId++
		}
	} else {
		module.snapshotReceived[fromId] = true
		module.handleUponReqSnapshot()
	}
}

// ------------------------------------------------------------------------------------
// ------- funcoes de ajuda
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) sendToLink(address string, content string, space string) {
	module.outDbg(space + " ---->>>>   to: " + address + "     msg: " + content)
	module.Pp2plink.Req <- PP2PLink.PP2PLink_Req_Message{
		To:      address,
		Message: content}
}

func before(oneId, oneTs, othId, othTs int) bool {
	if oneTs < othTs {
		return true
	} else if oneTs > othTs {
		return false
	} else {
		return oneId < othId
	}
}

func (module *DIMEX_Module) outDbg(s string) {
	if module.dbg {
		fmt.Println(". . . . . . . . . . . . [ DIMEX : " + s + " ]")
	}
}
