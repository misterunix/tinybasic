# TBIL -- The **TINY BASIC** Interpreter Language

The **TINY BASIC** interpreter is, in the words of Dennis Allison who conceived it, something like an onion.  There is an inner machine language program (ML) which interprets a second program written in an intermediate language (IL), which in turn interprets the BASIC program, and so on.  This document describes that intermediate language and the virtual machine which executes it.

The IL interpreter is a pure interpreter in the sense that the entire BASIC interpreter is implemented within the bounds of the language.  There are no deus ex machina escapes to machine language other than the well-defined machine-language subroutine call. The language is substantially the same as that defined by Dennis Allison in Dr. Dobb's Journal and PCC.

Most of the instructions in the IL occupy one byte of code. A few instructions may be followed by one or more bytes of immediate data and there are two jump instructions which are actually two bytes in length.

The interpreter itself uses no variable storage. Computations are performed on an expression stack, so that all procedures are capable of recursion. The interpreter does have access to memory Page 00 for data storage, but little use is made of this capability.

The IL code is self-relative. That is, all jumps are relative to the beginning of the IL interpreter. Thus the code can be moved to another part of memory without re-assembling it. The conditional branches are PC-relative and branch only forward to a maximum displacement of 31 bytes. An unconditional branch has a range of 31 bytes forward or backwards.  Since the interpreter is generally quite small this is not a serious limitation. There are only two or three places where a longer conditional branch would be needed; these are accommodated by branching to a jump. The jumps have an address space of 11 bits, or 2K bytes from the beginning of the IL code.

Two of the instructions include a literal text string as part of the code. This string follows the opcode and is of arbitrary length. The end of the string is signaled by the eighth bit on the last byte being set to one. Since the text is generally assumed to be ASCII, a 7-bit code, this is a reasonable way to save space.

There are two stacks in the virtual machine. The computational or Expression Stack has already been mentioned. The other stack is a control stack, used to hold subroutine return addresses. The same control stack is used for subroutine returns in three languages: BASIC, IL, and ML. Thus a certain amount of care is necessary in the maintenance of this stack. The ML interpreter will take care of its requirements by placing them on the stack top, and a special parameter ("SPARE") in the main program defines the maximum amount of space to be left on the stack for this purpose. Overflow of this part of the stack is not detectable, so it is essential that the reserved space be sufficiently large. Because the ML interpreter is not recursive this is reasonably safe, provided that the stack requirements of the I/O are known and limited. Beneath the ML stack is the IL stack. Subroutine calls in the IL have their return address pushed onto this stack. The ML interpreter does check for stack overflow by measuring the distance between the top of the IL stack and the end of the BASIC program; if this ever becomes less than the SPARE parameter, the stack is considered to have overflowed.

Beneath the IL stack is the BASIC stack. This holds the line numbers for GOSUB lines as they are executed. There are two instructions in the IL for accessing the top element of this stack (i.e. a Push and a Pop). For these instructions to work properly it is essential that the IL stack be empty. No check is made in the ML for this condition, and it is the responsibility of the IL program to insure this form of stack integrity. In other words, BASIC language GOSUBs and RETURNs cannot be processed within an IL subroutine.

The interpretation of the BASIC program is tied to a pointer (invisible to the IL) which points to the current character in the current line. The IL has no direct control over this pointer, but several of the IL instructions cause it to be advanced or otherwise modified. In particular it may be changed to point to the beginning of another BASIC line to implement the BASIC sequence control operations (GOTO, GOSUB, RETURN). It may also be exchanged with its logical dual, a pointer to the current character in the input line buffer. This permits the same interpreter to operate on the BASIC program stored in memory or on a direct execution statement in the line buffer.

