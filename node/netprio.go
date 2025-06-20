// Copyright (C) 2019-2025 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

package node

import (
	"encoding/base64"
	"fmt"

	"github.com/algorand/go-algorand/crypto"
	"github.com/algorand/go-algorand/data/account"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/protocol"
)

const netPrioChallengeSize = 32

const netPrioChallengeSizeBase64Encoded = 44 // 32 * (4/3) rounded up to nearest multiple of 4 -> 44

type netPrioResponse struct {
	_struct struct{} `codec:",omitempty,omitemptyarray"`

	Nonce string `codec:"Nonce,allocbound=netPrioChallengeSizeBase64Encoded"`
}

type netPrioResponseSigned struct {
	_struct struct{} `codec:",omitempty,omitemptyarray"`

	Response netPrioResponse
	Round    basics.Round
	Sender   basics.Address
	Sig      crypto.OneTimeSignature
}

func (npr netPrioResponse) ToBeHashed() (protocol.HashID, []byte) {
	return protocol.NetPrioResponse, protocol.Encode(&npr)
}

// NewPrioChallenge implements the network.NetPrioScheme interface
func (node *AlgorandFullNode) NewPrioChallenge() string {
	var rand [netPrioChallengeSize]byte
	crypto.RandBytes(rand[:])
	return base64.StdEncoding.EncodeToString(rand[:])
}

// MakePrioResponse implements the network.NetPrioScheme interface
func (node *AlgorandFullNode) MakePrioResponse(challenge string) []byte {
	if !node.config.AnnounceParticipationKey {
		return nil
	}

	rs := netPrioResponseSigned{
		Response: netPrioResponse{
			Nonce: challenge,
		},
	}

	// Find the participation key that has the highest weight in the
	// latest round.
	var maxWeight uint64
	var maxPart account.ParticipationRecordForRound

	latest := node.ledger.LastRound()
	proto, err := node.ledger.ConsensusParams(latest)
	if err != nil {
		return nil
	}

	// Use the participation key for 2 rounds in the future, so that
	// it's unlikely to be deleted from underneath of us.
	voteRound := latest + 2
	for _, part := range node.accountManager.Keys(voteRound) {
		parent := part.Account
		data, err := node.ledger.LookupAgreement(latest, parent)
		if err != nil {
			continue
		}

		weight := data.MicroAlgosWithRewards.ToUint64()
		if weight > maxWeight {
			maxPart = part
			maxWeight = weight
		}
	}

	if maxWeight == 0 {
		return nil
	}

	signer := maxPart.VotingSigner()
	ephID := basics.OneTimeIDForRound(voteRound, signer.KeyDilution(proto.DefaultKeyDilution))

	rs.Round = voteRound
	rs.Sender = maxPart.Account
	rs.Sig = signer.Sign(ephID, rs.Response)

	return protocol.Encode(&rs)
}

// VerifyPrioResponse implements the network.NetPrioScheme interface
func (node *AlgorandFullNode) VerifyPrioResponse(challenge string, response []byte) (addr basics.Address, err error) {
	var rs netPrioResponseSigned
	err = protocol.Decode(response, &rs)
	if err != nil {
		return
	}

	if rs.Response.Nonce != challenge {
		err = fmt.Errorf("challenge/response mismatch")
		return
	}

	balanceRound := rs.Round.SubSaturate(2)
	proto, err := node.ledger.ConsensusParams(balanceRound)
	if err != nil {
		return
	}

	data, err := node.ledger.LookupAgreement(balanceRound, rs.Sender)
	if err != nil {
		return
	}

	ephID := basics.OneTimeIDForRound(rs.Round, proto.EffectiveKeyDilution(data.VoteKeyDilution))
	if !data.VoteID.Verify(ephID, rs.Response, rs.Sig) {
		err = fmt.Errorf("signature verification failure")
		return
	}

	addr = rs.Sender
	return
}

// GetPrioWeight implements the network.NetPrioScheme interface
func (node *AlgorandFullNode) GetPrioWeight(addr basics.Address) uint64 {
	latest := node.ledger.LastRound()
	data, err := node.ledger.LookupAgreement(latest, addr)
	if err != nil {
		return 0
	}

	return data.MicroAlgosWithRewards.ToUint64()
}
