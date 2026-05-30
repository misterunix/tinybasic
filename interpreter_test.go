package tinybasic

import (
	"strings"
	"testing"
)

func newTestInterp() (*Interpreter, *strings.Builder) {
	interp := New()
	var buf strings.Builder
	interp.SetOutputCallback(func(s string) { buf.WriteString(s) })
	return interp, &buf
}

func TestNew(t *testing.T) {
	interp := New()
	if interp == nil {
		t.Fatal("New() returned nil")
	}
	if interp.variables == nil {
		t.Error("variables map not initialized")
	}
	if interp.labels == nil {
		t.Error("labels map not initialized")
	}
}

func TestLoadProgram_Empty(t *testing.T) {
	interp, _ := newTestInterp()
	if err := interp.LoadProgram(""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(interp.program) != 0 {
		t.Errorf("expected 0 statements, got %d", len(interp.program))
	}
}

func TestLoadProgram_REMSkipped(t *testing.T) {
	interp, _ := newTestInterp()
	src := "REM this is a comment\nEND"
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// REM line is skipped during load; only END is in program
	if len(interp.program) != 1 {
		t.Errorf("expected 1 statement, got %d", len(interp.program))
	}
}

func TestLoadProgram_LabelRegistered(t *testing.T) {
	interp, _ := newTestInterp()
	src := "START: END"
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := interp.labels["START"]; !ok {
		t.Error("label START not registered")
	}
}

func TestPrintHello(t *testing.T) {
	interp, buf := newTestInterp()
	src := `PRINT "Hello, World!"
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	if got := buf.String(); got != "Hello, World!\n" {
		t.Errorf("expected %q, got %q", "Hello, World!\n", got)
	}
}

func TestLetAndPrint(t *testing.T) {
	interp, buf := newTestInterp()
	src := `LET X = 42
PRINT X
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	if got := buf.String(); got != "42\n" {
		t.Errorf("expected %q, got %q", "42\n", got)
	}
}

func TestImplicitLet(t *testing.T) {
	interp, buf := newTestInterp()
	src := `X = 7
PRINT X
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	if got := buf.String(); got != "7\n" {
		t.Errorf("expected %q, got %q", "7\n", got)
	}
}

func TestArithmetic(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		{"2 + 3", "5\n"},
		{"10 - 4", "6\n"},
		{"3 * 4", "12\n"},
		{"10 / 4", "2.5\n"},
		{"2 ^ 3", "8\n"},
	}
	for _, tc := range tests {
		interp, buf := newTestInterp()
		src := "PRINT " + tc.expr + "\nEND"
		if err := interp.LoadProgram(src); err != nil {
			t.Fatalf("expr %q: load error: %v", tc.expr, err)
		}
		if err := interp.Run(); err != nil {
			t.Fatalf("expr %q: run error: %v", tc.expr, err)
		}
		if got := buf.String(); got != tc.want {
			t.Errorf("expr %q: expected %q, got %q", tc.expr, tc.want, got)
		}
	}
}

func TestGotoLabel(t *testing.T) {
	interp, buf := newTestInterp()
	src := `GOTO SKIP
PRINT "should not print"
SKIP: PRINT "jumped"
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	if got := buf.String(); got != "jumped\n" {
		t.Errorf("expected %q, got %q", "jumped\n", got)
	}
}

func TestGosubReturn(t *testing.T) {
	interp, buf := newTestInterp()
	src := `GOSUB MYSUB
PRINT "back"
END
MYSUB: PRINT "sub"
RETURN`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	want := "sub\nback\n"
	if got := buf.String(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestForNextLoop(t *testing.T) {
	interp, buf := newTestInterp()
	src := `FOR I = 1 TO 3
PRINT I
NEXT I
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	want := "1\n2\n3\n"
	if got := buf.String(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestForNextStep(t *testing.T) {
	interp, buf := newTestInterp()
	src := `FOR I = 0 TO 6 STEP 2
PRINT I
NEXT I
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	want := "0\n2\n4\n6\n"
	if got := buf.String(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestIfThen(t *testing.T) {
	interp, buf := newTestInterp()
	src := `LET X = 5
IF X > 3 THEN DONE
PRINT "not reached"
DONE: PRINT "done"
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	if got := buf.String(); got != "done\n" {
		t.Errorf("expected %q, got %q", "done\n", got)
	}
}

func TestIfThenFalse(t *testing.T) {
	interp, buf := newTestInterp()
	src := `LET X = 1
IF X > 3 THEN SKIP
PRINT "not skipped"
SKIP: END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	if got := buf.String(); got != "not skipped\n" {
		t.Errorf("expected %q, got %q", "not skipped\n", got)
	}
}

func TestDataRead(t *testing.T) {
	interp, buf := newTestInterp()
	src := `DATA 10, 20, 30
READ A, B, C
PRINT A
PRINT B
PRINT C
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	want := "10\n20\n30\n"
	if got := buf.String(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestInput(t *testing.T) {
	interp, buf := newTestInterp()
	inputs := []string{"42"}
	idx := 0
	interp.SetInputCallback(func() (string, error) {
		s := inputs[idx]
		idx++
		return s, nil
	})
	src := `INPUT X
PRINT X
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	if got := buf.String(); got != "42\n" {
		t.Errorf("expected %q, got %q", "42\n", got)
	}
}

func TestVariablesAreCaseInsensitive(t *testing.T) {
	interp, buf := newTestInterp()
	inputs := []string{"9"}
	idx := 0
	interp.SetInputCallback(func() (string, error) {
		s := inputs[idx]
		idx++
		return s, nil
	})
	src := `let x = 3
PRINT X
INPUT y
PRINT Y
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	if got := buf.String(); got != "3\n9\n" {
		t.Errorf("expected %q, got %q", "3\n9\n", got)
	}
	if val, ok := interp.GetVariable("x"); !ok || val != 3 {
		t.Errorf("expected x to be 3, got %v, %v", val, ok)
	}
	if val, ok := interp.GetVariable("Y"); !ok || val != 9 {
		t.Errorf("expected Y to be 9, got %v, %v", val, ok)
	}
}

func TestPortAccessors(t *testing.T) {
	interp, buf := newTestInterp()
	interp.SetPort(7, 1234.5)

	if value, ok := interp.GetPort(7); !ok || value != 1234.5 {
		t.Fatalf("expected seeded port value 1234.5, got %g, %v", value, ok)
	}

	src := `PRINT IN(7)
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	if got := buf.String(); got != "1234.5\n" {
		t.Fatalf("expected %q, got %q", "1234.5\n", got)
	}
}

func TestPortCallbacks(t *testing.T) {
	interp, buf := newTestInterp()
	var wrotePort uint
	var wroteValue float64

	interp.SetPort(5, 11.25)
	interp.SetPortReadCallback(func(port uint) (float64, bool) {
		if port == 5 {
			return 99.75, true
		}
		return 0, false
	})
	interp.SetPortWriteCallback(func(port uint, value float64) {
		wrotePort = port
		wroteValue = value
	})

	src := `PRINT IN(5)
RESULT = OUT(9, 77.5)
PRINT RESULT
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}

	if got := buf.String(); got != "99.75\n77.5\n" {
		t.Fatalf("expected %q, got %q", "99.75\n77.5\n", got)
	}
	if wrotePort != 9 || wroteValue != 77.5 {
		t.Fatalf("expected write callback to capture port 9 value 77.5, got port %d value %g", wrotePort, wroteValue)
	}
	if value, ok := interp.GetPort(9); !ok || value != 77.5 {
		t.Fatalf("expected port 9 to store 77.5, got %g, %v", value, ok)
	}
}

func TestOutStatement(t *testing.T) {
	interp, _ := newTestInterp()
	var wrotePort uint
	var wroteValue float64

	interp.SetPortWriteCallback(func(port uint, value float64) {
		wrotePort = port
		wroteValue = value
	})

	src := `OUT 12, 34.5
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}

	if wrotePort != 12 || wroteValue != 34.5 {
		t.Fatalf("expected write callback to capture port 12 value 34.5, got port %d value %g", wrotePort, wroteValue)
	}
	if value, ok := interp.GetPort(12); !ok || value != 34.5 {
		t.Fatalf("expected port 12 to store 34.5, got %g, %v", value, ok)
	}
}

func TestOutStatementRequiresPortAndValue(t *testing.T) {
	interp, _ := newTestInterp()
	src := `OUT 1
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err == nil {
		t.Fatal("expected error for OUT without value, got nil")
	}
}

func TestInFunction(t *testing.T) {
	interp, buf := newTestInterp()
	interp.SetPort(3, 88.25)

	src := `VALUE = IN(3)
PRINT VALUE
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}

	if got := buf.String(); got != "88.25\n" {
		t.Fatalf("expected %q, got %q", "88.25\n", got)
	}
}

func TestInFunctionUsesReadCallback(t *testing.T) {
	interp, buf := newTestInterp()
	interp.SetPort(3, 10)
	interp.SetPortReadCallback(func(port uint) (float64, bool) {
		if port == 3 {
			return 44.5, true
		}
		return 0, false
	})

	src := `VALUE = IN(3)
PRINT VALUE
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}

	if got := buf.String(); got != "44.5\n" {
		t.Fatalf("expected %q, got %q", "44.5\n", got)
	}
}

func TestInFunctionStatementFormIsUnknownCommand(t *testing.T) {
	interp, _ := newTestInterp()
	src := `IN 1
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err == nil {
		t.Fatal("expected error for IN statement form, got nil")
	}
}

func TestReset(t *testing.T) {
	interp, _ := newTestInterp()
	interp.variables["X"] = 99
	interp.currentStmt = 5
	interp.done = true
	interp.Reset()
	if interp.variables["X"] != 0 {
		t.Error("Reset did not clear variables")
	}
	if interp.currentStmt != 0 {
		t.Error("Reset did not reset currentStmt")
	}
	if interp.done {
		t.Error("Reset did not clear done flag")
	}
}

func TestGotoUndefinedLabel(t *testing.T) {
	interp, _ := newTestInterp()
	src := `GOTO NOPE
END`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err == nil {
		t.Error("expected error for undefined label, got nil")
	}
}

func TestReturnWithoutGosub(t *testing.T) {
	interp, _ := newTestInterp()
	src := `RETURN`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err == nil {
		t.Error("expected error for RETURN without GOSUB, got nil")
	}
}

func TestUnknownCommand(t *testing.T) {
	interp, _ := newTestInterp()
	src := `FOOBAR`
	if err := interp.LoadProgram(src); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := interp.Run(); err == nil {
		t.Error("expected error for unknown command, got nil")
	}
}

func TestIsValidLabel(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"START", true},
		{"Loop1", true},
		{"my_label", true},
		{"", false},
		{"1bad", false},
		{"bad-label", false},
	}
	for _, tc := range tests {
		if got := isValidLabel(tc.s); got != tc.want {
			t.Errorf("isValidLabel(%q) = %v, want %v", tc.s, got, tc.want)
		}
	}
}

func TestTokenize_String(t *testing.T) {
	tokens, err := tokenize(`"hello"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 || tokens[0].Type != TokenString || tokens[0].Value != "hello" {
		t.Errorf("unexpected tokens: %v", tokens)
	}
}

func TestTokenize_UnterminatedString(t *testing.T) {
	_, err := tokenize(`"hello`)
	if err == nil {
		t.Error("expected error for unterminated string")
	}
}

func TestTokenize_Number(t *testing.T) {
	tokens, err := tokenize("3.14")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 || tokens[0].Type != TokenNumber || tokens[0].Value != "3.14" {
		t.Errorf("unexpected tokens: %v", tokens)
	}
}

func TestTokenize_Operators(t *testing.T) {
	tokens, err := tokenize("<= >= <>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ops := []string{"<=", ">=", "<>"}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	for i, op := range ops {
		if tokens[i].Value != op {
			t.Errorf("token[%d]: expected %q, got %q", i, op, tokens[i].Value)
		}
	}
}

func TestTokenize_UnexpectedChar(t *testing.T) {
	_, err := tokenize("@bad")
	if err == nil {
		t.Error("expected error for unexpected character")
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}
}
