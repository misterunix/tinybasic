# TBIL ASSEMBLER

To aid in developing and modifying the IL program an assembler has been written in **TINY BASIC**. This assembler accepts the mnemonics for the IL assembly language and outputs a hexadecimal object code suitable for loading into memory. It is a two-pass assembler, building the symbol table on the first pass and generating the full hex object code on the second pass.

Since **TINY BASIC** does not allow strings or arrays, the source file and the symbol table are manipulated using the USR function to call on the standard machine language subroutines to load and store bytes in memory. This is unfortunately very slow, so a third subroutine, which loads two bytes, is also used in an effort to speed things up a little. Comments in the source listing of the assembler indicate how such a routine may be coded. The assembler is still compute-bound, and can be expected to take several hours on each pass. This is considered acceptable only because of the infrequent need to assemble the IL code.

The assembler accepts free-form input with two kinds of source lines: Comment lines and program instruction lines. Each line of either kind must begin with a line number. This is actually a kludge to convince **TINY BASIC** to read the source line with an INPUT command, and the number has no significance to the assembler other than that it is zero on the last line of the program.

Comment lines are indicated to the assembler by a period following the line number. They are not processed further.

Instruction lines may begin with a label or not. A label is signified by a leading colon (which is not part of the label) followed by a letter and up to three more letters and/or digits, and terminated by a blank.

The next field after the label, or the first field of a line without a label, is the instruction mnemonic. This is one of the two-letter codes (or one letter in the case of J) defined earlier.

The instructions which require operands should be followed by at least one blank, then the operand in the correct format. Jumps and branches accept a label reference; the branches also accept the single symbol `*` to signify an error stop branch. The SX instruction requires a single octal digit (1-7).

The LB and LN instructions should be followed by a decimal number. This number is processed by the BASIC INPUT command which accepts expressions and ignores blanks, so care must be taken in what is allowed to follow the number. In particular it may not be followed by more decimal digits or the characters `+` `-` `*` or `/`. The number must start with a digit.

The BC and PC instructions are followed by a string (after the label in the case of BC). The string is enclosed in a pair of delimiters which may be any non-blank character except the ASCII circumflex (hex 5E which sometimes prints as an up-arrow). Any character within the string which is followed by a circumflex has a hex 40 subtracted from its code, making it possible to generate strings with control characters in them. The last character of the string has the most significant bit set to one in the object code.

Everything on the source line after the operands, if any, is treated by the assembler as comments.

The operation of the assembler is shaped by the restrictions imposed by **TINY BASIC**. The source lines must not be larger than 60 or so characters to leave room in the expression stack. Each source line must end in a DC3 control (X-OFF) unless other reader control is used, since several tens of seconds are required to process each line.

The program is loaded and started with a RUN. It will ask for the addresses of the byte load and store routines, which should be typed in in decimal. It will also ask for the memory address that the program is to load into. This address is only used in the generation of the location counter output and has no effect on the code generation.

One of the first things done in the assembler is to search for the mnemonic table, which is imbedded in pseudo-comment lines near the beginning of the assembler. These are identified by the leading asterisk on the line, although the search is keyed to line number 3. The symbol table is also initialized at empty.

Each line of the assembled program will have the hexadecimal memory address, the hexadecimal object code to be loaded into that address, a semicolon marking the end of the machine code, then the next source line. Notice that the source line is echoed as it is read (this is done by the I/O routines), so the assembled code for that line is at the beginning of the next line. If the source file contains a linefeed character after each carriage return, then the object code will appear on the same line in the listing, but in fact the object code follows it in the output file. In the case of the *LN*, *PC*, and *BC* opcodes, which generate more than two bytes of code, a second line will be used for the excess object code. The listing produced for Pass 1 will look very much like that for Pass 2, except that some of the object code will be incomplete.

