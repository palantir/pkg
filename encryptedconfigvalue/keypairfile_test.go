// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/nmiyake/pkg/dirs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/encryptedconfigvalue"
)

func TestLoadKeyPairFromDefaultPath_AES(t *testing.T) {
	origWd, err := os.Getwd()
	defer func() {
		require.NoError(t, os.Chdir(origWd))
	}()
	require.NoError(t, err)

	tmpDir, cleanup, err := dirs.TempDir("", "")
	defer cleanup()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	err = os.MkdirAll("var/conf", 0755)
	require.NoError(t, err)

	const aesKey = "AES:s0u/5zMvOz9bd2/7QSJ0yaRpav9kgAmLh6GyXkttwC4="
	err = ioutil.WriteFile("var/conf/encrypted-config-value.key", []byte(aesKey), 0644)
	require.NoError(t, err)

	kp, err := encryptedconfigvalue.LoadKeyPairFromDefaultPath()
	require.NoError(t, err)

	assert.Equal(t, encryptedconfigvalue.AES, kp.Algorithm())
	pubSKWA, err := kp.PublicKey().ToSerializable()
	require.NoError(t, err)
	assert.Equal(t, aesKey, pubSKWA.SerializedStringForm())
	assert.Nil(t, kp.PrivateKey())
}

func TestLoadKeyPairFromDefaultPath_RSA(t *testing.T) {
	origWd, err := os.Getwd()
	defer func() {
		require.NoError(t, os.Chdir(origWd))
	}()
	require.NoError(t, err)

	tmpDir, cleanup, err := dirs.TempDir("", "")
	defer cleanup()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	err = os.MkdirAll("var/conf", 0755)
	require.NoError(t, err)

	const (
		rsaPubKey  = "RSA:LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMWx0ekVycGl5aEVSc3VxY28wNnEKemxWZVhIUExQYzJSV1JiTmE0aWczYUFkRU5KUjRJUjdGY3ptUUFReHFvMlAydlFXcEo5eGdmaVBaV000L3c5cwp1MlBMTDgySEN2QWxOdWdNU0s4OUZkTGd3MTVHWnpkOWp6dHU1QkVuZ0VSNlVVNHFBKzFRZUxQbGRwUHlmemNTClRsaStWVTZYdjlMSTE5RU9naEJUSXN5azcweEcrc25hMjRLcmZEdHRBUE9BaTlrYlRqV3ZDYTV1M3lrbFdQZFAKbHhSMElIZUlHTlJKWGMwNmZMMXFQMjVJRjFNR2hkN2lKWE1reEEreTF3T0RhellzU1NaTEY5dXB6TytBdi9DVwowOTFtNGhMaG9ocUdobHBRZ2lqVWgrdWc5aXF2Z2ZSVm5aSXdlSTU0aU1zRWlkMDY4eHVEZ084bXptaHBrWksxCnRRSURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K"
		rsaPrivKey = "RSA:MIIEvQIBADALBgkqhkiG9w0BAQEEggSpMIIEpQIBAAKCAQEA1ltzErpiyhERsuqco06qzlVeXHPLPc2RWRbNa4ig3aAdENJR4IR7FczmQAQxqo2P2vQWpJ9xgfiPZWM4/w9su2PLL82HCvAlNugMSK89FdLgw15GZzd9jztu5BEngER6UU4qA+1QeLPldpPyfzcSTli+VU6Xv9LI19EOghBTIsyk70xG+sna24KrfDttAPOAi9kbTjWvCa5u3yklWPdPlxR0IHeIGNRJXc06fL1qP25IF1MGhd7iJXMkxA+y1wODazYsSSZLF9upzO+Av/CW091m4hLhohqGhlpQgijUh+ug9iqvgfRVnZIweI54iMsEid068xuDgO8mzmhpkZK1tQIDAQABAoIBAQCDTWEfh6wbunjs72kjX3yhBwnV99f284Sk3aLWy8o992XWd/5PWNdMc0ZW0DrcDfqgVAPKsyAETQ0JPc4b7obcAjTkAzFFMfSZvWpI246/X3zuL0FQ2FzA79btPNTFbSy/wPFblnJEfW2BRP61jjZYZ2OvPYUWqzb7e8M3SGikVzdgnHHr+5udJ9ywRfa6ZFjqztY7Ruz7930xO32W22HMltJQ9w7qTXL0t/9gt6hywW/XQmj1KmcV+fgJTZC8moD57QtNS0TggzhDuswfKQuwSxx3ELRh86Z5e+FsbMI1C2gQ14nfPEb4S2AsJPASclhprNN/E1af+OLlxJebAD/BAoGBAOTizpkUjS/NdJp7BtF4WH9bBc5sDYy4nAMLjQF9urXoL1C1B2xSrv7RVvIUx7CkfnBNrAqAvsxgb6Q/LztJwv5nCoZ9Y5Tt2o5mb9P6m6bVVM3iFTiWjqo8ytUz7CS09od5HhmyKu2ChOyr8jGarWBL+PS51Iwfe/s641pdl1MrAoGBAO/AC0jVULhdhChRfQGQZDEmxtiAbXat5KbzL0dZSrpIZjd4cDYW+o6XNSjsqNYBhMtIxUlqHcMe5dK8jQrSdeXrSYKHW1awOkqY4iMq0kQ2tdGdPcQRG13jGsnreyO4wA+df4H5YHo9A4SKxExKcKB970/XSEGjrBYOUcDyUyqfAoGARlTTOwKvp6KwU8+99pvORcQIcreNKlKHzf+8olqqBr+D2n7l+wklMLPOzbBI9CR3nbagSNHqzw5K/+NSdhtiSZ4MA+t/sAGuiNc9QZvePFONLX5tGuhYikMH6J99zoG0x0gWUbsHqdfTVI45a7il0dNGepynjS8Xf8lGlzvvBeUCgYEAttXTBUFAZMlUbtbuKRIvhlhXDmaqk/Y7SKJubNAIsBVkdmsP0AAoJjPkI4iPnVzdI5Ykdj9J4TKf+901BorHxIZxsex92JdebOM4ma8fWUwLzoZGw050e14lYNWHPA+50G7A/aLrU21SUHLvDms6hvpjVZUNEpm6M7vJ1wY2LGsCgYEA2WFv9L0ezHvAGyLfVo5Jag1mOWm4+z09QY/gn/LiT68dBajJAp93DBSWegPiIR5Gug++3jahZs339/ep9t0wbTRpj3978lVoiwIwPVD3qyk8WlmVn/e5IuX4sjv30wXUJfFDwDP3vNUtuQcrBCfmbg9xsrmYMiGdRqi9PJT1TO0="
	)
	err = ioutil.WriteFile("var/conf/encrypted-config-value.key", []byte(rsaPubKey), 0644)
	require.NoError(t, err)
	err = ioutil.WriteFile("var/conf/encrypted-config-value.private", []byte(rsaPrivKey), 0644)
	require.NoError(t, err)

	kp, err := encryptedconfigvalue.LoadKeyPairFromDefaultPath()
	require.NoError(t, err)

	assert.Equal(t, encryptedconfigvalue.RSA, kp.Algorithm())
	pubSKWA, err := kp.PublicKey().ToSerializable()
	require.NoError(t, err)
	assert.Equal(t, rsaPubKey, pubSKWA.SerializedStringForm())
	privSKWA, err := kp.PrivateKey().ToSerializable()
	require.NoError(t, err)
	assert.Equal(t, rsaPrivKey, privSKWA.SerializedStringForm())
}
