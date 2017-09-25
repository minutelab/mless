package envstr

import "testing"

func TestEncode(t *testing.T) {
	env := map[string]string{
		"A":  "B",
		"C=": "D;",
		"C":  "=D",
		"E":  `\n`,
	}

	t.Log("Encoding:", env)
	val := Encode(env)
	t.Log("Encoded :", val)

	expected := `A=B;C=\=D;C\==D\;;E=\\n`
	if val != expected {
		t.Error("Different than expected: ", expected)
	}
}
