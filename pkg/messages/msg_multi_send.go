package messages

import (
	configTypes "main/pkg/config/types"
	"main/pkg/data_fetcher"
	"main/pkg/types"
	"main/pkg/types/event"
	"main/pkg/utils"

	cosmosTypes "github.com/cosmos/cosmos-sdk/types"
	cosmosBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/proto"
)

type MultiSendEntry struct {
	Address configTypes.Link
	Amount  types.Amounts
}

type MsgMultiSend struct {
	Inputs  []MultiSendEntry
	Outputs []MultiSendEntry
}

func ParseMsgMultiSend(data []byte, chain *configTypes.Chain, height int64) (types.Message, error) {
	var parsedMessage cosmosBankTypes.MsgMultiSend
	if err := proto.Unmarshal(data, &parsedMessage); err != nil {
		return nil, err
	}

	return &MsgMultiSend{
		Inputs: utils.Map(parsedMessage.Inputs, func(input cosmosBankTypes.Input) MultiSendEntry {
			return MultiSendEntry{
				Address: chain.GetWalletLink(input.Address),
				Amount:  utils.Map(input.Coins, types.AmountFrom),
			}
		}),
		Outputs: utils.Map(parsedMessage.Outputs, func(output cosmosBankTypes.Output) MultiSendEntry {
			return MultiSendEntry{
				Address: chain.GetWalletLink(output.Address),
				Amount:  utils.Map(output.Coins, types.AmountFrom),
			}
		}),
	}, nil
}

func (m MsgMultiSend) Type() string {
	return "/cosmos.bank.v1beta1.MsgMultiSend"
}

func (m *MsgMultiSend) GetAdditionalData(fetcher data_fetcher.DataFetcher) {
	price, found := fetcher.GetPrice()
	if !found {
		return
	}

	for _, input := range m.Inputs {
		if alias := fetcher.AliasManager.Get(fetcher.Chain.Name, input.Address.Value); alias != "" {
			input.Address.Title = alias
		}

		for _, amount := range input.Amount {
			if amount.Denom != fetcher.Chain.BaseDenom {
				continue
			}

			amount.AddUSDPrice(fetcher.Chain.DisplayDenom, fetcher.Chain.DenomCoefficient, price)
		}
	}

	for _, output := range m.Outputs {
		if alias := fetcher.AliasManager.Get(fetcher.Chain.Name, output.Address.Value); alias != "" {
			output.Address.Title = alias
		}

		for _, amount := range output.Amount {
			if amount.Denom != fetcher.Chain.BaseDenom {
				continue
			}

			amount.AddUSDPrice(fetcher.Chain.DisplayDenom, fetcher.Chain.DenomCoefficient, price)
		}
	}
}

func (m *MsgMultiSend) GetValues() event.EventValues {
	values := []event.EventValue{
		event.From(cosmosTypes.EventTypeMessage, cosmosTypes.AttributeKeyAction, m.Type()),
	}

	for _, input := range m.Inputs {
		values = append(values, []event.EventValue{
			event.From(cosmosBankTypes.EventTypeTransfer, cosmosBankTypes.AttributeKeySpender, input.Address.Value),
			event.From(cosmosBankTypes.EventTypeCoinSpent, cosmosBankTypes.AttributeKeySpender, input.Address.Value),
			event.From(cosmosTypes.EventTypeMessage, cosmosTypes.AttributeKeySender, input.Address.Value),
			event.From(cosmosBankTypes.EventTypeTransfer, cosmosTypes.AttributeKeyAmount, input.Amount.String()),
		}...)
	}

	for _, output := range m.Outputs {
		values = append(values, []event.EventValue{
			event.From(cosmosBankTypes.EventTypeTransfer, cosmosBankTypes.AttributeKeyRecipient, output.Address.Value),
			event.From(cosmosBankTypes.EventTypeCoinReceived, cosmosBankTypes.AttributeKeyReceiver, output.Address.Value),
			event.From(cosmosBankTypes.EventTypeTransfer, cosmosTypes.AttributeKeyAmount, output.Amount.String()),
		}...)
	}

	return values
}
