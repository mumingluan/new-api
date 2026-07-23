package common

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectAudioFormatUsesMagicBytesBeforeExtension(t *testing.T) {
	audio := bytes.NewReader(append([]byte("#!AMR\n"), append([]byte{0x0c}, make([]byte, 13)...)...))
	_, err := audio.Seek(3, io.SeekStart)
	require.NoError(t, err)

	format := DetectAudioFormat(audio, ".wav")

	assert.Equal(t, ".amr", format.Extension)
	assert.Equal(t, "audio/amr", format.MIMEType)
	position, err := audio.Seek(0, io.SeekCurrent)
	require.NoError(t, err)
	assert.Equal(t, int64(3), position)
}

func TestGetAudioDurationParsesAMRWithIncorrectWAVExtension(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		frames   []byte
		expected float64
	}{
		{
			name:     "AMR-NB",
			header:   "#!AMR\n",
			frames:   append(append([]byte{0x04}, make([]byte, 12)...), append([]byte{0x3c}, make([]byte, 31)...)...),
			expected: 0.04,
		},
		{
			name:     "AMR-WB",
			header:   "#!AMR-WB\n",
			frames:   append(append([]byte{0x04}, make([]byte, 17)...), append([]byte{0x4c}, make([]byte, 5)...)...),
			expected: 0.04,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			audio := bytes.NewReader(append([]byte(test.header), test.frames...))

			duration, err := GetAudioDuration(context.Background(), audio, ".wav")

			require.NoError(t, err)
			assert.InDelta(t, test.expected, duration, 0.000001)
		})
	}
}

func TestGetAudioDurationRejectsTruncatedAMRFrame(t *testing.T) {
	audio := bytes.NewReader(append([]byte("#!AMR\n"), append([]byte{0x0c}, make([]byte, 12)...)...))

	_, err := GetAudioDuration(context.Background(), audio, ".wav")

	require.ErrorContains(t, err, "truncated amr frame")
}
