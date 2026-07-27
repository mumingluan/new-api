package relayconvert

import sharedgemini "github.com/QuantumNous/new-api/service/relayconvert/internal/shared/gemini"

// CleanGeminiTools sanitizes native Gemini function schemas before they are
// serialized for the upstream API.
func CleanGeminiTools(tools []byte) ([]byte, error) {
	return sharedgemini.CleanTools(tools)
}
