package crypto

import (
	"crypto/aes"
	"crypto/cipher"
)

type Context struct {
	Keys
}

func NewContext(keys Keys) *Context {
	return &Context{
		Keys: keys,
	}
}

func (c *Context) doDecrypt(ciphertext []byte, block cipher.Block) ([]byte, error) {
	// ECB works on full blocks only, so we need to trim the last block
	trimsize := len(ciphertext) % block.BlockSize()
	plaintext := make([]byte, len(ciphertext)-trimsize)
	copy(plaintext, ciphertext[:len(ciphertext)-trimsize])

	mode := NewECBDecrypter(block)
	mode.CryptBlocks(plaintext, plaintext)

	// Tack the last block back on
	plaintext = append(plaintext, ciphertext[len(ciphertext)-trimsize:]...)
	return plaintext, nil
}

func (c *Context) DecryptAES(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return nil, err
	}

	return c.doDecrypt(ciphertext, block)
}

func (c *Context) DecryptNG(ciphertext []byte, filename string, filesize uint32) ([]byte, error) {
	ngKey := c.NgKeyForFile(filename, filesize)
	block := NewNGCipher(ngKey, c.ngDecryptTable)

	return c.doDecrypt(ciphertext, block)
}
