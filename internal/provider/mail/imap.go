package mail

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Pauloo27/mail-notifier/internal/provider"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

var _ provider.MailProvider = Mail{}

type Mail struct {
	Host, Username, Password string
	Port                     int

	client *client.Client
}

func init() {
	provider.Factories["imap"] = func(info map[string]interface{}) (provider.MailProvider, error) {
		return NewMail(info["host"].(string), int(info["port"].(float64)), info["username"].(string), info["password"].(string))
	}
}

func (m *Mail) Connect() error {
	c, err := client.DialTLS(fmt.Sprintf("%s:%d", m.Host, m.Port), nil)
	m.client = c
	return err
}

func (m Mail) GetAddress() string {
	return m.Username // FIXME: the username is always not the complete address,
	// maybe i can  get it from imap?
}

func NewMail(host string, port int, username, password string) (Mail, error) {
	m := Mail{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
	err := m.Connect()
	if err == nil {
		err = m.client.Login(username, password)
	}
	return m, err
}

func (m *Mail) Disconnect() error {
	return m.client.Logout()
}

func (m Mail) FetchMessage(id string) (message provider.MailMessage, err error) {
	seq := new(imap.SeqSet)
	seq.Add(id)

	msgCh := make(chan *imap.Message, 1)

	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	err = m.client.Fetch(seq, items, msgCh)
	if err != nil {
		return
	}

	msg := <-msgCh
	if msg == nil {
		err = errors.New("not found")
		return
	}

	body := msg.GetBody(&section)

	var mr *mail.Reader
	mr, err = mail.CreateReader(body)
	if err != nil {
		return
	}

	var date time.Time
	var subject, from string
	var to []string

	var addrs []*mail.Address

	header := mr.Header
	date, err = header.Date()
	if err != nil {
		return
	}

	subject, err = header.Subject()
	if err != nil {
		return
	}
	fmt.Println("the subject is", subject)

	addrs, err = header.AddressList("From")
	if err != nil {
		return
	}
	from = addrs[0].String()

	addrs, err = header.AddressList("To")
	if err != nil {
		return
	}

	for _, add := range addrs {
		to = append(to, add.String())
	}

	message = MailMessage{
		id:      id,
		date:    date,
		from:    from,
		to:      to,
		subject: subject,
		loaded:  true,
	}

	return
}

func (m Mail) FetchMessages(onlyUnread bool) (messages []provider.MailMessage, count int, err error) {
	criteria := imap.NewSearchCriteria()

	if onlyUnread {
		criteria.WithoutFlags = []string{imap.SeenFlag}
	}

	_, err = m.client.Select("INBOX", true)
	if err != nil {
		return
	}

	criteria.Since = time.Now().AddDate(-1, 0, 0) // limit search on 1 year

	var rawIDs []uint32
	rawIDs, err = m.client.Search(criteria)

	for _, id := range rawIDs {
		messages = append(messages, MailMessage{id: strconv.Itoa(int(id)), loaded: false, mail: &m})
	}

	count = len(messages)

	return
}