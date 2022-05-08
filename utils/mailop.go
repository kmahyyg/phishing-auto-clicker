package utils

import (
	"context"
	"errors"
	"fmt"
	imaplib "github.com/emersion/go-imap"
	imapcli "github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"mvdan.cc/xurls/v2"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"phishingAutoClicker/common"
	"runtime"
	"strings"
	"time"
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
	if start < end {
		return nil, errors.New("invalid range for mail message sequence num")
	}
	seqSet := new(imaplib.SeqSet)
	seqSet.AddRange(start, end)
	log.Printf("Fetching unread from %d to %d in inbox \n", start, end)
	msgList = make(chan *imaplib.Message, 100)
	// what should we fetch? all of email body
	var section imaplib.BodySectionName
	err = client.Fetch(seqSet, []imaplib.FetchItem{section.FetchItem()}, msgList)
	if err != nil {
		return nil, err
	}
	return msgList, nil
}

func ParseEmailMessageAndWork(msg *imaplib.Message, worktype int) (err error) {
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
	mailReader, err := mail.CreateReader(r)
	if err != nil {
		return err
	}
	// log incoming
	log.Printf("Incoming transmission: [ %s ] from [ %s ] \n", msg.Envelope.Subject, msg.Envelope.From)
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
	return errors.New("unknown error")
}

func clickBaitOnline(url string) {
	// if url is ending with .exe / .doc / .docm / .xlsm / .elf , download and execute it
	if strings.HasSuffix(url, ".exe") || strings.HasSuffix(url, ".doc") || strings.HasSuffix(url, ".docm") || strings.HasSuffix(url, ".xlsm") || strings.HasSuffix(url, ".elf") {
		// download	files
		resp, err := http.Get(url)
		if err != nil {
			log.Println(err)
			return
		}
		// get extensions
		lastIdx := strings.LastIndex(url, ".")
		fileExt := url[lastIdx:]
		respData, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			log.Println(err)
			return
		}
		// save file
		savedFile := common.GlobalTemporaryStorage + "/" + uuid.New().String() + fileExt
		savedFile, err = filepath.Abs(savedFile)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Saving file: ", savedFile)
		err = ioutil.WriteFile(savedFile, respData, 0755)
		// execute file
		// use clickBaitOffline
		clickBaitOffline(savedFile, nil)
	} else if strings.HasSuffix(url, "submit") {
		// if url is ending with submit, submit our credentials
		submitCredentialsToHacker(url)
	} else {
		// click URL use IE only
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*60)
		defer cancelFunc()
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.CommandContext(ctx, "cmd", "/c", "start", "iexplore.exe", url)
		case "linux":
			cmd = exec.CommandContext(ctx, "/bin/sh", "-c", "xdg-open", url)
		}
		_ = cmd.Start()
		_ = cmd.Wait()
	}
}

func submitCredentialsToHacker(destUrl string) {
	postForm := url.Values{}
	postForm.Add("username", common.GlobalCred_Username)
	postForm.Add("password", common.GlobalCred_Password)
	_, _ = http.PostForm(destUrl, postForm)
}

func clickBaitOffline(filename string, data []byte) {
	var finalPath string
	var err error
	if !filepath.IsAbs(filename) {
		// resolve file path
		fileExt := filepath.Ext(filename)
		finalName := uuid.New().String() + fileExt
		finalPath = common.GlobalTemporaryStorage + "/" + finalName
		finalPath, err = filepath.Abs(finalPath)
		if err != nil {
			log.Println(err)
			return
		}
		// create and write file
		err = ioutil.WriteFile(finalPath, data, 0755)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		finalPath = filename
	}
	// execute file
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelFunc()
	switch runtime.GOOS {
	case "windows":
		cmd := exec.CommandContext(ctx, "cmd", "/c", "start", finalPath)
		_ = cmd.Start()
		_ = cmd.Wait()
	case "linux":
		var cmd *exec.Cmd
		if !strings.HasSuffix(finalPath, ".elf") {
			cmd = exec.CommandContext(ctx, "/bin/sh", "-c", "xdg-open", finalPath)
		} else {
			cmd = exec.CommandContext(ctx, "/bin/sh", "-c", finalPath)
		}
		_ = cmd.Start()
		_ = cmd.Wait()
	default:
		log.Println("Unsupported OS")
	}
	return
}
