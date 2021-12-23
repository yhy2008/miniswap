// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package exchange

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"
	"testing"

	"miniswap/gobinding/token"
	"miniswap/gobinding/factory"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

func etherToWei(eth *big.Float) *big.Int {
	truncInt, _ := eth.Int(nil)
	truncInt = new(big.Int).Mul(truncInt, big.NewInt(params.Ether))
	fracStr := strings.Split(fmt.Sprintf("%.18f", eth), ".")[1]
	fracStr += strings.Repeat("0", 18-len(fracStr))
	fracInt, _ := new(big.Int).SetString(fracStr, 10)
	wei := new(big.Int).Add(truncInt, fracInt)
	return wei
}

func weiToEtherStr(amount *big.Int) string {
	compact_amount := big.NewInt(0)
	reminder := big.NewInt(0)
	divisor := big.NewInt(1e18)
	compact_amount.QuoRem(amount, divisor, reminder)
	return fmt.Sprintf("%v.%018s", compact_amount.String(), reminder.String())
}

func deploy() (*backends.SimulatedBackend, *bind.TransactOpts, *token.Token, *Exchange, common.Address) {
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: etherToWei(big.NewFloat(1000000))}}, math.MaxInt64)
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	auth.GasPrice = gasPrice

	tokenAddr, _, token, err := token.DeployToken(auth, sim, "Token", "TKN", etherToWei(big.NewFloat(1000000)))
	if err != nil {
		log.Fatalf("Failed to deploy new token contract: %v", err)
	}
	exchangeAddr, _, exchange, err := DeployExchange(auth, sim, tokenAddr)
	if err != nil {
		log.Fatalf("Failed to deploy new exchange contract: %v", err)
	}
	sim.Commit()
	return sim, auth, token, exchange, exchangeAddr
}

func TestExchangeTransactor_AddLiquidity(t *testing.T) {
	backend, auth, token, exchange, exchangeAddr := deploy()
	_, err := token.Approve(auth, exchangeAddr, etherToWei(big.NewFloat(2000)))
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	type args struct {
		token *big.Float
		ether *big.Float
	}
	type want struct {
		token   string
		ether   string
		lpToken string
		err     string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "first time add liquidity",
			args: args{
				token: big.NewFloat(0),
				ether: big.NewFloat(0),
			},
			want: want{
				token:   "0.000000000000000000",
				ether:   "0.000000000000000000",
				lpToken: "0.000000000000000000",
			},
		},
		{
			name: "second time add liquidity",
			args: args{
				token: big.NewFloat(200),
				ether: big.NewFloat(100),
			},
			want: want{
				token:   "200.000000000000000000",
				ether:   "100.000000000000000000",
				lpToken: "100.000000000000000000",
			},
		},
		{
			name: "insufficient token",
			args: args{
				token: big.NewFloat(100),
				ether: big.NewFloat(100),
			},
			want: want{
				token:   "200.000000000000000000",
				ether:   "100.000000000000000000",
				lpToken: "100.000000000000000000",
				err:     "execution reverted: insufficient token amount",
			},
		},
		{
			name: "insufficient ether",
			args: args{
				token: big.NewFloat(200),
				ether: big.NewFloat(50),
			},
			want: want{
				token:   "300.000000000000000000",
				ether:   "150.000000000000000000",
				lpToken: "150.000000000000000000",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth.Value = etherToWei(tt.args.ether)
			_, err = exchange.AddLiquidity(auth, etherToWei(tt.args.token))
			if err != nil && err.Error() != tt.want.err {
				t.Fatalf("AddLiquidity() error = %v, want %v", err, tt.want.err)
			}
			backend.Commit()

			got, err := backend.BalanceAt(context.Background(), exchangeAddr, nil)
			if err != nil {
				t.Fatalf("BalanceAt() error = %v", err)
				return
			}
			if weiToEtherStr(got) != tt.want.ether {
				t.Errorf("BalanceAt() = %v, want %v", weiToEtherStr(got), tt.want.ether)
			}
			got, err = exchange.GetReserve(nil)
			if err != nil {
				t.Fatalf("GetReserve() error = %v", err)
				return
			}
			if weiToEtherStr(got) != tt.want.token {
				t.Errorf("GetReserve() = %v, want %v", weiToEtherStr(got), tt.want.token)
			}
			got, err = exchange.BalanceOf(nil, auth.From)
			if err != nil {
				t.Fatalf("BalanceOf() error = %v", err)
				return
			}
			if weiToEtherStr(got) != tt.want.lpToken {
				t.Errorf("BalanceOf() = %v, want %v", weiToEtherStr(got), tt.want.lpToken)
			}
		})
	}
}

