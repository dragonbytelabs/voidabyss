package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
)

// Client represents an LSP client connected to a language server
type Client struct {
	cmd            *exec.Cmd
	stdin          io.WriteCloser
	stdout         io.ReadCloser
	reader         *bufio.Reader
	mu             sync.Mutex
	nextID         int32
	pending        map[int]chan *Response
	onNotification func(method string, params json.RawMessage)
	initialized    bool
	capabilities   ServerCapabilities
	rootURI        string
}

// NewClient creates a new LSP client for the given language server command
func NewClient(serverCmd string, args []string, rootURI string) (*Client, error) {
	cmd := exec.Command(serverCmd, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start server: %w", err)
	}

	client := &Client{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		reader:  bufio.NewReader(stdout),
		pending: make(map[int]chan *Response),
		rootURI: rootURI,
	}

	// Start reading responses in background
	go client.readLoop()

	return client, nil
}

// Initialize sends the initialize request and waits for response
func (c *Client) Initialize() error {
	params := InitializeParams{
		ProcessID: nil, // null means don't kill on process exit
		RootURI:   c.rootURI,
		Capabilities: ClientCapabilities{
			TextDocument: TextDocumentClientCapabilities{
				Definition: &DefinitionClientCapabilities{
					DynamicRegistration: false,
					LinkSupport:         true,
				},
			},
		},
	}

	var result InitializeResult
	if err := c.Call("initialize", params, &result); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	c.capabilities = result.Capabilities
	c.initialized = true

	// Send initialized notification
	if err := c.Notify("initialized", struct{}{}); err != nil {
		return fmt.Errorf("initialized notification: %w", err)
	}

	return nil
}

// Call sends a request and waits for the response
func (c *Client) Call(method string, params interface{}, result interface{}) error {
	id := int(atomic.AddInt32(&c.nextID, 1))

	req := &Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	respChan := make(chan *Response, 1)
	c.mu.Lock()
	c.pending[id] = respChan
	c.mu.Unlock()

	if err := c.send(req); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return err
	}

	resp := <-respChan

	if resp.Error != nil {
		return fmt.Errorf("rpc error: %s", resp.Error.Message)
	}

	if result != nil && resp.Result != nil {
		return json.Unmarshal(*resp.Result, result)
	}

	return nil
}

// Notify sends a notification (no response expected)
func (c *Client) Notify(method string, params interface{}) error {
	var rawParams json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		rawParams = data
	}

	notif := &Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  rawParams,
	}
	return c.send(notif)
}

// send writes a message to the server
func (c *Client) send(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := c.stdin.Write([]byte(header)); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	if _, err := c.stdin.Write(data); err != nil {
		return fmt.Errorf("write data: %w", err)
	}

	return nil
}

// readLoop reads messages from the server
func (c *Client) readLoop() {
	for {
		msg, err := c.readMessage()
		if err != nil {
			if err != io.EOF {
				// Server disconnected or error
			}
			return
		}
		c.handleMessage(msg)
	}
}

// readMessage reads a single LSP message (header + content)
func (c *Client) readMessage() (json.RawMessage, error) {
	// Read headers
	var contentLength int
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if line == "\r\n" {
			break // End of headers
		}
		var key, value string
		if _, err := fmt.Sscanf(line, "%s %d", &key, &value); err == nil {
			if key == "Content-Length:" {
				fmt.Sscanf(value, "%d", &contentLength)
			}
		}
	}

	// Read content
	content := make([]byte, contentLength)
	if _, err := io.ReadFull(c.reader, content); err != nil {
		return nil, err
	}

	return content, nil
}

// handleMessage processes a message from the server
func (c *Client) handleMessage(data json.RawMessage) {
	// Try to parse as response first
	var resp Response
	if err := json.Unmarshal(data, &resp); err == nil && resp.ID != nil {
		c.mu.Lock()
		ch, ok := c.pending[*resp.ID]
		if ok {
			delete(c.pending, *resp.ID)
		}
		c.mu.Unlock()

		if ok {
			ch <- &resp
			close(ch)
		}
		return
	}

	// Parse as notification
	var notif Notification
	if err := json.Unmarshal(data, &notif); err == nil {
		if c.onNotification != nil {
			c.onNotification(notif.Method, notif.Params)
		}
	}
}

// Close shuts down the language server
func (c *Client) Close() error {
	if c.initialized {
		_ = c.Notify("shutdown", nil)
		_ = c.Notify("exit", nil)
	}

	c.stdin.Close()
	c.stdout.Close()

	if c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}

	return nil
}
