package connection

import "math/rand"

// We don't want any capitals, so only lowercase and numbers
var chars = "abcdefghijklmnopqrstuvwxyz1234567890"

func randomString(n int) string {
	ll := len(chars)
	b := make([]byte, n)
	rand.Read(b) // generates len(b) random bytes
	for i := 0; i < n; i++ {
		b[i] = chars[int(b[i])%ll]
	}
	return string(b)
}

func (s *service) generateConnectionID(n int) string {
	var id string
	for i := 0; i < 5; i++ {
		id = randomString(n)
		if _, ok := s.connections[id]; !ok {
			break
		}
	}

	return id
}
