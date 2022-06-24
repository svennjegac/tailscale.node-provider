package state

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/svennjegac/tailscale.node-provider/internal/fileutil"
	"github.com/svennjegac/tailscale.node-provider/internal/tscos"
)

type State struct {
	Nodes  map[int]*VPNNode `json:"nodes"`
	LastID int              `json:"last_id"`
}

type VPNNode struct {
	TscalectlID   int    `json:"tscalectl_id"`
	TscalectlName string `json:"tscalectl_name"`

	Region       string `json:"region"`
	InstanceType string `json:"instance_type"`
	AMI          string `json:"ami"`
}

func AddNewNode(region string, instanceType string, ami string) *VPNNode {
	fileutil.MkdirAll(tscos.TscalectlDir())

	unlock := fileutil.Lock(tscos.StateFile())
	defer unlock()

	s := getState()

	tscalectlID := s.getNextTscalectlID()
	tscalectlIDStr := strconv.Itoa(tscalectlID)

	if tscalectlID < 10 {
		tscalectlIDStr = leftPad(2, "0", tscalectlIDStr)
	} else if tscalectlID < 100 {
		tscalectlIDStr = leftPad(1, "0", tscalectlIDStr)
	}

	node := &VPNNode{
		TscalectlID:   tscalectlID,
		TscalectlName: fmt.Sprintf("%s-%s-%s", tscalectlIDStr, region, instanceType),
		Region:        region,
		InstanceType:  instanceType,
		AMI:           ami,
	}
	s.Nodes[tscalectlID] = node

	storeState(s)

	return node
}

func GetNode(tscalectlID int) *VPNNode {
	fileutil.MkdirAll(tscos.TscalectlDir())

	unlock := fileutil.Lock(tscos.StateFile())
	defer unlock()

	s := getState()

	node, ok := s.Nodes[tscalectlID]
	if !ok {
		panic(errors.New("node with provided ID does not exist in CLI state"))
	}

	return node
}

func RemoveNode(tscalectlID int) {
	fileutil.MkdirAll(tscos.TscalectlDir())

	unlock := fileutil.Lock(tscos.StateFile())
	defer unlock()

	s := getState()

	delete(s.Nodes, tscalectlID)

	storeState(s)
}

func GetState() *State {
	fileutil.MkdirAll(tscos.TscalectlDir())

	unlock := fileutil.Lock(tscos.StateFile())
	defer unlock()

	return getState()
}

func (s *State) getNextTscalectlID() int {
	id := s.LastID
	s.LastID = (s.LastID + 1) % 1000
	return id
}

func leftPad(pads int, padChars string, s string) string {
	return strings.Repeat(padChars, pads) + s
}

func getState() *State {
	if _, err := os.Stat(tscos.StateFile()); os.IsNotExist(err) {
		return &State{Nodes: make(map[int]*VPNNode)}
	} else if err != nil {
		panic(errors.Wrap(err, "get state, os stat"))
	}

	b := fileutil.ReadFile(tscos.StateFile())

	var state State
	err := json.Unmarshal(b, &state)
	if err != nil {
		panic(errors.Wrap(err, "get state, json unmarshal"))
	}

	return &state
}

func storeState(s *State) {
	b, err := json.Marshal(s)
	if err != nil {
		panic(errors.Wrap(err, "store state json marshal"))
	}

	fileutil.WriteFile(tscos.StateFile(), b)
}
