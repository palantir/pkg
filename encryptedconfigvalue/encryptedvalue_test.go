// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/encryptedconfigvalue"
)

const (
	testPlaintext = "plaintext"

	testAESEncryptedValKey = "AES:LICx0yKzQm5a6IE13aJ3xOsRv+8AujqHocTFI4yk4Jw="
	testAESEncryptedVal    = "enc:eyJ0eXBlIjoiQUVTIiwibW9kZSI6IkdDTSIsImNpcGhlcnRleHQiOiJNOTRrSXlvYTUrMloiLCJpdiI6InVBR3FSbFA5d2l6cGRCMHoiLCJ0YWciOiJBQ1N1ekR3VFVMb21zanhwRk1rWUtBPT0ifQ=="

	testRSAEncryptedValPrivKey = "RSA-PRIV:MIIEvAIBADALBgkqhkiG9w0BAQEEggSoMIIEpAIBAAKCAQEAtKj+bpwUCq22ABjeJLBje+mD5XWmUAc8K2NbEGNGFaWGVAE1h/2Pjgxmj+LR4Bgt3OleYOnfV99ToqMNgB+HnNOJCg5LkHfq+WD6tRwhxFQMCmt73k9i8fgg7OCb1yTWo6pLCBIVWeisO0j0b1CYeIHebRemkx+8AK0ebsv4tdrIwAlb4jJTSz2rKZpEw7rLcGr8dFOYP5pg/jLJneittODD/uJj+1lpOze/AUT3bcuF6Ku0Oh4zNIvPcmm72bbr7+61lFOJB1IbDg1ahklE9m439/OOi3OOTdqq/HOu0k/dThrvovV1eedoL6UQz6RdijHNUt3iZqiues/Mq5dLSwIDAQABAoIBAQCEQyTi/cl+d+bC83HPEoQC99bkatmzxVg7u6WzvbpVprVNUwVJ5kzvBg0gUkKs+Ya6MPAzq4Uj5BBrBUyg/HRgUE4H2qdfwSt6H5HsfggKoC2gg0hQXXZnB+2y/k2ZmRK7B7We1v5isIFHdgXeaPb3YrzgyWveUmFlbVjWbOZM3AAJ0FczP2b3DErFS/iMyzdjCY9xwwXhQediMASj24c44/VLsaRCFesPXHoAXCvLLlPmNhfaw6ZVtHblg0QlFNftOUlIXC+s9yIN2ec38C10VR/yfGqVSYz+owXqNKRpMfsqNe1jWnl3+BVaqO53vsXzkYU8n8/vHdRSRZOiKpwhAoGBAMVutRUefOcApu5iEpHK+7Jte0o1kNFIwCXqiujjZcU/DKjDj2yK90ioza7Ntp7EHI9MUgCknyyiMlI/1VtHl3KiNfi0FQ646/AxOgzfrmUZTTyUgq02ToxFnAr1XYBzwwAPHKM4p2nJrf+G/7FpXhCMhK4qwGfMJ4C+i0pCoiJ/AoGBAOpAktM4SZGmBdtPyRpp0Z8tkrHoRNwn1YK+VS7XfkKmeMrsPEev7cjesaNJnMBjtlpGrAzVJC/ycEz5lZW7gBA5i/hDOLGegLjuu1SOTKXU4IFw5+vjTe4ecMFLLRE/rTeWMR3RfslzTiV66zLKZ9zuhq4YccGi3IFKKXVp+Fk1AoGAN86kVxToH2/6v7VvJFDpNrVlvUNI7S+QSOd0XoIwuUGqNWYZ+4eIgLxeb4PslBJBNGxRXacq6zXp3X/3sjaZY6jgcq2Mqj2xS5LOoubzZ9ZwE6izC30nVNU0V5Cl3nJac4DSCn0wLWH50hn52s867JibxJOHEZAOtoCl5NbS98cCgYAWoCoOUK96a+jA6BHqhTIEB+jVWjPcd9R9jli374R4d4/POcYQvoNfFXNe7CtBwd/JFG5lxuh54RbLuIekMLoL1yMX1ZZSQZb5RcW+QwhQNCGDHx6ngAr05ufJI7O0qMvYRJ9129g9KO/xWtAA1d/2TOuhQScrpslZi4o5lwSvyQKBgQCw/nLpPPlBGeA6jA0yZOuMPDZMGStLOAsGMmhV6LnBBllE475qQRPD/1xgcoWU7+u9H6sJNBR5p/WJq58IZFHzVCFVEBijLbNXDKOF9nDaczzXID5pM2Pspoz7JPpZkIFk0D2IR73M2RfoWNxYPRJCImDaL7HOXND6SNA+p6kkMg=="
	testRSAEncryptedVal        = "enc:eyJ0eXBlIjoiUlNBIiwibW9kZSI6Ik9BRVAiLCJjaXBoZXJ0ZXh0IjoiRXdHRENVcXpvQzNReVF4T3V4cE5sc2FzWldOQm5FN0d2bVdFWXNKUm1sbHM4R0s4MzFyQ0M2SGZGeXhmWmhzU3FqSVpKZnUxOU50ZERXTlJZVnlWK3p6OE9Ndk5mTjZVeFpxRTZwdFF2R1RHbUdBQk9CK2tFZGhNeFZFR0J3TW9YSnh0Zlg5Smk0ZHhLdHpZZlhEbGlYdXU4OTVwejlQN2l2Nm9GR2Q1U21PTlQwWXNra2piUlkwaFRYRWV6dGlvRW0wMHNXak8ySnJndXUwNUZ5SkhmRkpINStPcDRsZFZkRnA1ZWpWK2xROEVMNUVBWjJoNytscVY5Rmp0emxBdWtxbUp4OEVwMGhJbGhPUjIzNndJcWQwZHpBVzZnODYybENLb1Nob0dvS04ydHR5alU2TURRUWo0a0hBNFdCbTBOdm9XREZHN1p3T1ozbDIwVnY3R1ZRPT0iLCJvYWVwLWFsZyI6IlNIQS0yNTYiLCJtZGYxLWFsZyI6IlNIQS0yNTYifQ=="

	javaPlaintext             = "my secret. I don't want anyone to know this"
	javaAESKey                = "AES:rqrvWpLld+wKLOyxJYxQVg=="
	javaLegacyAESEncryptedVal = "enc:QjR4AHIYoIzvjEHf53XETM3QYnCl1mgFYC51Q7x4ebwM+h3PHVqSt/1un/+KvpJ2mZfMH0tifu+htRVxEPyXmt88lyKB83NpesNJEoLFLL+wBWCkppaLRuc/1w=="

	javaRSALegacyPrivKey      = "RSA:MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDN+p4EwZeMdOqs5z4TDWcRaMt03EEEq0deHlaYUO2HpmtR5DzfZrTuTbMMLueGgV8hwj5cnXhZ+n3eMG86jZ2LGtYWmbyo4NwTPjTtjz27v8vkLzuwYHsWxq6jbCp66leOhjYFYefpn0mA224S4VpBHSNOTEl3z6Wg5FAaIF9T7VRnT/xZYt2KFNWelgZBngzWSE6B1g9nWIorxWHrCygCpKOTGgfKoVYqQhT+pKukstxIV5kE/UXff9GQ8zSLoCejEoVqQe8nwbrLKmihP1kfjfxh9qtEcBB/4Bs+GQW41kkC7DaVL//5cYcVj0T5gJzPSxYnokmDyl++vkBr2YZtAgMBAAECggEAEer8NgO1MDW3eGUBRF0FG0GXeUnzqflQUwKmm8dmckdqzIvjM7fWg2hk6+lkoJG+ecxQ6nOUVZdxvZNPCbPqAYDLINoszDALVO0zY3rzbtKnZOkq8xPhgUC1TmgJZfnetfo81skGiI8fsMLl12SdGk7zlEsUlQSOLunNgghQ4pb5dpMfhyp0Q4ThmlfCBhY/XsRm9KLF98Il94QO9orYCJnVjOos/lWd6UKuLWEOf3CL/ucIaUAkUmu8PMO/AHX9xW6vNIr76rvdasocUjv3KpFtV5gQX3IhKhehuQlW758a/EeNL725QhjfesF7tKPtsSPWzQ8dyFjHWF6xn214jQKBgQD+1Zj7yHF57/nfHEvXRyhkbaDkqU/uGNFUTGg0TvucAS9sas8CjJ7WHBrUfjfWWxrCNqAY2sfpxlUd+0di3aWrUwyM1h91dYAhYk5NHnzkjhSi4wcbwHjN+BRPRMjgp+BsF/ZySpZK/tHUbCUgyQHWtJvkpHdiHcTDZh5wII9/4wKBgQDO68/I6qmoTpHRxF9zpOePTnwjWBwJ7r3qlTnFoQJGNEusDglI2GiaD3lRSxF1TfKnivUYbEhrHbMXbfn2lOwPlHtpjESVAseWU6Qmz6r/TITk4M8kIEzo+yomM6QeBJwd4JAgjot456sT5X+Vv0NCtfweB0ex2geZsK4X9MERbwKBgGlVPsvr+UOutrjLCGouhnqkedmqRlijN3tBrdzZPNUqBEErEO/70fesXEay+T+IHtJiI+DCJdnyWeJvp/0sorrjNA/OvegeLl0eNkFYNcV/GPaPIrQM5aI1RafSRbneijwD16E8RU0wcOj93objrvfhZYKnnJUYuukNf81XGBmDAoGAXDJj/eDZQV3oyS+XXD7A0nClDVaH/8D5rBlbiXxJOCC7CumiJ2wNh3+XjapGGB9oHFDlDkHJLrkoACuHceA/Il4Fcy0FreN0LL4N6SEkzuY4XIbypOUjf7fRuv3NhXaGXSWe8nKxIGkRKCdc5ss22/WcZYDW6B7+vfMkTxZGJE8CgYEAv67Q70wtHRsl/3tnVUTgzBeB9HipilEinkkCUkDqYEf3pH6dhlmtkPi9YHvV38VH7AT6zqiI86mlPE7iQKEkBrYajrGEQ0UrqkjebVyN3wTwtKBXfhDkg4f2E58tcQrsaiGfMYG2/F8/BIRhPpqFUQzq03mgmFZtAqyhXl62o2w="
	javaLegacyRSAEncryptedVal = "enc:GNOe/P/KQ8fvuhhBVNMZQ2jDu+cdv7im1N4GamZ64u9LhvoiLP6RiSFnHFRcbIupEIxJQ1IM/9cJ0DpUsxPpObH+vV0fCZZ/Aqrb08s46hodTPDLU76JNrtaxlCssXYxFN/Ni8k95pKauwPxRfvTP0SUf7o9rsZrY6LdV9+M3y6mNrEIKevAZQZtNmvXriclQGV1CwRzV/0sNVuTfNqNw0lDsI4hcvC26DhLrXla8jCUiKEYDFAqVr2DaTwtV3htxtCB36Jk6Lg5abdcc9B/ZqV7lfUIddGEuXFzhz8KIIGtwVVXqis15Dw1ECSNJhicHZp43vSYN9y9NJTnvTAhCQ=="
)

