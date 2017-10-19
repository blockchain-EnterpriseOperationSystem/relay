/*

  Copyright 2017 Loopring Project Ltd (Loopring Foundation).

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/

package ipfs_test

import (
	"encoding/json"
	"github.com/Loopring/ringminer/test"
	"github.com/Loopring/ringminer/types"
	"github.com/ipfs/go-ipfs-api"
	"math/big"
	"math/rand"
	"strconv"
	"testing"
)

func TestA(t *testing.T) {

	sh := shell.NewLocalShell()

	suffix := "000000000000"
	//scheme 1:MarginSplitPercentage = 0
	for i := 1; i <= 1000; i++ {
		s1 := rand.Intn(100)
		b1 := rand.Intn(10)
		if b1 == 0 {
			b1 = b1 + 1
		}
		if s1 == 0 {
			s1 = s1 + 1
		}
		amountS1, _ := new(big.Int).SetString(strconv.Itoa(s1)+suffix, 0)
		amountB1, _ := new(big.Int).SetString(strconv.Itoa(b1)+suffix, 0)
		saddr := rand.Intn(10)
		baddr := rand.Intn(10)
		order1 := test.CreateOrder(
			types.HexToAddress("0x45fbfca061477ed69cf8d94673d04295bba15f2e"+strconv.Itoa(saddr)),
			types.HexToAddress("0xab4c5bda9cfa40e6877f29b0d729af66ad217958"+strconv.Itoa(baddr)),
			amountS1,
			amountB1,
			types.Hex2Bytes("11293da8fdfe3898eae7637e429e7e93d17d0d8293a4d1b58819ac0ca102b446"),
		)
		order1.Owner = types.HexToAddress("0xb5fab0b11776aad5ce60588c16bd59dcfd61a1c2")

		data1, _ := json.Marshal(order1)
		pubMessage(sh, string(data1))

		s2 := rand.Intn(100)
		b2 := rand.Intn(10)
		if b2 == 0 {
			b2 = b2 + 1
		}
		if s2 == 0 {
			s2 = s2 + 1
		}
		amountS2, _ := new(big.Int).SetString(strconv.Itoa(s2)+suffix, 0)
		amountB2, _ := new(big.Int).SetString(strconv.Itoa(b2)+suffix, 0)
		order2 := test.CreateOrder(
			types.HexToAddress("0xab4c5bda9cfa40e6877f29b0d729af66ad217958"+strconv.Itoa(saddr)),
			types.HexToAddress("0x45fbfca061477ed69cf8d94673d04295bba15f2e"+strconv.Itoa(baddr)),
			amountS2,
			amountB2,
			types.Hex2Bytes("07ae9ee56203d29171ce3de536d7742e0af4df5b7f62d298a0445d11e466bf9e"),
		)
		order2.Owner = types.HexToAddress("0x48ff2269e58a373120FFdBBdEE3FBceA854AC30A")
		data2, _ := json.Marshal(order2)
		pubMessage(sh, string(data2))
	}

}

func TestB(t *testing.T) {
	sh := shell.NewLocalShell()

	suffix := "0"

	//scheme 1:MarginSplitPercentage = 0
	amountS1, _ := new(big.Int).SetString("1"+suffix, 0)
	amountB1, _ := new(big.Int).SetString("10"+suffix, 0)
	order1 := test.CreateOrder(
		types.HexToAddress("0x937ff659c8a9d85aac39dfa84c4b49bb7c9b226e"),
		types.HexToAddress("0x8711ac984e6ce2169a2a6bd83ec15332c366ee4f"),
		amountS1,
		amountB1,
		types.Hex2Bytes("11293da8fdfe3898eae7637e429e7e93d17d0d8293a4d1b58819ac0ca102b446"),
	)
	order1.Owner = types.HexToAddress("0xb5fab0b11776aad5ce60588c16bd59dcfd61a1c2")
	data1, _ := json.Marshal(order1)
	pubMessage(sh, string(data1))

	amountS2, _ := new(big.Int).SetString("20"+suffix, 0)
	amountB2, _ := new(big.Int).SetString("1"+suffix, 0)
	order2 := test.CreateOrder(
		types.HexToAddress("0x8711ac984e6ce2169a2a6bd83ec15332c366ee4f"),
		types.HexToAddress("0x937ff659c8a9d85aac39dfa84c4b49bb7c9b226e"),
		amountS2,
		amountB2,
		types.Hex2Bytes("07ae9ee56203d29171ce3de536d7742e0af4df5b7f62d298a0445d11e466bf9e"),
	)
	order2.Owner = types.HexToAddress("0x48ff2269e58a373120FFdBBdEE3FBceA854AC30A")
	data2, _ := json.Marshal(order2)
	pubMessage(sh, string(data2))
}

func pubMessage(sh *shell.Shell, data string) {
	err := sh.PubSubPublish("test_topic", data)
	if err != nil {
		panic(err.Error())
	}
}
