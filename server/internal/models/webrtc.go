package models

type SDP struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

type ICECandidate struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
}
