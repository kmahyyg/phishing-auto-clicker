package utils

import (
	"errors"
	"fmt"
	imaplib "github.com/emersion/go-imap"
	imapcli "github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"io"
	"io/ioutil"
	"log"
	"mvdan.cc/xurls/v2"
	"phishingAutoClicker/common"
	"strconv"
)

func LoginMailboxAndCheck(mailClient *imapcli.Client, username string, password string) (*imaplib.MailboxStatus, error) {
	// login with creds
	err := mailClient.Login(username, password)
	if err != nil {
		panic(err)
	}
	log.Println("Login successful.")
	common.GlobalCred_Username = username
	common.GlobalCred_Password = password

	// list mailboxes
	mailboxes := make(chan *imaplib.MailboxInfo, 10)
	done := make(chan error, 1)
	// async request, recv all
	go func() {
		done <- mailClient.List("", "*", mailboxes)
	}()

	log.Printf("List Mailboxes: ")
	foundInboxFlag := false
	for m := range mailboxes {
		fmt.Println(" " + m.Name + " ")
		if m.Name == "INBOX" {
			foundInboxFlag = true
			log.Println("Mailbox: INBOX found.")
			break
		}
	}

	if err = <-done; err != nil {
		log.Fatalln(err)
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

func FetchMsgRangeFromInbox(start uint32, end uint32, client *imapcli.Client) (msgList chan *imaplib.Message, err error) {
	// build fetch range, start and end must be larger than 0
	if start > end {
		return nil, errors.New("invalid range for mail message sequence num")
	}
	seqSet := new(imaplib.SeqSet)
	seqSet.AddRange(start, end)
	log.Printf("Fetching unread from %d to %d in inbox \n", start, end)
	msgList = make(chan *imaplib.Message, 100)
	// what should we fetch? all of email body
	var section imaplib.BodySectionName
	// BUG HERE: QQMail DOES NOT support "ALL"
	// Exchange support it but has some limitations
	// switch to "ENVELOPE" to get the whole email header
	err = client.Fetch(seqSet, []imaplib.FetchItem{section.FetchItem(), imaplib.FetchEnvelope}, msgList)
	if err != nil {
		return nil, err
	}
	return msgList, nil
}

func ParseEmailMessageAndWork(msg *imaplib.Message, worktype int, mailClient *imapcli.Client) (err error) {
	// check input
	if worktype != 1 && worktype != 2 {
		return errors.New("invalid worktype")
	}
	// parse email message
	var sections imaplib.BodySectionName
	r := msg.GetBody(&sections)
	if r == nil {
		return errors.New("server does not return message body")
	}
	err = markEmailMessageAsSeen(msg.SeqNum, mailClient)
	if err != nil {
		log.Println("Mark message as seen failed. Continue.")
		log.Println(err)
	}
	mailReader, err := mail.CreateReader(r)
	if err != nil {
		return err
	}
	// log incoming
	fromAddr := msg.Envelope.From
	if len(fromAddr) == 0 {
		log.Println("Mail Message FromAddr is empty!")
	}
	fromAddrStrLst := []string{}
	for _, v := range fromAddr {
		fromAddrStrLst = append(fromAddrStrLst, v.Address())
	}
	log.Printf("Incoming transmission: [ %s ] from [ %v ] \n", msg.Envelope.Subject, fromAddrStrLst)
	// get each parts
	foundFinal := false
	for {
		if foundFinal {
			return nil
		}
		p, err := mailReader.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			// text part
			if worktype == 1 {
				continue
			}
			// worktype == 2, means we need to extract url
			b, _ := ioutil.ReadAll(p.Body)
			// extract url
			urlRegEx := xurls.Strict()
			retUrl := urlRegEx.FindString(string(b))
			if retUrl != "" && retUrl[:4] == "http" {
				log.Printf("Found url: %s \n", retUrl)
				foundFinal = true
				// click URL
				go clickBaitOnline(retUrl)
			}
		case *mail.AttachmentHeader:
			// attachment
			if worktype == 2 {
				continue
			}
			// we only want the very first attachment
			foundFinal = true
			// get attachment
			fdName, _ := h.Filename()
			log.Printf("Found attachment: %s \n", fdName)
			// save attachment
			b, _ := ioutil.ReadAll(p.Body)
			go clickBaitOffline(fdName, b)
			foundFinal = true
		}
	}
	return errors.New("single email part parsed, but no satisfied data found")
}

func markEmailMessageAsSeen(seqNum uint32, mailClient *imapcli.Client) error {
	seqSeth := new(imaplib.SeqSet)
	mailSeqNumStr := strconv.FormatUint(uint64(seqNum), 10)
	_ = seqSeth.Add(mailSeqNumStr)
	item := imaplib.FormatFlagsOp(imaplib.AddFlags, true)
	flags := []interface{}{imaplib.SeenFlag}
	err := mailClient.Store(seqSeth, item, flags, nil)
	if err != nil {
		return err
	}
	log.Printf("Mail [Seq No. %s ] marked as seen. (Exchange server do this by default) \n", mailSeqNumStr)
	return nil
}
