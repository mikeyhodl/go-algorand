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

package upgrades

import (
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/protocol"
	"github.com/algorand/go-algorand/test/framework/fixtures"
	"github.com/algorand/go-algorand/test/partitiontest"
)

func GenerateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil
	}

	return b
}

// this test checks that two accounts can send money to one another
// across a protocol upgrade.
func TestAccountsCanSendMoneyAcrossUpgradeV15toV16(t *testing.T) {
	partitiontest.PartitionTest(t)
	defer fixtures.ShutdownSynchronizedTest(t)

	testAccountsCanSendMoneyAcrossUpgrade(t, filepath.Join("nettemplates", "TwoNodes50EachV15Upgrade.json"), "")
}

func TestAccountsCanSendMoneyAcrossUpgradeV21toV22(t *testing.T) {
	partitiontest.PartitionTest(t)
	defer fixtures.ShutdownSynchronizedTest(t)

	testAccountsCanSendMoneyAcrossUpgrade(t, filepath.Join("nettemplates", "TwoNodes50EachV21Upgrade.json"), "")
}

func TestAccountsCanSendMoneyAcrossUpgradeV22toV23(t *testing.T) {
	partitiontest.PartitionTest(t)
	defer fixtures.ShutdownSynchronizedTest(t)

	testAccountsCanSendMoneyAcrossUpgrade(t, filepath.Join("nettemplates", "TwoNodes50EachV22Upgrade.json"), "")
}

func TestAccountsCanSendMoneyAcrossUpgradeV23toV24(t *testing.T) {
	partitiontest.PartitionTest(t)
	defer fixtures.ShutdownSynchronizedTest(t)

	testAccountsCanSendMoneyAcrossUpgrade(t, filepath.Join("nettemplates", "TwoNodes50EachV23Upgrade.json"), "")
}

func TestAccountsCanSendMoneyAcrossUpgradeV24toV25(t *testing.T) {
	partitiontest.PartitionTest(t)
	defer fixtures.ShutdownSynchronizedTest(t)

	testAccountsCanSendMoneyAcrossUpgrade(t, filepath.Join("nettemplates", "TwoNodes50EachV24Upgrade.json"), "")
}

func TestAccountsCanSendMoneyAcrossUpgradeV32toV35(t *testing.T) {
	partitiontest.PartitionTest(t)
	defer fixtures.ShutdownSynchronizedTest(t)

	targetVersion := consensusTestFastUpgrade(protocol.ConsensusV35)
	testAccountsCanSendMoneyAcrossUpgrade(t, filepath.Join("nettemplates", "TwoNodes50EachV32Upgrade.json"), targetVersion)
}

// ConsensusTestFastUpgrade is meant for testing of protocol upgrades:
// during testing, it is equivalent to another protocol with the exception
// of the upgrade parameters, which allow for upgrades to take place after
// only a few rounds.
func consensusTestFastUpgrade(proto protocol.ConsensusVersion) protocol.ConsensusVersion {
	return "test-fast-upgrade-" + proto
}

func generateFastUpgradeConsensus() (fastUpgradeProtocols config.ConsensusProtocols) {
	fastUpgradeProtocols = make(config.ConsensusProtocols)

	for proto, params := range config.Consensus {
		fastParams := params
		fastParams.UpgradeVoteRounds = 5
		fastParams.UpgradeThreshold = 3
		fastParams.DefaultUpgradeWaitRounds = 5
		fastParams.MinUpgradeWaitRounds = 0
		fastParams.MaxUpgradeWaitRounds = 0
		fastParams.MaxVersionStringLen += len(consensusTestFastUpgrade(""))
		fastParams.ApprovedUpgrades = make(map[protocol.ConsensusVersion]uint64)
		// set the small lambda to 500 for the duration of dependent tests.
		fastParams.AgreementFilterTimeout = 500 * time.Millisecond
		fastParams.AgreementFilterTimeoutPeriod0 = 500 * time.Millisecond

		for ver := range params.ApprovedUpgrades {
			fastParams.ApprovedUpgrades[consensusTestFastUpgrade(ver)] = 0
		}

		fastUpgradeProtocols[consensusTestFastUpgrade(proto)] = fastParams

	}
	return
}

