package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

type smtpSender struct {
	host, address, security string
	from, user, password    string
}

func loadSMTPSender() (*smtpSender, error) {
	host := os.Getenv("AUTH_SMTP_HOST")
	if host == "" {
		return nil, nil
	}
	sender := &smtpSender{
		host: host, from: os.Getenv("AUTH_SMTP_FROM"),
		security: os.Getenv("AUTH_SMTP_SECURITY"),
		user:     os.Getenv("AUTH_SMTP_USER"), password: os.Getenv("AUTH_SMTP_PASSWORD"),
	}
	if sender.security == "" {
		sender.security = "starttls"
	}
	switch sender.security {
	case "starttls", "tls":
	case "plain":
		if sender.user != "" || sender.password != "" {
			return nil, errors.New("SMTP credentials require TLS")
		}
	default:
		return nil, errors.New("invalid AUTH_SMTP_SECURITY")
	}
	if (sender.user == "") != (sender.password == "") {
		return nil, errors.New("configure both SMTP user and password")
	}
	from, err := mail.ParseAddress(sender.from)
	if err != nil || from.Address != sender.from {
		return nil, errors.New("AUTH_SMTP_FROM must contain a single email address")
	}
	port := os.Getenv("AUTH_SMTP_PORT")
	if port == "" {
		port = "587"
		if sender.security == "tls" {
			port = "465"
		}
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return nil, errors.New("invalid AUTH_SMTP_PORT")
	}
	sender.address = net.JoinHostPort(sender.host, port)
	return sender, nil
}

func (s *smtpSender) sendCode(ctx context.Context, to, code string, purpose emailPurpose) error {
	ctx, cancel := context.WithTimeout(ctx, 7*time.Second)
	defer cancel()
	deadline, _ := ctx.Deadline()
	tlsConfig := &tls.Config{ServerName: s.host, MinVersion: tls.VersionTLS12}
	dialer := &net.Dialer{Deadline: deadline}
	var conn net.Conn
	var err error
	if s.security == "tls" {
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: tlsConfig}).DialContext(ctx, "tcp", s.address)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", s.address)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetDeadline(deadline); err != nil {
		return err
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()

	var client *smtp.Client
	if s.security == "starttls" {
		client, err = smtp.NewClientStartTLS(conn, tlsConfig)
		if err != nil {
			return err
		}
	} else {
		client = smtp.NewClient(conn)
	}
	if s.user != "" {
		if err := client.Auth(sasl.NewPlainClient("", s.user, s.password)); err != nil {
			return err
		}
	}
	if err := client.Mail(s.from, nil); err != nil {
		return err
	}
	if err := client.Rcpt(to, nil); err != nil {
		return err
	}
	body, err := client.Data()
	if err != nil {
		return err
	}
	subject := mime.QEncoding.Encode("UTF-8", "AI Dungeon Master — "+purpose.caption())
	_, err = fmt.Fprintf(body, "From: AI Dungeon Master <%s>\r\nTo: %s\r\nSubject: %s\r\n"+
		"MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n"+
		"Content-Transfer-Encoding: base64\r\n\r\n%s\r\n",
		s.from, to, subject, emailCodeBody(code))
	if err != nil {
		return err
	}
	return body.Close()
}

func emailCodeBody(code string) string {
	text := "Код подтверждения: " + code + "\r\n\r\nКод действует 10 минут. Если вы не запрашивали код, проигнорируйте письмо.\r\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	var body strings.Builder
	for len(encoded) > 76 {
		body.WriteString(encoded[:76] + "\r\n")
		encoded = encoded[76:]
	}
	body.WriteString(encoded)
	return body.String()
}
