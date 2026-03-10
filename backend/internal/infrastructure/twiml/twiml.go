// Package twiml provides a builder for Twilio Markup Language (TwiML) XML responses.
package twiml

import (
	"encoding/xml"
	"fmt"
)

// Node is the interface that all TwiML elements implement.
type Node interface {
	// xmlElement returns the XML-encodable representation of the node.
	xmlElement() interface{}
}

// Say represents a TwiML <Say> verb that speaks text to the caller.
type Say struct {
	// Text is the text to speak.
	Text string
	// Voice is the voice to use (e.g., "alice", "man", "woman").
	Voice string
	// Language is the language code (e.g., "en-US").
	Language string
}

// xmlSay is the XML-serializable representation of the Say verb.
type xmlSay struct {
	XMLName  xml.Name `xml:"Say"`
	Voice    string   `xml:"voice,attr,omitempty"`
	Language string   `xml:"language,attr,omitempty"`
	Text     string   `xml:",chardata"`
}

func (s Say) xmlElement() interface{} {
	return xmlSay{
		Voice:    s.Voice,
		Language: s.Language,
		Text:     s.Text,
	}
}

// Play represents a TwiML <Play> verb that plays an audio file.
type Play struct {
	// URL is the URL of the audio file to play.
	URL string
}

// xmlPlay is the XML-serializable representation of the Play verb.
type xmlPlay struct {
	XMLName xml.Name `xml:"Play"`
	URL     string   `xml:",chardata"`
}

func (p Play) xmlElement() interface{} {
	return xmlPlay{URL: p.URL}
}

// Gather represents a TwiML <Gather> verb that collects user input.
type Gather struct {
	// Input is the type of input to collect (e.g., "speech", "dtmf", "speech dtmf").
	Input string
	// Action is the URL to submit gathered input to.
	Action string
	// Method is the HTTP method for the action URL.
	Method string
	// Timeout is the seconds to wait for input.
	Timeout int
	// SpeechTimeout is the seconds of silence before speech input ends.
	SpeechTimeout string
	// Language is the language for speech recognition.
	Language string
	// Children are nested TwiML nodes inside the Gather (e.g., Say, Play).
	Children []Node
}

// xmlGather is the XML-serializable representation of the Gather verb.
type xmlGather struct {
	XMLName       xml.Name `xml:"Gather"`
	Input         string   `xml:"input,attr,omitempty"`
	Action        string   `xml:"action,attr,omitempty"`
	Method        string   `xml:"method,attr,omitempty"`
	Timeout       int      `xml:"timeout,attr,omitempty"`
	SpeechTimeout string   `xml:"speechTimeout,attr,omitempty"`
	Language      string   `xml:"language,attr,omitempty"`
	Children      []interface{}
}

func (g Gather) xmlElement() interface{} {
	children := make([]interface{}, 0, len(g.Children))
	for _, c := range g.Children {
		children = append(children, c.xmlElement())
	}
	return xmlGather{
		Input:         g.Input,
		Action:        g.Action,
		Method:        g.Method,
		Timeout:       g.Timeout,
		SpeechTimeout: g.SpeechTimeout,
		Language:      g.Language,
		Children:      children,
	}
}

// Redirect represents a TwiML <Redirect> verb that transfers control.
type Redirect struct {
	// URL is the URL to redirect to.
	URL string
	// Method is the HTTP method for the redirect URL.
	Method string
}

// xmlRedirect is the XML-serializable representation of the Redirect verb.
type xmlRedirect struct {
	XMLName xml.Name `xml:"Redirect"`
	Method  string   `xml:"method,attr,omitempty"`
	URL     string   `xml:",chardata"`
}

func (r Redirect) xmlElement() interface{} {
	return xmlRedirect{Method: r.Method, URL: r.URL}
}

// Hangup represents a TwiML <Hangup> verb that ends the call.
type Hangup struct{}

// xmlHangup is the XML-serializable representation of the Hangup verb.
type xmlHangup struct {
	XMLName xml.Name `xml:"Hangup"`
}

func (h Hangup) xmlElement() interface{} {
	return xmlHangup{}
}

// Pause represents a TwiML <Pause> verb that waits silently.
type Pause struct {
	// Length is the number of seconds to pause.
	Length int
}

// xmlPause is the XML-serializable representation of the Pause verb.
type xmlPause struct {
	XMLName xml.Name `xml:"Pause"`
	Length  int      `xml:"length,attr,omitempty"`
}

func (p Pause) xmlElement() interface{} {
	return xmlPause{Length: p.Length}
}

// xmlResponse is the top-level TwiML Response element.
type xmlResponse struct {
	XMLName  xml.Name      `xml:"Response"`
	Children []interface{}
}

// Build takes a slice of TwiML nodes and produces the complete XML string
// with the XML declaration and <Response> wrapper.
func Build(nodes []Node) (string, error) {
	children := make([]interface{}, 0, len(nodes))
	for _, n := range nodes {
		children = append(children, n.xmlElement())
	}

	resp := xmlResponse{Children: children}
	data, err := xml.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal twiml: %w", err)
	}

	return xml.Header + string(data), nil
}
