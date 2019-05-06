package warmed

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// DefaultKeepAlive is the normal keep alive interval a warmed client uses.
const DefaultKeepAlive = 30 * time.Second

// Client is the warmed client wrapper around a normal http.Client (using
// composition) that has a goroutine running to keeps target host in the
// keep-alive connection pool by calling them with a HEAD reqeust periodically.
//
// Use New() to create such a Client:
//  client := warmed.New("https://example.org")
//
// Target hosts/urls can be set by passing them on creation, using the Target()
// method or by using the client to do http calls, those urls will become
// targets automatically.
type Client struct {
	http.Client

	transport *http.Transport
	targets   map[string]struct{}
	lock      *sync.Mutex

	close  chan struct{}
	closed bool

	KeepAlive time.Duration
}

// New creates a new warmed.Client and starts the goroutine that keeps
// the target connections alive. Optionally you can pass target urls
// on this creation otherwise just calling hosts using this client will
// make the targets.
func New(targets ...string) *Client {
	client := &Client{
		targets: make(map[string]struct{}),
		close:   make(chan struct{}),
		lock:    &sync.Mutex{},

		KeepAlive: DefaultKeepAlive,
	}

	spy := func(req *http.Request) (*url.URL, error) {
		client.Target(req.URL.String())
		return http.ProxyFromEnvironment(req)
	}

	dialer := &net.Dialer{
		Timeout:   client.KeepAlive,
		KeepAlive: client.KeepAlive,
	}

	client.transport = &http.Transport{
		MaxIdleConnsPerHost: 1024,
		Dial:                dialer.Dial,
		Proxy:               spy,
	}
	client.Transport = client.transport

	if len(targets) > 0 {
		client.Target(targets...)
		client.touchTargets()
	}

	go client.keepalive()
	return client
}

// Close will shutdown this keepalive process and a normal
// http.Client remains.
func (c *Client) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.closed {
		c.closed = true
		close(c.close)
		c.transport.CloseIdleConnections()
	}
}

// Target will add one or more new targets to keep alive.
func (c *Client) Target(targets ...string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, target := range targets {
		u, err := url.Parse(target)
		if err == nil && u.Scheme != "" {
			base, _ := u.Parse("/")
			c.targets[base.String()] = struct{}{}
		}
	}
}

// Targets returns the list of urls/hosts that are kept
// alive.
func (c *Client) Targets() []string {
	c.lock.Lock()
	defer c.lock.Unlock()

	var targets []string
	for h := range c.targets {
		targets = append(targets, h)
	}
	return targets
}

func (c *Client) keepalive() {
	for {
		select {
		case <-time.NewTimer(c.KeepAlive / 10 * 9).C:
			c.touchTargets()
		case <-c.close:
			break
		}
	}
}

func (c *Client) touchTargets() {
	for _, target := range c.Targets() {
		req, _ := http.NewRequest(http.MethodHead, target, nil) // ignore err, url already parsed
		resp, err := c.Do(req)
		if err != nil {
			continue
		}
		io.Copy(ioutil.Discard, resp.Body) // nolint: errcheck
		resp.Body.Close()
	}
}