Assembly errors which do not crash the program will be identified by a two letter indication enclosed in a pair of asterisks. The following is a summary of the errors recognized and flagged by this assembler:

        *DL* Duplicate label (Pass 1 only)
        *IE* Unidentifiable mnemonic
        *OP* Incorrectly formed Operand
        *US* Undefined symbol in jump or branch
        *LE* Premature line end

Some source program errors will be trapped by the **TINY BASIC** interpreter and halt the assembler. These are catastrophic in the sense that not only is the assembly aborted, but the remainder of the source file is loaded by TINY into memory over the assembler as if it were a BASIC program, thus destroying the integrity of the assembler. Errors which are catastrophic are:

        Lines without a line number
        Excessively long lines
        Invalid expression as the operand of LN or LB
        Symbol table overflow

This version of the assembler may be expected to run in something under 8K bytes of memory, depending on how many of the comment lines and excess blanks are removed.

Operationally, the program is fairly direct with few tricky kludges.

The symbol table is built by the assembler by stealing space out of the GOSUB stack. For each label to be added to the table, three unRETURNed GOSUBs are executed, making six bytes available. Symbols with less than 4 characters are filled out with spaces. The same symbol table search routine is used for both definition (to check for duplicates) and reference. The table is searched with the memory fetch USR commands.

The opcode table is searched in a similar way. The hex codes are never actually converted to binary, but a special subroutine selects the appropriate digit printing statement based on the ASCII value of the codes. In the few cases where the operand is imbedded into the opcode, the extra bits are added in before output.

The type of instruction (i.e. the kind of operands accepted for the particular instruction) is determined by its position in the table: The first position is SX; the next two are jumps; the next five are branches, followed by the string opcodes (note the overlap). The literal byte and number opcodes are finally followed by all the generics (no operand). The assembler knows how many opcodes there are, and stops looking when this count is reached, rather than looking for some end-of-table flag. The table is broken up into several lines of **TINY BASIC**; the line boundaries are aligned with the mnemonic positions in the table, so they represent opcodes which never match (the mnemonic would be CR-NUL).

The operation of the remainder of the assembler is fairly self-evident and needs no further discussion.



