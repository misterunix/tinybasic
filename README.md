# TinyBASIC 0.0.1-alpha

A label-based BASIC interpreter written in Go with floating-point support and extensive math functions. Designed as a module for embedding in other Go programs.

## Status

Not fully tested yet.

[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go)](https://go.dev/)

## Features

- **Label-based control flow** - Use meaningful labels instead of line numbers
- **Floating-point arithmetic** - Full support for floating-point numbers and operations
- **Extensive math functions** - Trigonometric, logarithmic, and utility functions
- **Module design** - Easy to embed in other Go programs
- **Step and Run modes** - Execute programs statement-by-statement or all at once
- **Callback system** - Custom input/output handlers for integration
- **Optional LET** - Modern syntax with optional LET keyword
- **Standard BASIC statements** - FOR/NEXT loops, GOSUB/RETURN, IF/THEN, and more

## Installation

```bash
go get github.com/misterunix/tinybasic
```

## Quick Start

```go
package main

import (
    "fmt"
    "tinybasic"
)

func main() {
    // Create a new interpreter
    interp := tinybasic.New()

    // Load a BASIC program
    program := `
        X = 10
        Y = 20
        RESULT = X + Y
        PRINT "The sum is:", RESULT
    `
    
    if err := interp.LoadProgram(program); err != nil {
        fmt.Println("Error:", err)
        return
    }

    // Run the program
    if err := interp.Run(); err != nil {
        fmt.Println("Runtime error:", err)
    }
}
```

## API Reference

### Creating an Interpreter

```go
interp := tinybasic.New()
```

### Loading a Program

```go
err := interp.LoadProgram(source string)
```

### Execution Control

```go
// Run the entire program
err := interp.Run()

// Execute a single statement (for debugging or interactive mode)
err := interp.Step()

// Check if execution is complete
done := interp.IsDone()

// Reset interpreter state for a new run
interp.Reset()
```

### Variable Access

```go
// Get a variable value
value, exists := interp.GetVariable("X")

// Set a variable value
interp.SetVariable("X", 42.5)

// Get all variables
vars := interp.GetAllVariables()
```

### Program Information

```go
// Get current statement index
line := interp.GetCurrentLine()

// Get total number of statements
size := interp.GetProgramSize()
```

### Port Access

```go
// Seed a port value that BASIC can read with IN(port)
interp.SetPort(100, 42.5)

// Read the last stored value for a port
value, exists := interp.GetPort(100)
```

### Custom I/O Callbacks

```go
// Set input callback
interp.SetInputCallback(func() (string, error) {
    var input string
    fmt.Scanln(&input)
    return input, nil
})

// Set output callback
interp.SetOutputCallback(func(s string) {
    fmt.Print(s)
})

// Override IN(port) reads from the parent Go program.
// Return ok=false to fall back to the interpreter's stored port map.
interp.SetPortReadCallback(func(port uint) (float64, bool) {
    if port == 7 {
        return 1234.5, true
    }
    return 0, false
})

// Observe or mirror OUT(port, value) writes.
interp.SetPortWriteCallback(func(port uint, value float64) {
    fmt.Printf("port %d <- %g\n", port, value)
})
```

## BASIC Language Reference

### Variable Assignment

```basic
X = 10
LET Y = 20.5          REM LET is optional
RESULT = X + Y * 2
```

### Labels

Labels provide named points in your program for jumps and subroutines:

```basic
START: PRINT "Beginning program"
       X = 1

LOOP:  PRINT X
       X = X + 1
       IF X <= 10 THEN LOOP
```

### Operators

- Arithmetic: `+`, `-`, `*`, `/`, `^` (power)
- Comparison: `=`, `<>`, `<`, `>`, `<=`, `>=`
- Parentheses for grouping: `(`, `)`

### Math Functions

#### Trigonometric Functions

- `SIN(x)` - Sine
- `COS(x)` - Cosine
- `TAN(x)` - Tangent
- `ASIN(x)` - Arcsine
- `ACOS(x)` - Arccosine
- `ATAN(x)` - Arctangent

#### Logarithmic and Exponential

- `LOG(x)` - Base-10 logarithm
- `LN(x)` - Natural logarithm
- `EXP(x)` - e^x
- `SQRT(x)` - Square root

#### Utility Functions

- `ABS(x)` - Absolute value
- `INT(x)` - Convert to integer (truncate)
- `FLOOR(x)` - Floor function
- `CEIL(x)` - Ceiling function
- `ROUND(x)` - Round to nearest integer
- `SGN(x)` - Sign (-1, 0, or 1)
- `RND(x)` - Pseudo-random number

#### Multi-argument Functions

- `POW(x, y)` - x to the power of y
- `MIN(x, y)` - Minimum of two values
- `MAX(x, y)` - Maximum of two values

#### I/O Functions

- `IN(port)` - Read a floating-point value from the specified port (returns 0 if port is uninitialized)
- `OUT(port, value)` - Write a floating-point value to the specified port and return the value
  - Port: unsigned integer (0 to max uint)
    - Value: floating-point number

From Go, you can seed ports with `SetPort`, inspect them with `GetPort`, override reads with `SetPortReadCallback`, and observe writes with `SetPortWriteCallback`.

### Statements

#### PRINT - Output

```basic
PRINT "Hello, World!"
PRINT "X =", X
PRINT X, Y, Z
```

#### INPUT - Input

```basic
INPUT X
INPUT "Enter a number:", Y
```

#### IF...THEN - Conditional

```basic
IF X > 10 THEN PRINT "X is large"
IF X = Y THEN DONE
```

#### GOTO - Jump to Label

```basic
GOTO START
GOTO LOOP
```

#### GOSUB...RETURN - Subroutine

```basic
GOSUB CALCULATION
PRINT RESULT
GOTO DONE

CALCULATION: RESULT = X * 2
             RETURN

DONE: END
```

#### FOR...TO...STEP...NEXT - Loops

```basic
FOR I = 1 TO 10
    PRINT I
NEXT I

FOR X = 10 TO 0 STEP -1
    PRINT X
NEXT X

FOR J = 0 TO 100 STEP 5
    PRINT J
NEXT J
```

#### DATA...READ - Static Data

```basic
READ X, Y, Z
PRINT X + Y + Z
DATA 10, 20, 30
```

#### REM - Comments

```basic
REM This is a comment
PRINT "Hello"  REM Comment on same line not supported
```

#### END - End Program

```basic
END
```

## Example Programs

### Calculate Circle Area

```basic
PI = 3.14159265359

INPUT "Enter radius:", R
AREA = PI * R ^ 2
CIRCUMFERENCE = 2 * PI * R

PRINT "Area:", AREA
PRINT "Circumference:", CIRCUMFERENCE
END
```

### Factorial Calculator

```basic
INPUT "Enter a number:", N
RESULT = 1
I = 1

LOOP: IF I > N THEN DONE
      RESULT = RESULT * I
      I = I + 1
      GOTO LOOP

DONE: PRINT "Factorial:", RESULT
      END
```

### Fibonacci Sequence

```basic
N = 10
A = 0
B = 1

PRINT "Fibonacci sequence:"
FOR I = 1 TO N
    PRINT A
    TEMP = A + B
    A = B
    B = TEMP
NEXT I
END
```

### Trigonometry Table

```basic
PRINT "Angle", "Sin", "Cos", "Tan"

FOR ANGLE = 0 TO 90 STEP 15
    RAD = ANGLE * 3.14159 / 180
    S = SIN(RAD)
    C = COS(RAD)
    T = TAN(RAD)
    PRINT ANGLE, S, C, T
NEXT ANGLE
END
```

### Temperature Converter

```basic
START: INPUT "Celsius:", C
       F = C * 9 / 5 + 32
       K = C + 273.15
       
       PRINT "Fahrenheit:", F
       PRINT "Kelvin:", K
       
       INPUT "Convert another? (1=yes, 0=no):", AGAIN
       IF AGAIN = 1 THEN START
       END
```

### Port I/O Operations

```basic
REM Demonstrate IN and OUT port I/O functions

REM Write values to various ports
RESULT = OUT(100, 42.5)
PRINT "Wrote 42.5 to port 100, returned:", RESULT

RESULT = OUT(101, -255.5)
PRINT "Wrote -255.5 to port 101"

RESULT = OUT(102, 1000.25)
PRINT "Wrote 1000.25 to port 102"

REM Read values back from ports
VAL1 = IN(100)
VAL2 = IN(101)
VAL3 = IN(102)

PRINT "Read from port 100:", VAL1
PRINT "Read from port 101:", VAL2
PRINT "Read from port 102:", VAL3

REM Reading from uninitialized port returns 0
VAL4 = IN(999)
PRINT "Read from port 999:", VAL4

REM Use ports for data exchange
FOR I = 0 TO 9
    DUMMY = OUT(200 + I, I * I)
NEXT I

PRINT "Squares stored in ports 200-209:"
FOR I = 0 TO 9
    VALUE = IN(200 + I)
    PRINT "Port", 200 + I, "=", VALUE
NEXT I

END
```

### Go Example for Port I/O

```go
package main

import (
    "fmt"

    "github.com/misterunix/tinybasic"
)

func main() {
    interp := tinybasic.New()

    program := `
RESULT = OUT(100, 42.5)
PRINT "OUT returned:", RESULT

VALUE = IN(100)
PRINT "IN read:", VALUE

MISSING = IN(999)
PRINT "Uninitialized port:", MISSING
END
`

    if err := interp.LoadProgram(program); err != nil {
        fmt.Println("Load error:", err)
        return
    }

    if err := interp.Run(); err != nil {
        fmt.Println("Runtime error:", err)
        return
    }
}
```

Expected output:

```text
OUT returned: 42.5
IN read: 42.5
Uninitialized port: 0
```

### Go Example with Parent-Supplied Port Values

```go
package main

import (
    "fmt"

    "github.com/misterunix/tinybasic"
)

func main() {
    interp := tinybasic.New()

    interp.SetPort(7, 1234.5)
    interp.SetPortReadCallback(func(port uint) (float64, bool) {
        if port == 8 {
            return 5678.25, true
        }
        return 0, false
    })
    interp.SetPortWriteCallback(func(port uint, value float64) {
        fmt.Printf("BASIC wrote port %d = %g\n", port, value)
    })

    program := `
PRINT IN(7)
PRINT IN(8)
RESULT = OUT(9, 4321.75)
PRINT RESULT
END
`

    if err := interp.LoadProgram(program); err != nil {
        fmt.Println("Load error:", err)
        return
    }

    if err := interp.Run(); err != nil {
        fmt.Println("Runtime error:", err)
        return
    }

    if value, ok := interp.GetPort(9); ok {
        fmt.Println("Stored port 9:", value)
    }
}
```

Expected output:

```text
1234.5
5678.25
BASIC wrote port 9 = 4321.75
4321.75
Stored port 9: 4321.75
```

## Integration Example

Here's a complete example showing how to embed the interpreter with custom I/O:

```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    "tinybasic"
)

func main() {
    interp := tinybasic.New()
    
    // Custom input handler
    reader := bufio.NewReader(os.Stdin)
    interp.SetInputCallback(func() (string, error) {
        fmt.Print("? ")
        input, err := reader.ReadString('\n')
        return strings.TrimSpace(input), err
    })
    
    // Custom output handler
    var output strings.Builder
    interp.SetOutputCallback(func(s string) {
        output.WriteString(s)
        fmt.Print(s)
    })
    
    // Load program
    program := `
        INPUT "Enter your name:", NAME$
        PRINT "Hello,", NAME$
    `
    
    if err := interp.LoadProgram(program); err != nil {
        fmt.Println("Load error:", err)
        return
    }
    
    // Run program
    if err := interp.Run(); err != nil {
        fmt.Println("Runtime error:", err)
        return
    }
    
    // Access program state
    fmt.Println("\nProgram completed.")
    fmt.Println("Variables:", interp.GetAllVariables())
}
```

## Step-by-Step Debugging

```go
interp := tinybasic.New()
interp.LoadProgram(source)

// Execute one statement at a time
for !interp.IsDone() {
    fmt.Printf("Line %d: ", interp.GetCurrentLine())
    if err := interp.Step(); err != nil {
        fmt.Println("Error:", err)
        break
    }
    
    // Inspect variables after each step
    vars := interp.GetAllVariables()
    fmt.Println("Variables:", vars)
}
```

## Language Notes

- **Case Insensitive**: Commands and function names are case-insensitive
- **No Line Numbers**: Uses labels instead of traditional line numbers
- **Whitespace**: Extra whitespace is generally ignored
- **String Literals**: Must be enclosed in double quotes
- **Comments**: Use REM at the start of a line

## Limitations

- String variables are not yet supported (planned for future release)
- Arrays are not supported
- File I/O is not supported
- No string manipulation functions

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Code of Conduct

This project follows a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## License

This project is licensed under the BSD 3-Clause License - see the [LICENSE](LICENSE) file for details.

## Author

**William Jones**

See [CONTRIBUTORS.md](CONTRIBUTORS.md) for a list of contributors.

## Acknowledgments

- Built with Go's powerful standard library
