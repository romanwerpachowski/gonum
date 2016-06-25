// Generated by running
//  go generate github.com/gonum/internal/asm
// DO NOT EDIT.

// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Some of the loop unrolling code is copied from:
// http://golang.org/src/math/big/arith_amd64.s
// which is distributed under these terms:
//
// Copyright (c) 2012 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

//+build !noasm,!appengine

#include "textflag.h"

// Hardware NOP, used to avoid splitting instructions across cache lines
#define NOP7 BYTE $0x0F; BYTE $0x1F; BYTE $0x80; BYTE $0x00; BYTE $0x00; BYTE $0x00; BYTE $0x00

// func AxpyUnitary(alpha float64, x, y []float64)
TEXT ·AxpyUnitary(SB), NOSPLIT, $0
	MOVQ    x_base+8(FP), SI  // SI := &x
	MOVQ    y_base+32(FP), DI // DI := &x
	MOVQ    x_len+16(FP), CX  // CX = min( len(x), len(y) )
	CMPQ    y_len+40(FP), CX
	CMOVQLE y_len+40(FP), CX
	CMPQ    CX, $0            // if CX == 0 { return }
	JE      end
	XORQ    AX, AX
	MOVSD   alpha+0(FP), X0   // X0 := { alpha, alpha }
	SHUFPD  $0, X0, X0
	MOVUPS  X0, X1            // X1 := X0   for pipelining
	MOVQ    DI, BX            // BX = DI % 16
	ANDQ    $15, BX
	JZ      no_trim

	// Align on 16-bit boundary
	MOVSD (SI), X2 // X2 := x[0]
	MULSD X0, X2   // X2 *= a
	ADDSD (DX), X2 // X2 += y[0]
	MOVSD X2, (DI) // y[0] = X2
	INCQ  AX       // i++
	DECQ  CX       // CX--
	JZ    end      // if CX == 0 { return }

no_trim:
	NOP7             // j
	MOVQ CX, BX
	ANDQ $7, BX      // BX := CX % 8
	SHRQ $3, CX      // CX = floor( CX / 8 )
	JZ   tail2_start // if CX == 0 { goto tail_start }

loop:  // do {
	// y[i] += alpha * x[i] unrolled 2x.
	MOVUPS (SI)(AX*8), X2   // X_i = x[i]
	MOVUPS 16(SI)(AX*8), X3
	MOVUPS 32(SI)(AX*8), X4
	MOVUPS 48(SI)(AX*8), X5

	MULPD X0, X2 // X_i *= a
	MULPD X1, X3
	MULPD X0, X4
	MULPD X1, X5

	ADDPD (DI)(AX*8), X2   // X_i += y[i]
	ADDPD 16(DI)(AX*8), X3
	ADDPD 32(DI)(AX*8), X4
	ADDPD 48(DI)(AX*8), X5

	MOVUPS X2, (DI)(AX*8)   // y[i] = X_i
	MOVUPS X3, 16(DI)(AX*8)
	MOVUPS X4, 32(DI)(AX*8)
	MOVUPS X5, 48(DI)(AX*8)

	ADDQ $8, AX // i += 2
	LOOP loop   // } while --CX > 0
	CMPQ BX, $0 // if BX == 0 { return }
	JE   end

tail2_start: // Reset loop registers
	MOVQ BX, CX // Loop counter: CX = BX
	SHRQ $1, CX // CX = floor( BX / 2 )
	JZ   tail

tail2:  // do {
	MOVUPS (SI)(AX*8), X2 // X2 = x[i]
	MULPD  X0, X2         // X2 *= a
	ADDPD  (DI)(AX*8), X2 // X2 += y[i]
	MOVUPS X2, (DI)(AX*8) // y[i] = X2
	ADDQ   $2, AX         // i += 2
	LOOP   tail2          // } while --CX > 0

tail:
	ANDQ $1, BX // BX = BX % 2
	JZ   end    // if BX % 2 == 0 { return }

	MOVSD (SI)(AX*8), X2 // X2 = x[i]
	MULSD X0, X2         // X2 *= a
	ADDSD (DI)(AX*8), X2 // X2 += y[i]
	MOVSD X2, (DI)(AX*8) // y[i] = X2

end:
	RET
