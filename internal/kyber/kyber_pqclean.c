// CGO wrapper for PQClean ML-KEM 768 (Kyber768).
// Uses the same implementation as the Node.js WASM for interoperability.
#include "api.h"
#include <stdint.h>
#include <string.h>

// mlkem768_encapsulate encapsulates a shared secret for the given public key.
// Returns 0 on success. ct must be 1088 bytes, ss must be 32 bytes, pk must be 1184 bytes.
int mlkem768_encapsulate(uint8_t *ct, uint8_t *ss, const uint8_t *pk) {
    return PQCLEAN_MLKEM768_CLEAN_crypto_kem_enc(ct, ss, pk);
}
