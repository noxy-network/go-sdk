//go:build cgo

package kyber

/*
#cgo CFLAGS: -I${SRCDIR}/pqclean/ml-kem-768 -I${SRCDIR}/pqclean/common
#cgo LDFLAGS: -lm

#include <stdint.h>
int mlkem768_encapsulate(uint8_t *ct, uint8_t *ss, const uint8_t *pk);
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	pkSize = 1184
	ctSize = 1088
	ssSize = 32
)

type pqcleanEncapsulator struct{}

func newEncapsulator() encapsulator {
	return &pqcleanEncapsulator{}
}

func (p *pqcleanEncapsulator) Encapsulate(publicKey []byte) (ciphertext []byte, sharedSecret []byte, err error) {
	if len(publicKey) != pkSize {
		return nil, nil, fmt.Errorf("invalid Kyber public key: expected %d bytes, got %d", pkSize, len(publicKey))
	}

	ct := make([]byte, ctSize)
	ss := make([]byte, ssSize)

	ret := C.mlkem768_encapsulate(
		(*C.uint8_t)(unsafe.Pointer(&ct[0])),
		(*C.uint8_t)(unsafe.Pointer(&ss[0])),
		(*C.uint8_t)(unsafe.Pointer(&publicKey[0])),
	)
	if ret != 0 {
		return nil, nil, fmt.Errorf("Kyber encapsulate failed: %d", ret)
	}
	return ct, ss, nil
}
