package endpoints

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"text/template"

	"github.com/gofrs/uuid"
	"github.com/sahib/config"
	"github.com/sahib/wishlist/cache"
	"github.com/sahib/wishlist/db"
)

const (
	EMailTemplateText = `

Hallo!
<br />
<br />
Du bekommst diese Mail weil du dich auf der Geschenkeliste angemeldet hast. <br />
Wenn du nicht weißt warum du diese EMail bekommst, ignoriere sie bitte einfach.
<br />
<br />
Bitte klicke auf diesen Link, damit wir wirklich wissen, dass du du bist:
<br />
<br />
	<a href="https://{{.Domain}}/api/v0/token/{{.SessionID}}">https://{{.Domain}}/api/v0/token/{{.SessionID}}</a>
<br />
<br />
Danke!
<br />
<br />
P.S: Du hast dich übrigens mit folgenden Namen angemeldet: <i>{{.Name}}</i>
`
)

var EMailTemplate *template.Template

func init() {
	var err error
	EMailTemplate, err = template.New("register-email").Parse(EMailTemplateText)
	if err != nil {
		log.Fatalf("failed to parse internal template: %v", err)
	}
}

type LoginRequest struct {
	Name  string `json="name"`
	Email string `json="email"`
}

type LoginResponse struct {
	Success           bool `json="success"`
	IsAlreadyLoggedIn bool `json="is_already_logged_in"`
}

type LoginHandler struct {
	db    *db.Database
	cache *cache.SessionCache
	cfg   *config.Config
}

func NewLoginHandler(db *db.Database, cache *cache.SessionCache, cfg *config.Config) *LoginHandler {
	return &LoginHandler{db: db, cache: cache, cfg: cfg}
}

func (lh *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json in request body: %v", err)
		return
	}

	req.Name = html.EscapeString(req.Name)
	req.Email = html.EscapeString(req.Email)

	user, err := IsAuthenticated(r, lh.cache, lh.db)
	if user != nil && err == nil {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			jsonifyErrf(w, http.StatusInternalServerError, "delete your cookies")
			return
		}

		if user.EMail == req.Email {
			// The email did not change and we are logged in already.
			// Set the cookie again.
			jsonify(w, http.StatusOK, &LoginResponse{
				Success:           true,
				IsAlreadyLoggedIn: true,
			})
			return
		}

		// Best to clean up the old session and make the user log in properly again:
		// (he might want to change accounts for whatever reason)
		if err := lh.cache.Forget(cookie.Value); err != nil {
			jsonifyErrf(w, http.StatusInternalServerError, "forget failed: %v", err)
			return
		}
	}

	sessionID, err := uuid.NewV4()
	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to gen uuid: %v", err)
		return
	}

	userID, err := lh.getOrAddUserID(req)
	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to add/get user: %v", err)
		return
	}

	if err := lh.sendLoginlMail(&req, sessionID); err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to send login mail: %v", err)
		return
	}

	if err := lh.cache.Remember(userID, sessionID.String()); err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to remember session: %v", err)
	}

	log.Printf("new login request; sending mail to %s", req.Email)

	jsonify(w, http.StatusOK, &LoginResponse{
		Success:           true,
		IsAlreadyLoggedIn: false,
	})
}

func (lh *LoginHandler) NeedsAuthentication() bool {
	return false
}

func (lh *LoginHandler) getOrAddUserID(req LoginRequest) (int64, error) {
	user, err := lh.db.GetUserByEMail(req.Email)
	if err != nil {
		return -1, err
	}

	if user == nil {
		userID, err := lh.db.AddUser(req.Name, req.Email)
		if err != nil {
			return -1, err
		}

		return userID, nil
	}

	return user.ID, nil
}

func (lh *LoginHandler) sendLoginlMail(req *LoginRequest, uuid uuid.UUID) error {
	domain := lh.cfg.String("server.domain")
	if port := lh.cfg.Int("server.port"); port != 80 && port != 443 {
		domain = fmt.Sprintf("%s:%d", domain, port)
	}

	type TemplateVars struct {
		Name      string
		Domain    string
		SessionID string
	}

	buf := &bytes.Buffer{}
	vars := TemplateVars{
		Name:      req.Name,
		SessionID: uuid.String(),
		Domain:    domain,
	}

	if err := EMailTemplate.Execute(buf, vars); err != nil {
		return err
	}

	return sendMail(
		lh.cfg.String("mail.from"),
		req.Email,
		"Login zur Geschenkliste",
		buf.String(),
		fmt.Sprintf("%s:%d", lh.cfg.String("mail.smtp_host"), lh.cfg.Int("mail.smtp_port")),
		lh.cfg.String("mail.smtp_password"),
	)
}

func sendMail(from, to, subj, body, servername, password string) error {
	// Stolen from: https://gist.github.com/chrisgillis/10888032
	fromAddr := mail.Address{"", from}
	toAddr := mail.Address{"", to}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = fromAddr.String()
	headers["To"] = toAddr.String()
	headers["Subject"] = subj
	headers["MIME-version"] = "1.0;"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\";"

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	host, _, _ := net.SplitHostPort(servername)
	auth := smtp.PlainAuth("", from, password, host)

	// TLS config
	tlsconfig := &tls.Config{
		ServerName: host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(fromAddr.Address); err != nil {
		return err
	}

	if err = c.Rcpt(toAddr.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}
