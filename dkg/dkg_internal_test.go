// Copyright © 2022-2023 Obol Labs Inc. Licensed under the terms of a Business Source License 1.1

package dkg

import (
	"testing"

	eth2p0 "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/stretchr/testify/require"

	"github.com/obolnetwork/charon/core"
	"github.com/obolnetwork/charon/eth2util"
	"github.com/obolnetwork/charon/eth2util/deposit"
	tblsv2 "github.com/obolnetwork/charon/tbls/v2"
	tblsconv2 "github.com/obolnetwork/charon/tbls/v2/tblsconv"
	"github.com/obolnetwork/charon/testutil"
)

func TestInvalidSignatures(t *testing.T) {
	const (
		n  = 4
		th = 3
	)

	secret, err := tblsv2.GenerateSecretKey()
	require.NoError(t, err)

	pubkey, err := tblsv2.SecretToPublicKey(secret)
	require.NoError(t, err)

	secretShares, err := tblsv2.ThresholdSplit(secret, n, th)
	require.NoError(t, err)

	pubshares := make(map[int]tblsv2.PublicKey)

	for idx, share := range secretShares {
		pubkey, err := tblsv2.SecretToPublicKey(share)
		require.NoError(t, err)

		pubshares[idx] = pubkey
	}

	shares := share{
		PubKey:       pubkey,
		SecretShare:  secretShares[0],
		PublicShares: pubshares,
	}

	getSigs := func(msg []byte) []core.ParSignedData {
		var sigs []core.ParSignedData
		for i := 0; i < n-1; i++ {
			sig, err := tblsv2.Sign(secretShares[i+1], msg)
			require.NoError(t, err)

			sigs = append(sigs, core.NewPartialSignature(tblsconv2.SigToCore(sig), i+1))
		}

		invalidSig, err := tblsv2.Sign(secretShares[n-1], []byte("invalid msg"))
		require.NoError(t, err)

		sigs = append(sigs, core.NewPartialSignature(tblsconv2.SigToCore(invalidSig), n))

		return sigs
	}

	corePubkey, err := core.PubKeyFromBytes(pubkey[:])
	require.NoError(t, err)

	// Aggregate and verify deposit data signatures
	msg := testutil.RandomDepositMsg(t)

	_, err = aggDepositData(
		map[core.PubKey][]core.ParSignedData{corePubkey: getSigs([]byte("any digest"))},
		[]share{shares},
		map[core.PubKey]eth2p0.DepositMessage{corePubkey: msg},
		eth2util.Goerli.Name,
	)
	require.EqualError(t, err, "invalid deposit data partial signature from peer")

	// Aggregate and verify cluster lock hash signatures
	lockMsg := []byte("cluster lock hash")

	_, _, err = aggLockHashSig(map[core.PubKey][]core.ParSignedData{corePubkey: getSigs(lockMsg)}, map[core.PubKey]share{corePubkey: shares}, lockMsg)
	require.EqualError(t, err, "invalid lock hash partial signature from peer: signature not verified")
}

func TestValidSignatures(t *testing.T) {
	const (
		n  = 4
		th = 3
	)

	secret, err := tblsv2.GenerateSecretKey()
	require.NoError(t, err)

	pubkey, err := tblsv2.SecretToPublicKey(secret)
	require.NoError(t, err)

	secretShares, err := tblsv2.ThresholdSplit(secret, n, th)
	require.NoError(t, err)

	pubshares := make(map[int]tblsv2.PublicKey)

	for idx, share := range secretShares {
		pubkey, err := tblsv2.SecretToPublicKey(share)
		require.NoError(t, err)

		pubshares[idx] = pubkey
	}

	shares := share{
		PubKey:       pubkey,
		SecretShare:  secret,
		PublicShares: pubshares,
	}

	getSigs := func(msg []byte) []core.ParSignedData {
		var sigs []core.ParSignedData
		for i := 0; i < n-1; i++ {
			pk := secretShares[i+1]
			sig, err := tblsv2.Sign(pk, msg)
			require.NoError(t, err)

			coreSig := tblsconv2.SigToCore(sig)
			sigs = append(sigs, core.NewPartialSignature(coreSig, i+1))
		}

		return sigs
	}

	corePubkey, err := core.PubKeyFromBytes(pubkey[:])
	require.NoError(t, err)
	eth2Pubkey, err := corePubkey.ToETH2()
	require.NoError(t, err)

	withdrawalAddr := testutil.RandomETHAddress()
	network := eth2util.Goerli.Name

	msg, err := deposit.NewMessage(eth2Pubkey, withdrawalAddr)
	require.NoError(t, err)
	sigRoot, err := deposit.GetMessageSigningRoot(msg, network)
	require.NoError(t, err)

	_, err = aggDepositData(
		map[core.PubKey][]core.ParSignedData{corePubkey: getSigs(sigRoot[:])},
		[]share{shares},
		map[core.PubKey]eth2p0.DepositMessage{corePubkey: msg},
		network,
	)
	require.NoError(t, err)

	// Aggregate and verify cluster lock hash signatures
	lockMsg := []byte("cluster lock hash")

	_, _, err = aggLockHashSig(map[core.PubKey][]core.ParSignedData{corePubkey: getSigs(lockMsg)}, map[core.PubKey]share{corePubkey: shares}, lockMsg)
	require.NoError(t, err)
}