func TestExchangeTransactor_RemoveLiquidity(t *testing.T) {
	backend, auth, token, exchange, exchangeAddr := deploy()
	_, err := token.Approve(auth, exchangeAddr, etherToWei(big.NewFloat(2000)))
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	auth.Value = etherToWei(big.NewFloat(100))
	_, err = exchange.AddLiquidity(auth, etherToWei(big.NewFloat(200)))
	if err != nil {
		t.Fatalf("AddLiquidity() error = %v", err)
	}
	backend.Commit()
	auth.Value = big.NewInt(0)
	type args struct {
		token *big.Float
	}
	type want struct {
		token   string
		ether   string
		lpToken string
		err     string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "remove zero amount liquidity",
			args: args{
				token: big.NewFloat(0),
			},
			want: want{
				token:   "200.000000000000000000",
				ether:   "100.000000000000000000",
				lpToken: "100.000000000000000000",
				err:     "execution reverted: invalid amount",
			},
		},
		{
			name: "remove half amount liquidity",
			args: args{
				token: big.NewFloat(50),
			},
			want: want{
				token:   "100.000000000000000000",
				ether:   "50.000000000000000000",
				lpToken: "50.000000000000000000",
			},
		},
		{
			name: "remove all liquidity",
			args: args{
				token: big.NewFloat(50),
			},
			want: want{
				token:   "0.000000000000000000",
				ether:   "0.000000000000000000",
				lpToken: "0.000000000000000000",
			},
		},
		{
			name: "remove more liquidity",
			args: args{
				token: big.NewFloat(50),
			},
			want: want{
				token:   "0.000000000000000000",
				ether:   "0.000000000000000000",
				lpToken: "0.000000000000000000",
				err:     "execution reverted",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := exchange.RemoveLiquidity(auth, etherToWei(tt.args.token))
			if err != nil && err.Error() != tt.want.err {
				t.Fatalf("ExchangeTransactor.RemoveLiquidity() error = %v, want %v", err, tt.want.err)
			}
			backend.Commit()

			got, err := backend.BalanceAt(context.Background(), exchangeAddr, nil)
			if err != nil {
				t.Fatalf("BalanceAt() error = %v", err)
				return
			}
			if weiToEtherStr(got) != tt.want.ether {
				t.Errorf("BalanceAt() = %v, want %v", weiToEtherStr(got), tt.want.ether)
			}
			got, err = exchange.GetReserve(nil)
			if err != nil {
				t.Fatalf("GetReserve() error = %v", err)
				return
			}
			if weiToEtherStr(got) != tt.want.token {
				t.Errorf("GetReserve() = %v, want %v", weiToEtherStr(got), tt.want.token)
			}
			got, err = exchange.BalanceOf(nil, auth.From)
			if err != nil {
				t.Fatalf("BalanceOf() error = %v", err)
				return
			}
			if weiToEtherStr(got) != tt.want.lpToken {
				t.Errorf("BalanceOf() = %v, want %v", weiToEtherStr(got), tt.want.lpToken)
			}
		})
	}
}

func TestExchangeCallerSession_GetTokenAmount(t *testing.T) {
	backend, auth, token, exchange, exchangeAddr := deploy()
	_, err := token.Approve(auth, exchangeAddr, etherToWei(big.NewFloat(2000)))
	auth.Value = etherToWei(big.NewFloat(1000))
	_, err = exchange.AddLiquidity(auth, etherToWei(big.NewFloat(2000)))
	if err != nil {
		t.Fatalf("AddLiquidity() error = %v", err)
	}
	backend.Commit()

	type args struct {
		soldETH *big.Float
	}
	type want struct {
		getToken string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "first sold token",
			args: args{
				soldETH: big.NewFloat(1),
			},
			want: want{
				getToken: "1.978041738678708079",
			},
		},
		{
			name: "second sold token",
			args: args{
				soldETH: big.NewFloat(100),
			},
			want: want{
				getToken: "180.163785259326660600",
			},
		},
		{
			name: "third sold token",
			args: args{
				soldETH: big.NewFloat(1000),
			},
			want: want{
				getToken: "994.974874371859296482",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exchange.GetTokenAmount(nil, etherToWei(tt.args.soldETH))
			if err != nil {
				t.Fatalf("GetTokenAmount() error = %v", err)
			}
			if weiToEtherStr(got) != tt.want.getToken {
				t.Errorf("GetTokenAmount() = %v, want %v", weiToEtherStr(got), tt.want.getToken)
			}
		})
	}
}

