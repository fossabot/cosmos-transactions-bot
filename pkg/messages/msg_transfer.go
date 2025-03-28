package messages

import (
	configTypes "main/pkg/config/types"
	dataFetcher "main/pkg/data_fetcher"
	"main/pkg/types"
	"main/pkg/types/event"

	cosmosTypes "github.com/cosmos/cosmos-sdk/types"
	ibcTypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/gogo/protobuf/proto"
)

type MsgTransfer struct {
	Token    *types.Amount
	Sender   configTypes.Link
	Receiver configTypes.Link
}

func ParseMsgTransfer(data []byte, chain *configTypes.Chain, height int64) (types.Message, error) {
	var parsedMessage ibcTypes.MsgTransfer
	if err := proto.Unmarshal(data, &parsedMessage); err != nil {
		return nil, err
	}

	return &MsgTransfer{
		Token:    types.AmountFrom(parsedMessage.Token),
		Sender:   chain.GetWalletLink(parsedMessage.Sender),
		Receiver: configTypes.Link{Value: parsedMessage.Receiver},
	}, nil
}

func (m MsgTransfer) Type() string {
	return "/ibc.applications.transfer.v1.MsgTransfer"
}

func (m *MsgTransfer) GetAdditionalData(fetcher dataFetcher.DataFetcher) {
	price, found := fetcher.GetPrice()
	if found && m.Token.Denom == fetcher.Chain.BaseDenom {
		m.Token.AddUSDPrice(fetcher.Chain.DisplayDenom, fetcher.Chain.DenomCoefficient, price)
	}

	if alias := fetcher.AliasManager.Get(fetcher.Chain.Name, m.Sender.Value); alias != "" {
		m.Sender.Title = alias
	}
}

func (m *MsgTransfer) GetValues() event.EventValues {
	return []event.EventValue{
		event.From(cosmosTypes.EventTypeMessage, cosmosTypes.AttributeKeyAction, m.Type()),
		event.From(ibcTypes.EventTypeTransfer, ibcTypes.AttributeKeyReceiver, m.Receiver.Value),
		event.From(ibcTypes.EventTypeTransfer, cosmosTypes.AttributeKeySender, m.Sender.Value),
		event.From(ibcTypes.EventTypeTransfer, cosmosTypes.AttributeKeyAmount, m.Token.String()),
	}
}