func testAccountsCanSendMoneyAcrossUpgrade(t *testing.T, templatePath string, targetVersion protocol.ConsensusVersion) {
	//t.Parallel()
	a := require.New(fixtures.SynchronizedTest(t))

	consensus := generateFastUpgradeConsensus()

	var fixture fixtures.RestClientFixture
	fixture.SetConsensus(consensus)
	fixture.Setup(t, templatePath)
	defer fixture.Shutdown()

	verifyAccountsCanSendMoneyAcrossUpgrade(a, &fixture, targetVersion)
}

func verifyAccountsCanSendMoneyAcrossUpgrade(a *require.Assertions, fixture *fixtures.RestClientFixture, targetVersion protocol.ConsensusVersion) {
	pingBalance, pongBalance, expectedPingBalance, expectedPongBalance := runUntilProtocolUpgrades(a, fixture, targetVersion)

	a.True(expectedPingBalance <= pingBalance, "ping balance is different than expected")
	a.True(expectedPongBalance <= pongBalance, "pong balance is different than expected")
}

func runUntilProtocolUpgrades(a *require.Assertions, fixture *fixtures.RestClientFixture, targetVersion protocol.ConsensusVersion) (uint64, uint64, uint64, uint64) {
	c := fixture.LibGoalClient
	initialStatus, err := c.Status()
	a.NoError(err, "getting status")

	pingClient := fixture.LibGoalClient
	pingAccountList, err := fixture.GetWalletsSortedByBalance()
	a.NoError(err, "fixture should be able to get wallets sorted by balance")
	a.NotEmpty(pingAccountList)
	pingAccount := pingAccountList[0].Address

	pongClient := fixture.GetLibGoalClientForNamedNode("Node")
	wh, err := pongClient.GetUnencryptedWalletHandle()
	a.NoError(err)
	pongAccountList, err := pongClient.ListAddresses(wh)
	a.NoError(err)
	pongAccount := pongAccountList[0]

	pingBalance, err := c.GetBalance(pingAccount)
	a.NoError(err)
	pongBalance, err := c.GetBalance(pongAccount)
	a.NoError(err)

	a.Equal(pingBalance, pongBalance, "both accounts should start with same balance")
	a.NotEqual(pingAccount, pongAccount, "accounts under study should be different")

	expectedPingBalance := pingBalance
	expectedPongBalance := pongBalance

	const transactionFee = uint64(9000)
	const amountPongSendsPing = uint64(10000)
	const amountPingSendsPong = uint64(11000)

	curStatus, err := c.Status()
	a.NoError(err, "getting status")
	var pingTxids []string
	var pongTxids []string

	pongWalletHandle, err := pongClient.GetUnencryptedWalletHandle()
	a.NoError(err)
	pingWalletHandle, err := pingClient.GetUnencryptedWalletHandle()
	a.NoError(err)
	startTime := time.Now()
	var lastTxnSendRound basics.Round
	for curStatus.LastVersion == initialStatus.LastVersion {
		iterationStartTime := time.Now()
		if lastTxnSendRound != curStatus.LastRound {
			pongTx, err := pongClient.SendPaymentFromWallet(pongWalletHandle, nil, pongAccount, pingAccount, transactionFee, amountPongSendsPing, GenerateRandomBytes(8), "", 0, 0)
			a.NoError(err, "fixture should be able to send money (pong -> ping)")
			pongTxids = append(pongTxids, pongTx.ID().String())

			pingTx, err := pingClient.SendPaymentFromWallet(pingWalletHandle, nil, pingAccount, pongAccount, transactionFee, amountPingSendsPong, GenerateRandomBytes(8), "", 0, 0)
			a.NoError(err, "fixture should be able to send money (ping -> pong)")
			pingTxids = append(pingTxids, pingTx.ID().String())

			expectedPingBalance = expectedPingBalance - transactionFee - amountPingSendsPong + amountPongSendsPing
			expectedPongBalance = expectedPongBalance - transactionFee - amountPongSendsPing + amountPingSendsPong

			lastTxnSendRound = curStatus.LastRound
		}

		curStatus, err = pongClient.Status()
		a.NoError(err)

		pongWalletHandle, err = pongClient.GetUnencryptedWalletHandle()
		a.NoError(err)
		pingWalletHandle, err = pingClient.GetUnencryptedWalletHandle()
		a.NoError(err)

		iterationDuration := time.Since(iterationStartTime)
		if iterationDuration < 500*time.Millisecond {
			time.Sleep(500*time.Millisecond - iterationDuration)
		}

		if time.Now().After(startTime.Add(5 * time.Minute)) {
			a.Fail("upgrade taking too long")
		}
	}

	// optionally wait until the target version if set
	startTime = time.Now()
	if targetVersion != protocol.ConsensusVersion("") {
		for curStatus.LastVersion != string(targetVersion) {
			time.Sleep(500 * time.Millisecond)

			if time.Now().After(startTime.Add(5 * time.Minute)) {
				a.Fail("upgrade to target taking too long")
			}
			curStatus, err = pongClient.Status()
			a.NoError(err)
		}
	}

	initialStatus, err = c.Status()
	a.NoError(err, "getting status")

	// submit a few more transactions to make sure payments work in new protocol
	// perform this for two rounds.
	for {
		curStatus, err = pongClient.Status()
		a.NoError(err)
		if curStatus.LastRound > initialStatus.LastRound+2 {
			break
		}

		iterationStartTime := time.Now()
		if lastTxnSendRound != curStatus.LastRound {
			pongTx, err := pongClient.SendPaymentFromWallet(pongWalletHandle, nil, pongAccount, pingAccount, transactionFee, amountPongSendsPing, GenerateRandomBytes(8), "", 0, 0)
			a.NoError(err, "fixture should be able to send money (pong -> ping)")
			pongTxids = append(pongTxids, pongTx.ID().String())

			pingTx, err := pingClient.SendPaymentFromWallet(pingWalletHandle, nil, pingAccount, pongAccount, transactionFee, amountPingSendsPong, GenerateRandomBytes(8), "", 0, 0)
			a.NoError(err, "fixture should be able to send money (ping -> pong)")
			pingTxids = append(pingTxids, pingTx.ID().String())

			expectedPingBalance = expectedPingBalance - transactionFee - amountPingSendsPong + amountPongSendsPing
			expectedPongBalance = expectedPongBalance - transactionFee - amountPongSendsPing + amountPingSendsPong

			lastTxnSendRound = curStatus.LastRound
		}

		pongWalletHandle, err = pongClient.GetUnencryptedWalletHandle()
		a.NoError(err)
		pingWalletHandle, err = pingClient.GetUnencryptedWalletHandle()
		a.NoError(err)

		iterationDuration := time.Since(iterationStartTime)
		if iterationDuration < 500*time.Millisecond {
			time.Sleep(500*time.Millisecond - iterationDuration)
		}
	}

	curStatus, err = pongClient.Status()
	a.NoError(err)

	// wait for all transactions to confirm
	for _, txid := range pingTxids {
		_, err = fixture.WaitForConfirmedTxn(curStatus.LastRound+5, txid)
		a.NoError(err, "waiting for txn")
	}

	for _, txid := range pongTxids {
		_, err = fixture.WaitForConfirmedTxn(curStatus.LastRound+5, txid)
		a.NoError(err, "waiting for txn")
	}

	// check balances
	pingBalance, err = c.GetBalance(pingAccount)
	a.NoError(err)
	pongBalance, err = c.GetBalance(pongAccount)
	a.NoError(err)
	return pingBalance, pongBalance, expectedPingBalance, expectedPongBalance
}
