package service

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type web3Service struct {
}

type Web3Service interface {
	VerifySignature(address, data, signatureData string) (bool, error)
}

func NewWeb3Service() Web3Service {
	return web3Service{}
}

// https://gist.github.com/dcb9/385631846097e1f59e3cba3b1d42f3ed#file-eth_sign_verify-go
func (w web3Service) _VerifySignatureOriginal(address, data, signatureData string) (bool, error) {
	sig := hexutil.MustDecode(signatureData)
	msg := accounts.TextHash([]byte(data))
	sig[crypto.RecoveryIDOffset] -= 27
	recovered, err := crypto.SigToPub(msg, sig)
	if err != nil {
		return false, nil
	}
	recoveredAddr := crypto.PubkeyToAddress(*recovered)
	return address == recoveredAddr.Hex(), nil
}

func (w web3Service) VerifySignature(address, data, signatureData string) (bool, error) {
	sig, err := hexutil.Decode(signatureData)
	if err != nil {
		return false, err
	}
	msg := accounts.TextHash([]byte(data))

	sig[crypto.RecoveryIDOffset] -= 27

	publicKey, err := crypto.SigToPub(msg, sig)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %v", err)
	}
	recoveredAddress := crypto.PubkeyToAddress(*publicKey)
	givenAddress := strings.ToLower(address)
	isVerified := strings.ToLower(recoveredAddress.Hex()) == givenAddress

	return isVerified, nil
}
