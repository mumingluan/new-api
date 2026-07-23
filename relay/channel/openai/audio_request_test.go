package openai

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertAudioRequestNormalizesMislabeledAMRForUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)

	amrData := append([]byte("#!AMR\n"), append([]byte{0x0c}, make([]byte, 13)...)...)
	var inboundBody bytes.Buffer
	inboundWriter := multipart.NewWriter(&inboundBody)
	require.NoError(t, inboundWriter.WriteField("model", "FunAudioLLM/SenseVoiceSmall"))
	require.NoError(t, inboundWriter.WriteField("response_format", "json"))
	filePart, err := inboundWriter.CreateFormFile("file", "audio.wav")
	require.NoError(t, err)
	_, err = filePart.Write(amrData)
	require.NoError(t, err)
	require.NoError(t, inboundWriter.Close())

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", &inboundBody)
	c.Request.Header.Set("Content-Type", inboundWriter.FormDataContentType())

	converted, err := (&Adaptor{}).ConvertAudioRequest(c, &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeAudioTranscription,
	}, dto.AudioRequest{Model: "FunAudioLLM/SenseVoiceSmall"})
	require.NoError(t, err)
	convertedBytes, err := io.ReadAll(converted)
	require.NoError(t, err)

	upstreamRequest := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", bytes.NewReader(convertedBytes))
	upstreamRequest.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	require.NoError(t, upstreamRequest.ParseMultipartForm(32<<20))
	require.Len(t, upstreamRequest.MultipartForm.File["file"], 1)

	upstreamFileHeader := upstreamRequest.MultipartForm.File["file"][0]
	assert.Equal(t, "audio.amr", upstreamFileHeader.Filename)
	assert.Equal(t, "audio/amr", upstreamFileHeader.Header.Get("Content-Type"))
	upstreamFile, err := upstreamFileHeader.Open()
	require.NoError(t, err)
	defer upstreamFile.Close()
	upstreamData, err := io.ReadAll(upstreamFile)
	require.NoError(t, err)
	assert.Equal(t, amrData, upstreamData)

	originalForm, err := common.ParseMultipartFormReusable(c)
	require.NoError(t, err)
	require.Len(t, originalForm.File["file"], 1)
	assert.Equal(t, "audio.wav", originalForm.File["file"][0].Filename)
}