func TestExchangeCallerSession_GetETHAmount(t *testing.T) {
	backend, auth, token, exchange, exchangeAddr := deploy()
	_, err := token.Approve(auth, exchangeAddr, etherToWei(big.NewFloat(2000)))
	auth.Value = etherToWei(big.NewFloat(1000))
	_, err = exchange.AddLiquidity(auth, etherToWei(big.NewFloat(2000)))
	if err != nil {
		t.Fatalf("AddLiquidity() error = %v", err)
		return
	}
	backend.Commit()

	type args struct {
		soldETH *big.Float
	}
	type want struct {
		getToken string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "first sold token",
			args: args{
				soldETH: big.NewFloat(2),
			},
			want: want{
				getToken: "0.989020869339354039",
			},
		},
		{
			name: "second sold token",
			args: args{
				soldETH: big.NewFloat(100),
			},
			want: want{
				getToken: "47.165316817532158170",
			},
		},
		{
			name: "third sold token",
			args: args{
				soldETH: big.NewFloat(2000),
			},
			want: want{
				getToken: "497.487437185929648241",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exchange.GetEthAmount(nil, etherToWei(tt.args.soldETH))
			if err != nil {
				t.Fatalf("GetEthAmount() error = %v", err)
			}
			if weiToEtherStr(got) != tt.want.getToken {
				t.Errorf("GetEthAmount() = %v, want %v", weiToEtherStr(got), tt.want.getToken)
			}
		})
	}
}

func TestExchangeTransactor_EthToTokenSwap(t *testing.T) {
	backend, auth, token, exchange, exchangeAddr := deploy()
	_, err := token.Approve(auth, exchangeAddr, etherToWei(big.NewFloat(2000)))
	auth.Value = etherToWei(big.NewFloat(1000))
	_, err = exchange.AddLiquidity(auth, etherToWei(big.NewFloat(2000)))
	if err != nil {
		t.Fatalf("AddLiquidity() error = %v", err)
		return
	}
	backend.Commit()

	type args struct {
		minToken *big.Float
		ether *big.Float
	}
	type want struct {
		token string
		ether string
		err string
	}
	tests := []struct {
		name    string
		args    args
		want    want
	}{
		{
			name: "sold zero ether",
			args: args{
				minToken: big.NewFloat(0),
				ether: big.NewFloat(0),
			},
			want: want{
				token: "998000.000000000000000000",
				ether: "998999.996413302375000000",
			},
		},
		{
			name: "sold 100 ether",
			args: args{
				minToken: big.NewFloat(100),
				ether: big.NewFloat(100),
			},
			want: want{
				token: "998180.163785259326660600",
				ether: "998899.996374539875000000",
			},
		},
		{
			name: "min token greater than actual token",
			args: args{
				minToken: big.NewFloat(200),
				ether: big.NewFloat(100),
			},
			want: want{
				token: "998180.163785259326660600",
				ether: "998899.996374539875000000",
				err: "execution reverted: insufficient output amount",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth.Value = etherToWei(tt.args.ether)
			_, err = exchange.EthToTokenSwap(auth, etherToWei(tt.args.minToken))
			if err != nil && err.Error() != tt.want.err {
				t.Errorf("ExchangeTransactor.EthToTokenSwap() error = %v, want %v", err, tt.want.err)
			}
			backend.Commit()

			auth.Value = big.NewInt(0)
			got, err := token.BalanceOf(nil, auth.From)
			if err != nil {
				t.Fatalf("BalanceOf() error = %v", err)
				return
			}
			if weiToEtherStr(got) != tt.want.token {
				t.Errorf("BalanceOf() = %v, want %v", weiToEtherStr(got), tt.want.token)
			}
			// got, err = backend.BalanceAt(context.Background(), auth.From, nil)
			// if err != nil {
			// 	t.Fatalf("BalanceAt() error = %v", err)
			// 	return
			// }
			// if weiToEtherStr(got) != tt.want.ether {
			// 	t.Errorf("BalanceAt() = %v, want %v", weiToEtherStr(got), tt.want.ether)
			// }
		})
	}
}

