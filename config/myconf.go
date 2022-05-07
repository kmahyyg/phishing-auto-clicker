package config

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	imaplib "github.com/emersion/go-imap"
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
	Protocol        string `json:"protocol" validate:"oneof=imap imaps,required"`
	ServerAddr      string `json:"server" validate:"required"`
	UserEmail       string `json:"user_email" validate:"required"`
	Password        string `json:"password" validate:"required"`
	SaveTo          string `json:"save_to,omitempty"`
	IsTLS           bool   `json:"enableTLS,omitempty"`
	TLSVerification int    `json:"tlsVerification,omitempty" validate:"min=1,max=2"` // 1 = none, 2 = normal
	mailClient      *imapcli.Client
	relyNetConn     net.Conn
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
	switch mode {
	case 1:
		c.startEmailAttachmentEventLoop()
	case 2:
		c.startEmailLinkEventLoop()
	default:
		panic(errors.New("unknown work mode"))
	}
}

func (c *MailConfigFile) createMailClient() error {
	var conn net.Conn
	var err error
	if c.IsTLS {
		conn, err = tls.Dial("tcp", c.ServerAddr, &tls.Config{
			InsecureSkipVerify: c.TLSVerification == 1,
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

func loginMailboxAndCheck(mailClient *imapcli.Client, conf *MailConfigFile) (*imaplib.MailboxStatus, error) {
	// login with creds
	err := mailClient.Login(conf.UserEmail, conf.Password)
	if err != nil {
		panic(err)
	}
	log.Println("Login successful.")

	// list mailboxes
	mailboxes := make(chan *imaplib.MailboxInfo, 10)
	doneCh := make(chan error, 1)
	go func() {
		doneCh <- mailClient.List("", "*", mailboxes)
	}()

	log.Printf("\n List Mailboxes: ")
	foundInboxFlag := false
	for m := range mailboxes {
		log.Println(" " + m.Name + " ,")
		if m.Name == "INBOX" {
			foundInboxFlag = true
			log.Println("Mailbox: INBOX found.")
			break
		}
	}
	// check if inbox exists
	if !foundInboxFlag {
		err = errors.New("cannot find correct INBOX mailbox folder")
		log.Println(err)
		return nil, err
	}
	// select mailbox
	mbox, err := mailClient.Select("INBOX", false)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return mbox, nil
}

func (c *MailConfigFile) startEmailAttachmentEventLoop() {
	for {
		// instantiate mail client
		time.Sleep(5 * time.Second)
		err := c.createMailClient()
		if err != nil {
			panic(err)
		}
		log.Println("Client created.")
		// login and get mailbox
		mbox, err := loginMailboxAndCheck(c.mailClient, c)
		if err != nil {
			panic(err)
		}

		// download attachment
		// save attachment
		// if office, execute
		// if zip, extract
		// execute then kill after timeout
		// delete mail
		c.exitConn()
	}
}

func (c *MailConfigFile) exitConn() {
	c.mailClient.Logout()
	c.relyNetConn.Close()
}

func (c *MailConfigFile) startEmailLinkEventLoop() {
	for {
		time.Sleep(2 * time.Minute)

		// parse each mail
		// find the first link start with [SPACE]https://xxx[SPACE]
		// if not found, try [SPACE]http://xxx[SPACE]
		// open page, kill browser after timeout

	}
}
