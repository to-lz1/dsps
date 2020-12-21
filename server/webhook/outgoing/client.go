package outgoing

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
	"golang.org/x/xerrors"
)

// Client is an outgoing-webhook client.
type Client interface {
	Send(ctx context.Context, msg domain.Message) error

	// Shutdown this client.
	// This method wait until all in-flight request ends.
	Close(ctx context.Context)

	String() string
}

type clientImpl struct {
	_isClosed int32 // 0: available, 1: closing

	method  string
	url     string
	headers map[string]string

	timeout time.Duration
	retry   retry

	// Note that Client does not own this object, ClientTemplate owns.
	h *http.Client
}

func newClientImpl(tpl *clientTemplate, tplEnv domain.TemplateStringEnv) (Client, error) {
	c := &clientImpl{
		_isClosed: 0,

		method:  tpl.Method,
		headers: make(map[string]string, len(tpl.Headers)),

		timeout: tpl.Timeout.Duration,
		retry:   newRetry(&tpl.Retry),

		h: tpl.h,
	}

	var err error
	c.url, err = tpl.URL.Execute(tplEnv)
	if err != nil {
		return nil, xerrors.Errorf(`failed to expand template of webhook URL "%s": %w`, tpl.URL, err)
	}
	for name, valueTpl := range tpl.Headers {
		c.headers[name], err = valueTpl.Execute(tplEnv)
		if err != nil {
			return nil, xerrors.Errorf(`failed to expand template of webhook header "%s", "%s": %w`, name, valueTpl, err)
		}
	}
	return c, nil
}

func (c *clientImpl) String() string {
	return fmt.Sprintf("%s %s", c.method, c.url)
}

func (c *clientImpl) Send(ctx context.Context, msg domain.Message) error {
	if c.isClosed() {
		return xerrors.Errorf("outgoing-webhook client already closed")
	}
	logger.Of(ctx).Debugf(logger.CatOutgoingWebhook, "sending outgoing webhook (channel: %s, messageID: %s) to %s", msg.ChannelID, msg.MessageID, c.url)

	body, err := encodeWebhookBody(ctx, msg)
	if err != nil {
		return xerrors.Errorf("failed to generate outgoing webhook body: %w", err)
	}

	return c.retry.Do(ctx, fmt.Sprintf("outgoing-webhook to %s", c.url), func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, c.method, c.url, strings.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		for name, value := range c.headers {
			// Should overwrite default headers, thus use Set() rather than Add()
			req.Header.Set(name, value)
		}

		res, err := c.h.Do(req)
		if res != nil {
			logger.Of(ctx).Debugf(logger.CatOutgoingWebhook, "received outgoing webhook response (%s %d, contentLength: %d)", res.Proto, res.StatusCode, res.ContentLength)
		}
		return res, err
	})
}

func (c *clientImpl) Close(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&c._isClosed, 0, 1) {
		return
	}
}

func (c *clientImpl) isClosed() bool {
	return atomic.LoadInt32(&c._isClosed) != 0
}