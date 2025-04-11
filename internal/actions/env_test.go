package actions

import "testing"

func TestEnv(t *testing.T) {
	env := Env{
		"TEST_VAR1": "114514",
		"TEST_VAR2": "123",
		"TEST_VAR3": "456",
	}

	expectedContent := "TEST_VAR1=114514\nTEST_VAR2=123\nTEST_VAR3=456"

	if env.String() != expectedContent {
		t.Errorf("unexpected content: want %s got %s", expectedContent, env.String())
	}
}