There are a number of IL operations which may result in an error condition. All errors abort the IL execution and print out on the console the IL program counter at which the error occurred. If the program execution flag was set, the most recently accessed BASIC line number is also typed out in the message. The relative address (relative to the beginning of the IL) becomes the error number in the error message. Since only one error is possible in most operations, this gives a unique identification of the difficulty. The IL address is printed in decimal and represents the address of the next byte which would have been executed but for the error. There is one operation which has two possible failure modes; for one of these the IL address is decremented before printing to distinguish it from the other. Error stops may be explicitly requested in the IL program by the execution of a branch with a zero offset.
     
After typing out the error addresses, the ML interpreter clears the ML and IL stacks (but not the BASIC stack!) and restarts the IL interpreter at relative address 0. Nothing else is changed, except that the execution flag is cleared, putting the interpreter into the command mode. When the Break condition is recognized in advancing to the next BASIC statement this is treated by the ML as an error condition, after forcing the IL program counter to relative zero.

The ML interpreter maintains a flag to distinguish program RUN mode from direct statement execution (command mode). Advancing to the next statement (an IL instruction) examines this flag, and if in the command mode, the IL program is restarted at the beginning. If the flag is set in the RUN mode, execution resumes at the IL address saved by the Execute instruction. It is important that the Execute instruction be given in the IL before any Next BASIC statement advance, but once it has been done there is no restriction (i.e. the saved address is never lost).
     
The Break condition is tested only during the execution of the statement advance (Next), so that resumption of an interrupted program leaves no computational gaps. The Break condition is also tested during a LIST operation, but only to abort the listing; if the LIST occurred within program execution (i.e. with the RUN mode flag set), a second Break condition is required to terminate the program.

The following is a detailed description of the operation of each of the IL opcodes. With each description is also given the hexadecimal opcode and the mnemonic recognized by the assembler. Not all of the opcodes are defined. Some have been incorporated into unused functions; others are reserved for possible future expansion and execute as NOPs.


## INTERPRETIVE LANGUAGE OPERATION CODES

```SX n    00-07   Stack Exchange.```
- Exchange the top byte of computational stack with that "n" bytes into the stack. The top byte of the stack is considered to be byte 0, so SX 0 does nothing. The sequence of
instructions may be used to exchange the top two numbers (two bytes each) on the stack. Only the top eight bytes on the stack are accessible, to this instruction. If the stack is empty an error stop may or may not occur, depending on which ML interpreter is implemented.
```
SX 1
SX 3
SX 1
SX 2
```


```NO      08      No Operation.```
- This may be used as a space filler (such as to ignore a skip).

```LB n    09nn    Push Literal Byte onto Stack.```
- This adds one byte to the computational stack, which is the second byte of the instruction. An error stop will occur if the stack overflows.

```LN n    0Annnn  Push Literal Number.```
- This adds the following two bytes to the computational stack, as a 16-bit number. Stack overflow results in an error stop.

```DS      0B      Duplicate Top Number (two bytes) on Stack.```
- An error stop will occur if there are less than two bytes on the expression stack or if the stack overflows.

```SP      0C      Stack Pop.```
- The top two bytes are removed from the computational stack and discarded. Underflow results in an error stop.

```SB      10      Save BASIC Pointer.```
- If the BASIC pointer is pointing into the input line buffer, it is copied to the Saved Pointer; otherwise the two pointers are exchanged.

```RB      11      Restore BASIC Pointer.```
- If the Saved Pointer is pointing into the input line buffer, it is replaced by the value in the BASIC pointer; otherwise the two pointers are exchanged.
- Normally the Saved Pointer will point to the next item in the  input line buffer while the BASIC pointer points to the program  being executed. When an INPUT instruction in BASIC is interpreted,  the two pointers are exchanged by the SB opcode so that the  expression handling capabilities of the interpreter may be applied  to the input data, then the pointers are restored (exchanged again)  by the RB. In direct execution (command mode) the BASIC pointer is already in the input line buffer, and the contents of the Saved  Pointer are meaningless; in this case the SB instruction does not  alter the BASIC pointer, and the RB opcode should leave both  pointers pointing to the next item in the input string.

