#include "textflag.h"

TEXT ·hostcall_vi32_i32(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_vu8_i32(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall__u32x2(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall__f64(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_j_i32(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_j_u32x2(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_jvi32_(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_jvu32_(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_ju32j_(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_ju32_(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_jvf32_(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_jf32_(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_jx2_(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_jx2vu32_(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_jx2vf32_(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_jx2vi32_(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_ju32x3_(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·hostcall_u32_(SB), NOSPLIT, $0
  CallImport
  RET