func TestDecryptEncryptedValue(t *testing.T) {
	for i, currCase := range []struct {
		name          string
		plaintext     string
		decryptionKey string
		encryptedVal  string
	}{
		{
			name:          "AES",
			plaintext:     testPlaintext,
			decryptionKey: testAESEncryptedValKey,
			encryptedVal:  testAESEncryptedVal,
		},
		{
			name:          "AES legacy",
			plaintext:     javaPlaintext,
			decryptionKey: javaAESKey,
			encryptedVal:  javaLegacyAESEncryptedVal,
		},
		{
			name:          "RSA",
			plaintext:     testPlaintext,
			decryptionKey: testRSAEncryptedValPrivKey,
			encryptedVal:  testRSAEncryptedVal,
		},
		{
			name:          "RSA legacy",
			plaintext:     javaPlaintext,
			decryptionKey: javaRSALegacyPrivKey,
			encryptedVal:  javaLegacyRSAEncryptedVal,
		},
	} {
		decKey, err := encryptedconfigvalue.NewKeyWithType(currCase.decryptionKey)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		ev, err := encryptedconfigvalue.NewEncryptedValue(currCase.encryptedVal)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		decrypted, err := ev.Decrypt(decKey)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		assert.Equal(t, currCase.plaintext, string(decrypted), "Case %d: %s", i, currCase.name)
	}
}