```FV      12      Fetch Variable.```
- The top byte of the computational stack is used to index into Page 00. It is replaced by the two bytes fetched. Error stops occur with stack overflow or underflow.

```SV      13      Store Variable.```
- The top two bytes of the computational stack are stored into memory at the Page 00 address specified by the third byte on the stack. All three bytes are deleted from the stack. Underflow results in an error stop.

```GS      14      GOSUB Save.```
- The line number on the current BASIC line is pushed onto the BASIC region of the control stack. It is essential that the IL stack be empty for this to work properly but no check is made for that condition. An error stop occurs on stack overflow.

```RS      15      Restore Saved Line.```
- Pop the top two bytes off the BASIC region of the control stack, making them the current line number. Set the BASIC pointer at the beginning of that line. Note that this is the line containing the GOSUB which caused the line number to be saved. As with the GS opcode, it is essential that the IL region of the control stack be empty. If the line number popped off the stack does not correspond to a line in the BASIC program an error stop occurs. An error stop also results from stack underflow.

```GO      16      GOTO.```
- Make current the BASIC line whose line number is equal to the value of the top two bytes in the expression stack. That is, the top two bytes are popped off the computational stack, and the BASIC program is searched until a matching line number is found. The BASIC pointer is then positioned at the beginning of that line and the RUN mode flag is turned on. Stack underflow and non-existent BASIC line result in error stops.

```NE      17      Negate (two's complement).```
- The number in the top two bytes of the expression stack is replaced with its negative.

```AD      18      Add.```
- Add the two numbers represented by the top four bytes of the expression stack, and replace them with the two-byte sum. Stack underflow results in an error stop.

```SU      19      Subtract.```
- Subtract the two-byte number on the top of the expression stack from the next two bytes and replace the four bytes with the two-byte difference. This is exactly equivalent to the two-instruction sequence and has the same error stop on underflow.

      NE
      AD

```MP      1A      Multiply.```
- Multiply the two numbers represented by the top four bytes of the computational stack, and replace them with the least significant 16 bits of the product. Stack underflow is possible.

