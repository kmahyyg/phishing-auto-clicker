package config

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	imapcli "github.com/emersion/go-imap/client"
	"github.com/go-playground/validator/v10"
	"io/ioutil"
	"log"
	"net"
	"phishingAutoClicker/utils"
	"time"
)

var XOR_KEY = []byte{0x8A}

type MailConfigFile struct {
	Protocol          string `json:"protocol" validate:"oneof=imap imaps,required"`
	ServerAddr        string `json:"server" validate:"required"`
	UserEmail         string `json:"user_email" validate:"required"`
	Password          string `json:"password" validate:"required"`
	SaveTo            string `json:"save_to,omitempty"`
	IsTLS             bool   `json:"enableTLS,omitempty"`
	NoTLSVerification int    `json:"noTlsVerification,omitempty" validate:"min=1,max=2"` // 1 = none, 2 = normal
	mailClient        *imapcli.Client
	relyNetConn       net.Conn
}

func (c *MailConfigFile) validate() error {
	vTor := validator.New()
	return vTor.Struct(c)
}

func (c *MailConfigFile) Load(path string) error {
	st, stype := utils.CheckExists(path)
	if st != true || stype < 0 {
		return errors.New("config file not found")
	}
	fData, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	orifData := utils.XORStream(XOR_KEY, fData)
	err = json.Unmarshal(orifData, c)
	if err != nil {
		return err
	}
	return c.validate()
}

func (c *MailConfigFile) StartWorker(mode int) {
	if mode != 1 && mode != 2 {
		panic(errors.New("unknown work mode"))
	}
	c.startEmailEventLoop(mode)
}

func (c *MailConfigFile) createMailClient() error {
	var conn net.Conn
	var err error
	if c.IsTLS {
		conn, err = tls.Dial("tcp", c.ServerAddr, &tls.Config{
			InsecureSkipVerify: c.NoTLSVerification == 1,
		})
	} else {
		conn, err = net.Dial("tcp", c.ServerAddr)
	}
	if err != nil {
		return err
	}
	conn.SetDeadline(time.Now().Add(time.Second * 180))
	c.relyNetConn = conn
	switch c.Protocol {
	case "imap":
		c.mailClient, err = imapcli.New(conn)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("specified protocol is not supported")
	}
}

func (c *MailConfigFile) startEmailEventLoop(worktype int) {
	for {
		// instantiate mail client
		time.Sleep(5 * time.Second)
		err := c.createMailClient()
		if err != nil {
			panic(err)
		}
		log.Println("Client created.")
		// login and get mailbox
		mbox, err := utils.LoginMailboxAndCheck(c.mailClient, c.UserEmail, c.Password)
		if err != nil {
			panic(err)
		}
		log.Println("loginMailboxAndCheck successful.")
		if mbox.UnseenSeqNum == 0 {
			log.Println("No more new message.")
			return
		}
		// fetch msgs
		msgsLst, err := utils.FetchMsgRangeFromInbox(mbox.UnseenSeqNum, mbox.Messages, c.mailClient)
		if err != nil {
			log.Println(err)
			c.exitConn()
			continue
		}
		log.Println("fetchMsgRangeFromInbox successful.")
		// for each message, op
		noMoreMsgs := false
		for {
			msg, exist := <-msgsLst
			// if no more messages, break
			if !exist {
				// early exit
				log.Println("No more messages to fetch.")
				noMoreMsgs = true
				break
			}
			// parse each message with workload type, log incoming, return attachment bytes
			err = utils.ParseEmailMessageAndWork(msg, worktype)
			if err != nil {
				log.Println(err)
				continue
			}
		}
		if noMoreMsgs {
			c.exitConn()
			continue
		}
		c.exitConn()
	}
}

func (c *MailConfigFile) exitConn() {
	c.mailClient.Logout()
	c.relyNetConn.Close()
}
