package factory

import "testing"

func TestEnsurePostTextAttributionPreservesMentionAfterPodcastTag(t *testing.T) {
	dict := ParseMentionDictionary(`{
		"podcasts": [
			{"name": "Silicon Valley Girl", "handle": "siliconvalleymm"}
		]
	}`)
	inputs := []string{
		"A sharp take.\n\n@siliconvalleymm @AnthropicAI",
		"A sharp take.\n\n@siliconvalleymm\n@AnthropicAI",
	}
	for _, input := range inputs {
		got := EnsurePostTextAttribution(input, "Silicon Valley Girl", dict)
		if got != input {
			t.Errorf("mention after podcast tag did not round-trip:\nwant %q\n got %q", input, got)
		}
	}
}