func TestExchangeTransactor_TokenToEthSwap(t *testing.T) {
	backend, auth, token, exchange, exchangeAddr := deploy()
	_, err := token.Approve(auth, exchangeAddr, etherToWei(big.NewFloat(3000)))
	auth.Value = etherToWei(big.NewFloat(1000))
	_, err = exchange.AddLiquidity(auth, etherToWei(big.NewFloat(2000)))
	if err != nil {
		t.Fatalf("AddLiquidity() error = %v", err)
		return
	}
	backend.Commit()

	auth.Value = big.NewInt(0)
	type args struct {
		token *big.Float
		minEther *big.Float
	}
	type want struct {
		token string
		ether string
		err string
	}
	tests := []struct {
		name    string
		args    args
		want    want
	}{
		{
			name: "sold zero token",
			args: args{
				token: big.NewFloat(0),
				minEther: big.NewFloat(0),
			},
			want: want{
				token: "998000.000000000000000000",
				ether: "998999.996404456125000000",
			},
		},
		{
			name: "sold 100 token",
			args: args{
				token: big.NewFloat(100),
				minEther: big.NewFloat(40),
			},
			want: want{
				token: "997900.000000000000000000",
				ether: "999047.161669478907158170",
			},
		},
		{
			name: "min ether greater than actual ether",
			args: args{
				token: big.NewFloat(200),
				minEther: big.NewFloat(100),
			},
			want: want{
				token: "997900.000000000000000000",
				ether: "999047.161669478907158170",
				err: "execution reverted: insufficient output amount",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = exchange.TokenToEthSwap(auth, etherToWei(tt.args.token), etherToWei(tt.args.minEther))
			if err != nil && err.Error() != tt.want.err {
				t.Errorf("ExchangeTransactor.TokenToEtherSwap() error = %v, want %v", err, tt.want.err)
			}
			backend.Commit()

			got, err := token.BalanceOf(nil, auth.From)
			if err != nil {
				t.Fatalf("BalanceOf() error = %v", err)
				return
			}
			if weiToEtherStr(got) != tt.want.token {
				t.Errorf("BalanceOf() = %v, want %v", weiToEtherStr(got), tt.want.token)
			}
			// got, err = backend.BalanceAt(context.Background(), auth.From, nil)
			// if err != nil {
			// 	t.Fatalf("BalanceAt() error = %v", err)
			// 	return
			// }
			// if weiToEtherStr(got) != tt.want.ether {
			// 	t.Errorf("BalanceAt() = %v, want %v", weiToEtherStr(got), tt.want.ether)
			// }
		})
	}
}

