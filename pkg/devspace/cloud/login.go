package cloud

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/devspace-cloud/devspace/pkg/devspace/cloud/client"
	"github.com/devspace-cloud/devspace/pkg/devspace/cloud/token"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	"github.com/devspace-cloud/devspace/pkg/util/survey"
	"github.com/pkg/errors"
)

// LoginEndpoint is the cloud endpoint that will log you in
const LoginEndpoint = "/login?cli=true"

// LoginSuccessEndpoint is the url redirected to after successful login
const LoginSuccessEndpoint = "/login-success"

// TokenEndpoint is the endpoint where to get a token from
const TokenEndpoint = "/auth/token"

// GetToken returns a valid access token to the provider
func (p *provider) GetToken() (string, error) {
	if p.Key == "" {
		return "", errors.New("Provider has no key specified")
	}
	if p.Token != "" && token.IsTokenValid(p.Token) {
		return p.Token, nil
	}

	resp, err := http.Get(p.Host + TokenEndpoint + "?key=" + p.Key)
	if err != nil {
		return "", errors.Wrap(err, "token request")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "read request body")
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("Error retrieving token: Code %v => %s. Try to relogin with 'devspace login'", resp.StatusCode, string(body))
	}

	p.Token = string(body)
	if token.IsTokenValid(p.Token) == false {
		return "", errors.New("Received invalid token from provider")
	}

	err = p.Save()
	if err != nil {
		return "", errors.Wrap(err, "token save")
	}

	return p.Token, nil
}

// Login logs the user into DevSpace Cloud
func (p *provider) Login() error {
	var (
		url        = p.Host + LoginEndpoint
		ctx        = context.Background()
		keyChannel = make(chan string)
	)
	var key string

	server := startServer(p.Host+LoginSuccessEndpoint, keyChannel, p.log)
	err := p.browser.Run(url)
	if err != nil {
		p.log.Infof("Unable to open web browser for login page.\n\n Please follow these instructions for manually loggin in:\n\n  1. Open this URL in a browser: %s\n  2. After logging in, click the 'Create Key' button\n  3. Enter a key name (e.g. my-key) and click 'Create Access Key'\n  4. Copy the generated key from the input field", p.Host+"/settings/access-keys")

		key, err = p.log.Question(&survey.QuestionOptions{
			Question:   "5. Enter the access key here:",
			IsPassword: true,
		})
		if err != nil {
			close(keyChannel)
			server.Shutdown(ctx)
			return err
		}

		key = strings.TrimSpace(key)

		p.log.WriteString("\n")

		// Check if we got access
		p.Key = key
		_, err := p.client.GetSpaces()
		if err != nil {
			close(keyChannel)
			server.Shutdown(ctx)
			return errors.Wrap(err, "login")
		}
	} else {
		p.log.Infof("If the browser does not open automatically, please navigate to %s", url)
		p.log.StartWait("Logging into cloud provider...")
		defer p.log.StopWait()

		key = <-keyChannel
	}

	close(keyChannel)
	err = server.Shutdown(ctx)
	if err != nil {
		return err
	}

	p.Key = key
	p.client = client.NewClient(p.Name, p.Host, key, p.Token)
	return nil
}

func startServer(redirectURI string, keyChannel chan string, log log.Logger) *http.Server {
	srv := &http.Server{Addr: ":25853"}

	http.HandleFunc("/key", func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["key"]
		if !ok || len(keys[0]) < 1 {
			log.Warn("Bad request")
			return
		}

		keyChannel <- keys[0]
		http.Redirect(w, r, redirectURI, http.StatusSeeOther)
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}
