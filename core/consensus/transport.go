// Copyright © 2022-2023 Obol Labs Inc. Licensed under the terms of a Business Source License 1.1

package consensus

import (
	"context"
	"crypto/rand"
	"sync"
	"time"

	k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/obolnetwork/charon/app/errors"
	"github.com/obolnetwork/charon/app/log"
	"github.com/obolnetwork/charon/core"
	pbv1 "github.com/obolnetwork/charon/core/corepb/v1"
	"github.com/obolnetwork/charon/core/qbft"
)

// transport encapsulates receiving and broadcasting for a consensus instance/duty.
type transport struct {
	// Immutable state
	component  *Component
	recvBuffer chan qbft.Msg[core.Duty, [32]byte] // Instance inner receive buffer.
	sniffer    *sniffer

	// Mutable state
	valueMu sync.Mutex
	values  map[[32]byte]*anypb.Any // maps any-wrapped proposed values to their hashes
}

// setValues caches the values and their hashes.
func (t *transport) setValues(msg msg) {
	t.valueMu.Lock()
	defer t.valueMu.Unlock()

	for k, v := range msg.values {
		t.values[k] = v
	}
}

// getValue returns the value by its hash.
func (t *transport) getValue(hash [32]byte) (*anypb.Any, error) {
	t.valueMu.Lock()
	defer t.valueMu.Unlock()

	pb, ok := t.values[hash]
	if !ok {
		return nil, errors.New("unknown value")
	}

	return pb, nil
}

// usePointerValues returns true if the transport should use pointer values in the message instead of the legacy
// duplicated values in QBFTMsg.
func (t *transport) usePointerValues() bool {
	// Equivalent to math/rand.Float64() just with less precision.
	b := make([]byte, 1)
	_, _ = rand.Read(b)
	f := float64(b[0]) / 255

	return f >= t.component.legacyProbability
}

// Broadcast creates a msg and sends it to all peers (including self).
func (t *transport) Broadcast(ctx context.Context, typ qbft.MsgType, duty core.Duty,
	peerIdx int64, round int64, valueHash [32]byte, pr int64, pvHash [32]byte,
	justification []qbft.Msg[core.Duty, [32]byte],
) error {
	// Get the values by their hashes if not zero.
	var (
		value *anypb.Any
		pv    *anypb.Any
		err   error
	)

	if valueHash != [32]byte{} {
		value, err = t.getValue(valueHash)
		if err != nil {
			return err
		}
	}

	if pvHash != [32]byte{} {
		pv, err = t.getValue(pvHash)
		if err != nil {
			return err
		}
	}

	// Make the message
	msg, err := createMsg(typ, duty, peerIdx, round, valueHash, value, pr, pvHash, pv, justification, t.component.privkey, t.usePointerValues())
	if err != nil {
		return err
	}

	// Send to self (async since buffer is blocking).
	go func() {
		select {
		case <-ctx.Done():
		case t.recvBuffer <- msg:
			t.sniffer.Add(msg.ToConsensusMsg())
		}
	}()

	for _, p := range t.component.peers {
		if p.ID == t.component.tcpNode.ID() {
			// Do not broadcast to self
			continue
		}

		err = t.component.sender.SendAsync(ctx, t.component.tcpNode, protocolID, p.ID, msg.ToConsensusMsg())
		if err != nil {
			return err
		}
	}

	return nil
}

// ProcessReceives processes received messages from the outer buffer until the context is closed.
//

func (t *transport) ProcessReceives(ctx context.Context, outerBuffer chan msg) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-outerBuffer:
			if err := validateMsg(msg); err != nil {
				log.Warn(ctx, "Dropping invalid message", err)
				continue
			}

			t.setValues(msg)

			select {
			case <-ctx.Done():
				return
			case t.recvBuffer <- msg:
				t.sniffer.Add(msg.ToConsensusMsg())
			}
		}
	}
}

// createMsg returns a new message by converting the inputs into a protobuf
// and wrapping that in a msg type.
func createMsg(typ qbft.MsgType, duty core.Duty,
	peerIdx int64, round int64,
	vHash [32]byte, value *anypb.Any,
	pr int64,
	pvHash [32]byte, pv *anypb.Any,
	justification []qbft.Msg[core.Duty, [32]byte], privkey *k1.PrivateKey,
	pointerValues bool,
) (msg, error) {
	values := make(map[[32]byte]*anypb.Any)
	if value != nil {
		values[vHash] = value
	}
	if pv != nil {
		values[pvHash] = pv
	}

	// Disable new pointer values, revert to legacy duplicated values.
	if !pointerValues {
		values = make(map[[32]byte]*anypb.Any)
		vHash = [32]byte{}
		pvHash = [32]byte{}
	}

	pbMsg := &pbv1.QBFTMsg{
		Type:              int64(typ),
		Duty:              core.DutyToProto(duty),
		PeerIdx:           peerIdx,
		Round:             round,
		Value:             value,
		ValueHash:         vHash[:],
		PreparedRound:     pr,
		PreparedValue:     pv,
		PreparedValueHash: pvHash[:],
	}

	pbMsg, err := signMsg(pbMsg, privkey)
	if err != nil {
		return msg{}, err
	}

	// Transform justifications into protobufs
	var justMsgs []*pbv1.QBFTMsg
	for _, j := range justification {
		impl, ok := j.(msg)
		if !ok {
			return msg{}, errors.New("invalid justification")
		}
		justMsgs = append(justMsgs, impl.msg) // Note nested justifications are ignored.
	}

	return newMsg(pbMsg, justMsgs, values)
}

// validateMsg returns an error if the message is invalid.
func validateMsg(_ msg) error {
	// TODO(corver): implement (incl signature verification).
	return nil
}

// newSniffer returns a new sniffer.
func newSniffer(nodes, peerIdx int64) *sniffer {
	return &sniffer{
		nodes:     nodes,
		peerIdx:   peerIdx,
		startedAt: time.Now(),
	}
}

// sniffer buffers consensus messages.
type sniffer struct {
	nodes     int64
	peerIdx   int64
	startedAt time.Time

	mu   sync.Mutex
	msgs []*pbv1.SniffedConsensusMsg
}

// Add adds a message to the sniffer buffer.
func (c *sniffer) Add(msg *pbv1.ConsensusMsg) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.msgs = append(c.msgs, &pbv1.SniffedConsensusMsg{
		Timestamp: timestamppb.Now(),
		Msg:       msg,
	})
}

// Instance returns the buffered messages as an instance.
func (c *sniffer) Instance() *pbv1.SniffedConsensusInstance {
	c.mu.Lock()
	defer c.mu.Unlock()

	return &pbv1.SniffedConsensusInstance{
		Nodes:     c.nodes,
		PeerIdx:   c.peerIdx,
		StartedAt: timestamppb.New(c.startedAt),
		Msgs:      c.msgs,
	}
}
