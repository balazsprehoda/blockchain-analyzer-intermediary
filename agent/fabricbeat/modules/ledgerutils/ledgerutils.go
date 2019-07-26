package ledgerutils

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	protoCommon "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"

	"github.com/hyperledger-elastic/agent/fabricbeat/modules/fabricutils"

	"github.com/elastic/beats/libbeat/logp"
)

func GetBlockHash(blockNumber uint64, ledgerClient *ledger.Client) (string, error) {
	blockResponse, blockError := ledgerClient.QueryBlock(blockNumber)
	if blockError != nil {
		return "", blockError
	}
	logp.Info("Querying last known block from ledger successful")

	blockHash := fabricutils.GenerateBlockHash(blockResponse.Header.PreviousHash, blockResponse.Header.DataHash, blockResponse.Header.Number)
	return blockHash, nil
}

func GetBlockHeight(ledgerClient *ledger.Client) (uint64, error) {
	// Get the block height of this channel
	infoResponse, err := ledgerClient.QueryInfo()
	if err != nil {
		return 0, err
	}
	blockHeight := infoResponse.BCI.Height
	return blockHeight, nil
}

func ProcessBlock(blockNumber uint64, ledgerClient *ledger.Client) (blockResponse *protoCommon.Block, typeInfo string, createdAt time.Time, txsFltr util.TxValidationFlags, err error) {
	blockResponse, blockError := ledgerClient.QueryBlock(blockNumber)
	if blockError != nil {
		return nil, "", time.Now(), nil, blockError
	}

	// Getting block creation timestamp
	env, err := utils.GetEnvelopeFromBlock(blockResponse.Data.Data[0])
	if err != nil {
		return nil, "", time.Now(), nil, err
	}

	channelHeader, err := utils.ChannelHeader(env)
	if err != nil {
		return nil, "", time.Now(), nil, err
	}
	typeCode := channelHeader.GetType()
	typeInfo = fabricutils.TypeCodeToInfo(typeCode)
	createdAt = time.Unix(channelHeader.GetTimestamp().Seconds, int64(channelHeader.GetTimestamp().Nanos))

	// Checking validity
	txsFltr = util.TxValidationFlags(blockResponse.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	return
}

func ProcessTx(txData []byte) (txId, channelId, creator, creatorOrg string, tx *peer.Transaction, err error) {
	env, err := utils.GetEnvelopeFromBlock(txData)
	if err != nil {
		return "", "", "", "", nil, err
	}

	payload, err := utils.GetPayload(env)
	if err != nil {
		return "", "", "", "", nil, err
	}
	chdr, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return "", "", "", "", nil, err
	}

	shdr, err := utils.GetSignatureHeader(payload.Header.SignatureHeader)
	if err != nil {
		return "", "", "", "", nil, err
	}

	tx, err = utils.GetTransaction(payload.Data)
	if err != nil {
		return "", "", "", "", nil, err
	}

	return chdr.TxId, chdr.ChannelId, fabricutils.ReturnCreatorString(shdr.Creator), fabricutils.ReturnCreatorOrgString(shdr.Creator), tx, nil
}

func ProcessEndorserTx(txData []byte) (txId, channelId, creator, creatorOrg string, txRWSet *rwsetutil.TxRwSet, chaincodeName, chaincodeVersion string, err error) {

	txId, channelId, creator, creatorOrg, tx, err := ProcessTx(txData)
	if err != nil {
		return "", "", "", "", nil, "", "", err
	}

	_, respPayload, payloadErr := utils.GetPayloads(tx.Actions[0])
	if payloadErr != nil {
		return "", "", "", "", nil, "", "", err
	}

	txRWSet = &rwsetutil.TxRwSet{}
	err = txRWSet.FromProtoBytes(respPayload.Results)
	if err != nil {
		return "", "", "", "", nil, "", "", err
	}

	return txId, channelId, creator, creatorOrg, txRWSet, respPayload.ChaincodeId.Name, respPayload.ChaincodeId.Version, nil
}
