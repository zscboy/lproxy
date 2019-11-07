package server

import (
	"testing"
)

func BenchmarkTokDecode(b *testing.B) {
	uuid := "738b935b-e5c9-44b0-8524-290146ec08e6"
	token := GenTK(uuid)

	for i := 0; i < b.N; i++ {
		parseTK(token)
	}
}

func BenchmarkTokEncode(b *testing.B) {
	uuid := "738b935b-e5c9-44b0-8524-290146ec08e6"
	for i := 0; i < b.N; i++ {
		GenTK(uuid)
	}
}