func TestExchangeTransactor_TokenToTokenSwap(t *testing.T) {
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: etherToWei(big.NewFloat(1000000))}}, math.MaxInt64)
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	auth.GasPrice = gasPrice

	tokenOneAddr, _, tokenOne, err := token.DeployToken(auth, sim, "TokenOne", "TKNO", etherToWei(big.NewFloat(1000000)))
	if err != nil {
		log.Fatalf("Failed to deploy new token contract: %v", err)
		return
	}
	tokenTwoAddr, _, tokenTwo, err := token.DeployToken(auth, sim, "TokenTwo", "TKNT", etherToWei(big.NewFloat(1000000)))
	if err != nil {
		log.Fatalf("Failed to deploy new token contract: %v", err)
		return
	}

	_, _, factory, err := factory.DeployFactory(auth, sim)
	if err != nil {
		log.Fatalf("Failed to deploy new factory contract: %v", err)
		return
	}

	_, err = factory.CreateExchange(auth, tokenOneAddr)
	if err != nil {
		log.Fatalf("Failed to create token one exchange: %v", err)
		return
	}

	_, err = factory.CreateExchange(auth, tokenTwoAddr)
	if err != nil {
		log.Fatalf("Failed to create token two exchange: %v", err)
		return
	}
	sim.Commit()

	tokenOneExchangeAddr, err := factory.GetExchange(nil, tokenOneAddr)
	if err != nil {
		log.Fatalf("Failed to get token one exchange address: %v", err)
		return
	}

	tokenTwoExchangeAddr, err := factory.GetExchange(nil, tokenTwoAddr)
	if err != nil {
		log.Fatalf("Failed to get token two exchange address: %v", err)
		return
	}

	tokenOneExchange, err := NewExchange(tokenOneExchangeAddr, sim)
	if err != nil {
		log.Fatalf("Failed to get token one exchange contract: %v", err)
		return
	}

	tokenTwoExchange, err := NewExchange(tokenTwoExchangeAddr, sim)
	if err != nil {
		log.Fatalf("Failed to get token two exchange contract: %v", err)
		return
	}

	_, err = tokenOne.Approve(auth, tokenOneExchangeAddr, etherToWei(big.NewFloat(3000)))
	if err != nil {
		log.Fatalf("Failed to approve token one: %v", err)
		return
	}

	_, err = tokenTwo.Approve(auth, tokenTwoExchangeAddr, etherToWei(big.NewFloat(3000)))
	if err != nil {
		log.Fatalf("Failed to approve token two: %v", err)
		return
	}

	auth.Value = etherToWei(big.NewFloat(1000))
	_, err = tokenOneExchange.AddLiquidity(auth, etherToWei(big.NewFloat(1000)))
	if err != nil {
		log.Fatalf("Failed to add liquidity to token one exchange: %v", err)
		return
	}
	_, err = tokenTwoExchange.AddLiquidity(auth, etherToWei(big.NewFloat(2000)))
	if err != nil {
		log.Fatalf("Failed to add liquidity to token two exchange: %v", err)
		return
	}

	sim.Commit()

	auth.Value = big.NewInt(0)
	type args struct {
		token *big.Float
		minToken *big.Float
	}
	type want struct {
		tokenOne string
		tokenTwo string
		err string
	}
	tests := []struct {
		name    string
		args    args
		want    want
	}{
		{
			name: "swap zero tokens",
			args: args{
				token: big.NewFloat(0),
				minToken: big.NewFloat(0),
			},
			want: want{
				tokenOne: "999000.000000000000000000",
				tokenTwo: "998000.000000000000000000",
			},
		},
		{
			name: "swap 100 tokens",
			args: args{
				token: big.NewFloat(100),
				minToken: big.NewFloat(100),
			},
			want: want{
				tokenOne: "998900.000000000000000000",
				tokenTwo: "998163.758030425810979022",
			},
		},
		{
			name: "min token greater than actual token",
			args: args{
				token: big.NewFloat(100),
				minToken: big.NewFloat(200),
			},
			want: want{
				tokenOne: "998900.000000000000000000",
				tokenTwo: "998163.758030425810979022",
				err: "execution reverted: insufficient output amount",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = tokenOneExchange.TokenToTokenSwap(auth, etherToWei(tt.args.token), etherToWei(tt.args.minToken), tokenTwoAddr)
			if err != nil && err.Error() != tt.want.err {
				t.Errorf("ExchangeTransactor.TokenToTokenSwap() error = %v, want %v", err, tt.want.err)
			}
			sim.Commit()

			got, err := tokenOne.BalanceOf(nil, auth.From)
			if err != nil {
				t.Fatalf("TokenOne BalanceOf() error = %v", err)
				return
			}
			if weiToEtherStr(got) != tt.want.tokenOne {
				t.Errorf("TokenOne BalanceOf() = %v, want %v", weiToEtherStr(got), tt.want.tokenOne)
			}
			got, err = tokenTwo.BalanceOf(nil, auth.From)
			if err != nil {
				t.Fatalf("TokenTwo BalanceAt() error = %v", err)
				return
			}
			if weiToEtherStr(got) != tt.want.tokenTwo {
				t.Errorf("TokenTwo BalanceAt() = %v, want %v", weiToEtherStr(got), tt.want.tokenTwo)
			}
		})
	}
}