```
1 REM **TINY BASIC** IL ASSEMBLER VERSION 0        1 JAN 1977
2 GOTO 100
3 *SX00JS30J 38BR40BVA0BNC0BEE0BC80PC24LB09LN0ANO08
4 *DS0BSP0CSB10RB11FV12SV13GS14RS15GO16NE17AD18SU19MP1ADV1B
5 *CP1CNX1DLS1FPN20PQ21PT22NL23GL27IL2AMT2BXQ2CWS2DUS2ERT2F
6
7               ....COPYRIGHT (C) 1977  BY TOM PITTMAN....
10 REMARKS:
11 LINES 3-5 ARE OPCODE TABLE
12 LABEL TABLE USES GOSUB STACK
13 .
14 THIS PROGRAM USES A 2-BYTE PEEK USR FUNCTION
15 PUT ITS ADDRESS IN VARIABLE D.
16 IN 6800:
17   LDA A,1,X        A IS LSB
18   LDA B,0,X
19   RTS
20 IN 6502:
21   STX $C3          ($C2=00)
22   LDA ($C2),Y      GET MSB
23   PHA              SAVE IT
24   INY
25   LDA ($C2),Y      GET LSB
26   TAX
27   PLA
28   TAY              Y=MSB
29   TXA
30   RTS
31 NOTE THAT THIS PROGRAM CORRECTS FOR 2-BYTE-DATA
32 IN 6502 FORMAT (LSB,MSB) WHEN INITIALIZING.
33 .
34 THE FOLLOWING VARIABLES ARE DEFINED:
35 A  STARTING ADDRESS
36 B  LINE BUFFER POINTER ADDRESS
37 C  LINE POINTER WORK
38 D  2-BYTE PEEK USR FUNCTION ADDRESS
39 E  END OF OPCODE TABLE
40 F  PASS #
41 G  PEEK USR FUNCTION ADDRESS
42 H  HEX WORK
43 I  TEMP WORK
44 J  TEMP WORK
45 K  TEMP WORK (HEX)
46 L  (RELATIVE) LOCATION COUNTER
47 M
48 N  LINE NUMBER
49 O  OP TABLE START
50 P  POKE USR FUNCTION ADDRESS
51 Q
52 R
53 S  SYMBOL TABLE START
54 T  TEMP (TABLE POINTER)
55 U
56 V  SYMBOL WORK
57 W  SYMBOL WORK
58 X ERROR COUNT
59 Y
60 Z
61 .
62 SOURCE FILE IS IN THE FORM
63 (LINE NUMBER) :LABEL OP OPND COMMENTS
64 THE LINE NUMBER MUST BE >0.
65 THE LABEL IS IDENTIFIED BY THE LEADING COLON,
66 AND MAY BE 1-4 CHARACTERS LONG (FIRST IS LETTER);
67 IT IS TERMINATED BY BLANK, AND MAY BE OMITTED.
68 .
69 OP IS THE 2-LETTER OPCODE.
70 OPND IS THE OPERAND:
71 FOR SX IT MUST BE A DIGIT 1-7
72 FOR LB OR LN, A DECIMAL NUMBER 0-255 OR 0-65535
73 FOR PC, A STRING OF THE FORM 'STRING'
74 FOR JUMPS & BRANCHES IT MUST BE A SYMBOL
75 BRANCHES MAY REFER TO SYMBOL "*"
76 TO INVOKE ERROR STOP FORM.
77 BC REQUIRES BOTH A SYMBOL AND A STRING,
78 SEPARATED BY ONE OR MORE SPACES.
79 COMMENTS SHOULD BE PRECEDED BY A SPACE,
80 AND SHOULD NOT BEGIN WITH A DIGIT OR (+,-,*,/)
81 COMMENT LINES HAVE A PERIOD
82 FOLLOWING THE LINE NUMBER.
83 THE END OF FILE IS A LINE NUMBER 0.
84 .
85 SOURCE IS LISTED ON BOTH PASSES.
86 OUTPUT IS: HEX ADDRESS, HEX CODE, SEMICOLON,
87 ON SAME LINE AS FOLLOWING SOURCE.
88 .
89 .
90 ERROR FLAGS:
91 *DL* DUPLICATE LABEL (PASS 1)
92 *OP* OPERAND FORMAT ERROR
93 *IE* UNDEFINED OP CODE
94 *LE* INCOMPLETE LINE
95 *US* UNDEFINED SYMBOL (PASS 2)
99 .
100 REM
101 REM LINES 101-199 ONLY NEED TO EXECUTE ONCE.
102 REM THEY SHOULD BE DELETED AT STOP.
103 REM INPUT ADDRESS CONSTANTS
104 PRINT "PLEASE TYPE IN USR ADDRESS FOR PEEK (IN DECIMAL)";
105 INPUT G
106 PRINT "ADDRESS FOR POKE";
107 INPUT P
108 PRINT "ADDRESS FOR 2-BYTE PEEK";
109 INPUT D
110 B=47
111 O=USR(D,32)
112 E=USR(D,34)
113 IF USR(G,B)>0 GOTO 118
114 B=46
115 O=USR(G,32)+USR(G,33)*256
116 E=USR(G,34)+USR(G,35)*256
118 E=E+1
119 REM FIND OPCODE TABLE (LINE 3)
120 O=O+1
121 IF USR(G,O)<>3 GOTO 120
122 O=O+2
130 Y=1
131 N=0
132 PRINT "DO YOU NEED INSTRUCTIONS (Y OR N)";
133 INPUT I
134 IF I=Y LIST 61,99
190 PRINT "REMOVE LINES 10-99, 101-199"
191 PRINT "OR IF YOU HAVE PLENTY OF MEMORY,"
192 PRINT "RETYPE LINE: 100 GOTO 200"
193 PRINT "THEN TYPE RUN."
198 END
199 REM 2-PASS ASSEMBLER. START FIRST PASS.
200 X=0
201 S=E
202 F=0
203 PRINT "(DECIMAL) STARTING ADDRESS";
204 INPUT A
205 F=F+1
206 IF F=3 GOTO 760
207 L=0
208 PRINT
209 PRINT "TBIL ASSEMBLER, PASS ";F
210 PRINT
211 GOSUB 460
212 PRINT ";    ";
213 REM GET NEXT INPUT LINE
214 I=USR(P,USR(G,B),13)
215 INPUT N
216 REM LINE NUMBER 0 IS EOF
217 IF N=0 GOTO 205
218 GOSUB 460
219 REM CHECK FOR COMMENT
220 I=USR(G,USR(G,B))
221 IF I<58 GOTO 212
222 REM PROCESS LABEL, IF ANY
223 IF I>64 GOTO 300
224 GOSUB 405
225 GOSUB 500
231 REM CHECK FOR DUPLICATES ON PASS 1
232 IF F>1 GOTO 300
234 IF T=0 GOSUB 237
235 GOTO 901
237 GOSUB 238
238 GOSUB 239
239 S=S-6
240 REM INSERT THIS ONE
241 I=USR(P,S,V/256)+USR(P,S+1,V)
242 I=USR(P,S+2,W/256)+USR(P,S+3,W)
243 I=USR(P,S+4,L/256)+USR(P,S+5,L)
290 REM LOOK AT OPCODE
300 GOSUB 410
301 IF I<65 GOTO 911
305 I=USR(D,USR(G,B))
306 GOSUB 404
307 REM SEARCH OPCODE TABLE
308 T=O
309 IF USR(D,T)=I GOTO 313
310 T=T+4
311 IF T<O+167 GOTO 309
312 GOTO 911
313 V=USR(G,T+2)
314 W=USR(G,T+3)
315 L=L+1
316 IF T=O GOTO 330
317 IF T<O+10 GOTO 340
318 IF T<O+30 GOTO 360
319 IF T=O+32 GOTO 380
320 IF T=O+36 GOTO 350
321 IF T=O+40 GOTO 550
322 REM THESE OPCODES HAVE NO OPERAND
323 H=V
324 GOSUB 434
325 H=W
326 GOSUB 434
327 PRINT ";  ";
328 GOTO 214
329 REM STACK EXCHANGE OPERATOR
330 GOSUB 410
331 W=USR(G,USR(G,B))
332 IF I>48 IF I<56 GOTO 323
333 REM OPERAND FORMAT ERROR
334 GOTO 921
336 IF F=1 GOTO 212
337 GOTO 931
339 REM JUMP & CALL
340 L=L+1
341 GOSUB 410
342 IF I<65 GOTO 334
344 K=W-W/16*16
345 GOSUB 500
346 IF T=0 GOTO 336
347 K=I+(K+48)*256
348 GOTO 356
349 REM PUSH LITERAL BYTE ON STACK
350 L=L+1
351 GOSUB 410
352 IF I<48 GOTO 334
353 IF I>57 GOTO 334
354 INPUT K
355 K=K+2304
356 GOSUB 440
357 PRINT ";";
358 GOTO 214
359 REM RELATIVE BRANCHES
360 K=T
362 GOSUB 410
363 IF I=42 GOTO 365
364 IF I<65 GOTO 334
365 GOSUB 500
366 IF T=0 IF K<O+28*F GOTO 336
367 IF I>L+31 GOTO 334
368 IF K=O+12 THEN I=I+32
369 IF I<L GOTO 334
370 I=I-L
371 T=K
372 H=USR(G,K+2)+I/16
373 K=I-I/16*16
374 GOSUB 434
375 GOSUB 455
376 IF T<O+28 GOTO 327
377 GOTO 381
379 REM STRING OPERATORS
380 PRINT "24";
381 GOSUB 410
382 J=L
383 T=I
384 GOSUB 405
385 K=USR(G,USR(G,B))
386 GOSUB 405
387 I=USR (G,USR(G,B))
388 IF I<>94 GOTO 391
389 K=K-64
390 GOTO 386
391 L=L+1
392 IF I=13 GOTO 334
393 IF T=I GOTO 397
394 GOSUB 450
395 K=I
396 GOTO 386
397 K=K+128
398 GOSUB 450
399 PRINT ";";
400 IF L=J+1 GOTO 214
401 GOTO 210
402 REM      ---      SUBROUTINES
403 REM ADVANCE INPUT LINE POINTER
404 GOSUB 405
405 C=USR(P,B,USR(G,B)+1)
406 RETURN
407 REM
408 REM SKIP BLANKS IN INPUT LINE
409 GOSUB 405
410 I=USR(G,USR(G,B))
411 IF I=32 GOTO 409
412 IF I>32 RETURN
413 GOTO 941
418 REM
419 REM PRINT HEX DIGITS
420 PRINT "A";
421 RETURN
422 PRINT "B";
423 RETURN
424 PRINT "C";
425 RETURN
426 PRINT "D";
427 RETURN
428 PRINT "E";
429 RETURN
430 PRINT "F";
431 RETURN
434 IF H>64 GOTO H+H+290
435 H=H-48
436 IF H>9 GOTO 400+H+H
437 PRINT H;
438 RETURN
439 REM PRINT NUMBER AS HEX
440 H=K/4096
441 IF K<0 THEN H=H-1
442 K=K-H*4096
443 IF H<0 THEN H=H+16
444 GOSUB 436
445 H=K/256
446 K=K-H*256
447 GOSUB 436
450 H=K/16
451 K=K-H*16
452 GOSUB 436
455 H=K
456 GOTO 436
458 REM
459 REM PRINT LOCATION COUNTER
460 K=A+L
461 GOSUB 440
462 PRINT " ";
463 RETURN
498 REM
499 REM LOOK UP SYMBOL IN TABLE
500 V=0
501 W=8224
502 C=USR(G,B)
503 I=USR(G,C)
504 IF I<48 GOTO 525
505 I=USR(G,C+1)
506 IF I<32 THEN I=(USR(P,C+1,32)+USR(P,C+2,13))*0+32
508 W=USR(D,C)
509 GOSUB 404
510 IF V>0 GOTO 513
511 V=W
512 GOTO 501
513 T=S
514 GOTO 518
515 I=USR(D,T+4)
516 IF V=USR(D,T) IF W=USR(D,T+2) RETURN
517 T=T+6
518 IF T<E GOTO 515
519 T=0
520 I=L
521 RETURN
524 REM ASTERISK OPERAND?
525 IF I<>42 GOTO 510
526 T=1
527 I=L
528 GOTO 405
548 REM
549 REM PUSH 2-BYTE LITERAL ONTO STACK
550 PRINT "0A;"
552 GOSUB 460
553 L=L+2
554 GOSUB 410
555 IF I<48 GOTO 334
556 IF I>57 GOTO 334
557 INPUT K
558 GOTO 356
700 REM PROGRAM END
760 PRINT
770 PRINT X;" ERRORS"
790 END
900 REM ERROR MESSAGES
901 PRINT "*DL* ";
902 X=X+1
903 GOTO 300
911 PRINT "*IE* ";
912 X=X+1
914 L=L+2
915 GOTO 214
921 PRINT "*OP* ";
922 X=X+1
923 GOTO 214
931 PRINT "*US* ";
932 X=X+1
933 GOTO 214
941 PRINT "*LE* ";
942 X=X+1
944 RETURN
999 END
```
