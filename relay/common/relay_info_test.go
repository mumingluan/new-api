package common

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/types"
	"github.com/stretchr/testify/require"
)

func TestRelayInfoResetResponseStateForRetry(t *testing.T) {
	t.Parallel()

	info := &RelayInfo{
		SendResponseCount:     3,
		ReceivedResponseCount: 4,
		FirstResponseTime:     time.Now(),
		isFirstResponse:       false,
		StreamStatus:          &StreamStatus{},
		ThinkingContentInfo: ThinkingContentInfo{
			HasSentThinkingContent: true,
		},
		ClaudeConvertInfo: &ClaudeConvertInfo{
			LastMessagesType: LastMessageTypeTools,
			Done:             true,
		},
		ResponsesUsageInfo: &ResponsesUsageInfo{
			BuiltInTools: map[string]*BuildInToolInfo{
				"web_search": {CallCount: 2},
			},
		},
	}

	info.ResetResponseStateForRetry()

	require.Zero(t, info.SendResponseCount)
	require.Zero(t, info.ReceivedResponseCount)
	require.True(t, info.FirstResponseTime.IsZero())
	require.True(t, info.isFirstResponse)
	require.Nil(t, info.StreamStatus)
	require.False(t, info.ThinkingContentInfo.HasSentThinkingContent)
	require.Equal(t, LastMessageTypeNone, info.ClaudeConvertInfo.LastMessagesType)
	require.False(t, info.ClaudeConvertInfo.Done)
	require.Zero(t, info.ResponsesUsageInfo.BuiltInTools["web_search"].CallCount)
}

func TestRelayInfoGetFinalRequestRelayFormatPrefersExplicitFinal(t *testing.T) {
	info := &RelayInfo{
		RelayFormat:             types.RelayFormatOpenAI,
		RequestConversionChain:  []types.RelayFormat{types.RelayFormatOpenAI, types.RelayFormatClaude},
		FinalRequestRelayFormat: types.RelayFormatOpenAIResponses,
	}

	require.Equal(t, types.RelayFormat(types.RelayFormatOpenAIResponses), info.GetFinalRequestRelayFormat())
}

func TestRelayInfoGetFinalRequestRelayFormatFallsBackToConversionChain(t *testing.T) {
	info := &RelayInfo{
		RelayFormat:            types.RelayFormatOpenAI,
		RequestConversionChain: []types.RelayFormat{types.RelayFormatOpenAI, types.RelayFormatClaude},
	}

	require.Equal(t, types.RelayFormat(types.RelayFormatClaude), info.GetFinalRequestRelayFormat())
}

func TestRelayInfoGetFinalRequestRelayFormatFallsBackToRelayFormat(t *testing.T) {
	info := &RelayInfo{
		RelayFormat: types.RelayFormatGemini,
	}

	require.Equal(t, types.RelayFormat(types.RelayFormatGemini), info.GetFinalRequestRelayFormat())
}

func TestRelayInfoGetFinalRequestRelayFormatNilReceiver(t *testing.T) {
	var info *RelayInfo
	require.Equal(t, types.RelayFormat(""), info.GetFinalRequestRelayFormat())
}
