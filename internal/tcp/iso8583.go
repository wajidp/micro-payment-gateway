package tcp

// ParseISO8583Message parses a raw ISO8583 message
// TODO: Implement the actual parsing logic to extract the user ID and amount from the ISO8583 message
func parseISO8583Message(rawMessage []byte) (string, int64, error) {

	amt := int64(0)
	user := ""
	// parse ISO8583 & get the amount, userId
	return user, amt, nil
}