```DV      1B      Divide.```
- Divide the number represented by the top two bytes of the computational stack into that represented by the next two. Replace the four bytes with the quotient and discard the remainder. This is a signed (two's complement) integer divide, resulting in a signed integer quotient. Stack underflow or attempted division by zero result in an error stop.

```CP      1C      Compare.```
- The number in the top two bytes of the expression stack is compared to (subtracted from) the number in the fourth and fifth bytes of the stack, and the result is determined to be Greater, Equal, or Less. The low three bits of the third byte mask a conditional skip in the IL program to test these conditions; if the result corresponds to a one bit, the next byte of the IL code is skipped and not executed. The three bits correspond to the conditions as follows:
```
bit 0   Result is Less
bit 1   Result is Equal
bit 2   Result is Greater
```
Whether the skip is taken or not, all five bytes are deleted from the stack. This is a signed (two's complement) comparison so that any positive number is greater than any negative number. Multiple conditions, such as greater-than-or-equal or unequal (i.e. greater- than-or-less-than), may be tested by forming the condition mask byte of the sum of the respective bits. In particular, a mask byte of 7 will force an unconditional skip and a mask byte of 0 will force no skip. The other five bits of the control byte are ignored. Stack underflow results in an error stop.

```NX      1D      Next BASIC Statement.```
- Advance to the next line in the BASIC program, if in the RUN mode, or restart the IL program if in the command mode. The remainder of the current line is ignored. In the Run mode if there is another line it becomes current with the pointer positioned at its beginning. At this time, if the Break condition returns true, execution is aborted and the IL program is restarted after printing an error message. Otherwise IL execution proceeds from the saved IL address (see the XQ instruction). If there are no more BASIC statements in the program an error stop occurs.

```LS      1F      List The Program.```
- The expression stack is assumed to have two 2-byte numbers. The top number is the line number of the last line to be listed, and the next is the line number of the first line to be listed. If the specified line numbers do not exist in the program, the next available line (i.e. with the next higher line number) is assumed instead in each case. If the last line to be listed comes before the first, no lines are listed. If the Break condition comes true during a List operation, the remainder of the listing is aborted. Zero is not a valid line number, and an error stop occurs if either line number specification is zero. The line number specifications are deleted from the stack.

```PN      20      Print Number.```
- The number represented by the top two bytes of the expression stack is printed in decimal with leading zero suppression. If it is negative, it is preceded by a minus sign (hyphen) and the magnitude is printed. Stack underflow is possible.

```PQ      21      Print BASIC String.```
- The ASCII characters beginning with the current position of the BASIC pointer are printed on the console. The string to be printed is terminated by the quotation mark ("), and the BASIC pointer is left at the character following the terminal quote. An error stop occurs if a carriage return is imbedded in the string.

```PT      22      Print Tab.```
- Print one or more spaces on the console, ending at the next multiple of eight character positions (from the left margin).

```NL      23      New Line.```
- Output a carriage-return-linefeed sequence to the console.

```PC "xxxx"  24xxxxxxXx   Print Literal String.```
- The ASCII string follows the opcode and its last byte has the most significant bit set to one. The character string is output to the console unmodified; that is, all eight bits of each byte is output, so that the last byte and only that byte is output with the parity bit set to one. This of course may be altered by the output routine.

```GL      27      Get Input Line.```
- ASCII characters are accepted from the console input to fill the line buffer. If the line length exceeds the available space, the excess characters are ignored and bell characters are output. The line is terminated by a carriage return. NUL and DEL codes (hex 00 and FF) are ignored; linefeed and DC3 respectively turn the "tape mode" on and off. Any characters which match the Backspace parameter result in the deletion of the previous character in the line buffer, if any; if the line buffer is empty the effect is that of a cancel. Any character which matches the Cancel parameter stores a carriage return in the first position of the line buffer and terminates the input. On completing one line of input, the BASIC pointer is set to point to the first character in the input line buffer, and a carriage-return-linefeed sequence is output.

```IL      2A      Insert BASIC Line.```
- Beginning with the current position of the BASIC pointer and continuing to the carriage return, the line is inserted into the BASIC program space; for a line number, the top two bytes of the expression stack are used. If this number matches a line already in the program it is deleted and the new one replaces it. If the new line consists of only a carriage return, it is not inserted, though any previous line with the same number will have been deleted. The lines are maintained in the program space sorted by line number. If the new line to be inserted is a different size than the old line being replaced, the remainder of the program is shifted over to make room for it or to close up the gap as necessary. If there is insufficient memory to fit in the new line, the program space is unchanged and an error stop occurs (with the IL address decremented). A normal error stop occurs on expression stack underflow or if the number is zero, which is not a valid line number. After completing the insertion, the IL program is restarted in the command mode.

```MT      2B      Mark the BASIC program space Empty.```
- Also clears the BASIC region of the control stack and restart the IL program in the command mode. The memory bounds and stack pointers are reset by this instruction to signify an empty program space, and the line number of the first line is set to zero, which is the indication of the end of the program. The remainder of the program is not altered, though it is now vulnerable to intrusion by the control stack. The program may be recovered if accidentally CLEARed by storing a non-zero line number in the first two bytes of the BASIC program space, then requesting a LIST. If this is made on a machine-readable medium, it may be reloaded. Any execution of the IL instruction after a MT instruction will destroy the contents of memory not enclosed by the program bounds in locations 0020-0025.

```XQ      2C      Execute.```
- Turns on RUN mode. This instruction also saves the current value of the IL program counter for use of the NX instruction, and sets the BASIC pointer to the beginning of the BASIC program space. An error stop occurs if there is no BASIC program. This instruction must be executed at least once before the first execution of a NX instruction.

```WS      2D      Stop.```
- Stop execution and restart the IL program in the command mode. The entire control stack (including the BASIC region) is also vacated by this instruction. This instruction effectively jumps to the Warm Start entry of the ML interpreter.

```US      2E      Machine Language Subroutine Call.```
- The top six bytes of the expression stack contain three numbers with the following interpretations: The top number is loaded into the A (or A and B) register; the next number is loaded into 16 bits of Index register; the third number is interpreted as the address of a machine language subroutine to be called using the normal subroutine call sequence (which is simulated for this purpose by the ML interpreter). These six bytes on the expression stack are replaced with the 16-bit result returned by the subroutine. Stack underflow results in an error stop.

```RT      2F      IL Subroutine Return.```
- The IL control stack is popped to give the address of the next IL instruction. An error stop occurs if the entire control stack (IL and BASIC) is empty.

```JS a    3000-37FF       IL Subroutine Call.```
- The least significant eleven bits of this 2-byte instruction are added to the base address of the IL program to become the address of the next instruction. The previous contents of the IL program counter are pushed onto the IL region of the control stack. Stack overflow results in an error stop.

```J a     3800-3FFF       Jump.```
- The low eleven bits of this 2-byte instruction are added to the IL program base address to determine the address of the next IL instruction. The previous contents of the IL program counter is lost.

```BR a    40-7F   Relative Branch.```
- The low six bits of this instruction opcode are added algebraically to the current value of the IL program counter to give the address of the next IL instruction. Bit 5 of the opcode is the sign, with + signified by 1, - by 0. The range of this branch is 31 bytes from address of the byte following the opcode, in either direction. An offset of zero (i.e. opcode 60) results in an error stop. The branch operation is unconditional.

```BC a "xxx"   80xxxxXx-9FxxxxXx  String Match Branch.```
- The ASCII character string in the IL following this opcode is compared to the string beginning with the current position of the BASIC pointer, ignoring blanks in the BASIC program. The comparison continues until either a mismatch is found, or an IL byte is reached with the most significant bit set to one. This is the last byte of the string in the IL, and it is compared as a 7-bit character; if equal, the BASIC pointer is positioned after the last matching character in the BASIC program and the IL program continues with the next instruction in sequence. Otherwise the BASIC pointer is not altered and the low five bits of the Branch opcode are added to the IL program counter to form the address of the next IL instruction. If the strings do not match and the branch offset is zero an error stop occurs.

```BV a    A0-BF   Branch if Not Variable.```
- If the next non-blank character pointed to by the BASIC pointer is a capital letter, its ASCII code is doubled and pushed onto the expression stack and the IL program advances to the next instruction in sequence, leaving the BASIC pointer positioned after the letter; if not a letter the branch is taken and the BASIC pointer is left pointing to that character. An error stop occurs if the next character is not a letter and the offset of the branch is zero, or on stack overflow.

```BN a    C0-DF   Branch if Not a Number.```
- If the next non-blank character pointed to by the BASIC pointer is not a decimal digit, the low five bits of the opcode are added to the IL program counter, or if zero an error stop occurs. If the next character is a digit, then it and all decimal digits following it (ignoring blanks) are converted to a 16-bit binary number which is pushed onto the expression stack. In either case the BASIC pointer is positioned at the next character which is neither blank nor digit. Stack overflow will result in an error stop.

```BE a    E0-FF   Branch if Not Endline.```
- If the next non-blank character pointed to by the BASIC pointer is a carriage return, the IL program advances to the next instruction in sequence; otherwise the low five bits of the opcode (if not zero) are added to the IL program counter to form the address of the next IL instruction. In either case the BASIC pointer is left pointing to the first non-blank character encountered; this instruction will not pass over the carriage return, which must remain for testing by the `NX` instruction. As with the other conditional branches, the branch may only advance the IL program counter from 1 to 31 bytes; an offset of zero results in an error stop.



