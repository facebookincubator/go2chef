package certs

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import "testing"

var sampleCert = `-----BEGIN CERTIFICATE-----
MIIDZTCCAk2gAwIBAgIUW1yJcKJrx6BZ8s4YsN+QSp76h9YwDQYJKoZIhvcNAQEL
BQAwQjELMAkGA1UEBhMCWFgxFTATBgNVBAcMDERlZmF1bHQgQ2l0eTEcMBoGA1UE
CgwTRGVmYXVsdCBDb21wYW55IEx0ZDAeFw0xOTA5MDQyMzAyNDhaFw0yMDA5MDMy
MzAyNDhaMEIxCzAJBgNVBAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAa
BgNVBAoME0RlZmF1bHQgQ29tcGFueSBMdGQwggEiMA0GCSqGSIb3DQEBAQUAA4IB
DwAwggEKAoIBAQDFYPFUVUKMZNIoWS7FzNovn5uHM54Zs4uMkquf2sbXr3XVGZdq
L5gSF21MaGqjSdPDVZ1w7mTn1ufhcERXD15z/rn2N33MJZGP6vEDsOa2IaVvYncw
frRXYm1JGS9/wjxF1g4ZZFzKt6I72xbuDeiDJqlAlwKh4NlYwKj7jb0aHwWV8Ofr
LQUGV3mZIQ2KE2oGFIoF8h5Y9ryPqYGS+F9aON3l2qNE0apjhiv3R3A83DBzHYUA
CUO5RFdDvJ5RkFLFFqgdgW+S2hkc7pTVvXBQZrGGHhWrtQhLrJ9XdP6CUKieC1os
qoJwObmRW02jHmn3nkaN82JPd5K7jlcAO3EjAgMBAAGjUzBRMB0GA1UdDgQWBBQw
5GripBKPRjSptYULkv7payTvMzAfBgNVHSMEGDAWgBQw5GripBKPRjSptYULkv7p
ayTvMzAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQCO3/nVFfbf
i0vkv/i5hO/7WG+uOQllqOv8vIMhLK5zkZ7qWr4W+gLiGZ4vDHArps5xKygqKlqz
XB65eb5HWO5SjRPWVvrh+O4DF05CFcmXzKBW2afcCGUV1BXT6xPX9g31xknNVgtg
rBUKnXqDfqjG5dyy7H9+Nzl8zubRRb+6FVpsXbHxVz+uVBMESTlrw8ZBGDhetsUH
H3GM89h3KOhT36xDMTpfM49MAMXb9QaejZg6C4dbJ2zPyVaUdbU56EycXKi8JY0F
pse+M9g1j6JUGmw+3NC7uiBeRutkaop14u0FVMtf8SssJccH8eYl7v0jRDL/AeSc
BYd0Cl4SoZtP
-----END CERTIFICATE-----`
var sampleKey = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDFYPFUVUKMZNIo
WS7FzNovn5uHM54Zs4uMkquf2sbXr3XVGZdqL5gSF21MaGqjSdPDVZ1w7mTn1ufh
cERXD15z/rn2N33MJZGP6vEDsOa2IaVvYncwfrRXYm1JGS9/wjxF1g4ZZFzKt6I7
2xbuDeiDJqlAlwKh4NlYwKj7jb0aHwWV8OfrLQUGV3mZIQ2KE2oGFIoF8h5Y9ryP
qYGS+F9aON3l2qNE0apjhiv3R3A83DBzHYUACUO5RFdDvJ5RkFLFFqgdgW+S2hkc
7pTVvXBQZrGGHhWrtQhLrJ9XdP6CUKieC1osqoJwObmRW02jHmn3nkaN82JPd5K7
jlcAO3EjAgMBAAECggEAH35VY2hrQu1/XvLD9Mm38qtb7Jm+20j7tkVc3xfQbG/R
tFvt/gJ0GEbmqK9sfHt2L4/EnFVdgmSXATChpuaL4qQ9Vd0K1H0WGcmaBUW/ukXq
GLi0XeeJrPvGhkhffNooNdhuzXxnFe1xFG3j3b4YYHzVurmdsOiopXGwRNsb1kPe
l4V+jwzdLdsz2nHlAyJuKAovPp3VEUXAV70m+Q+y2d6OXqsoDGX4SUfKQTdPiH/Z
+CfaQ2CAJBYLMkTYIKaH596A9ShfLuQ5bZitOrG/xe4p9VDMIc4QEo8tuKAa2wVv
OwSRZh9PftfQEyzal9wxRrgj8EXwgShd4flTYe23IQKBgQD/K1p6+3Egqfth9o/Z
1PjMAWdetxOD1PGT2SZg+VBKlw9Shs9MGLyv+odsffMYVs5XU2/rQBQcQXj2eg0b
7YJftQGzzUyEFedJ9z8u7MfYYbshWNdZ6a2KqJU45XP5+5l3rhkenTTIS8mME8Lr
+HeoEY5xmOs3Ntvu3FYRaDmekQKBgQDGBW3bs1pacMSJSRHv/eAyrl8whwcN5KCf
xwylNa7wH+MbKxtVa6J8vfFqxNDP3Hj3aCWY4yrab4otffbWfwRECW/6zyI7fHkt
vzqNtRWN86p22f9JSyTMOyNXE2/0ah1Lznw/XNC7rk8Of2iZkG/wzoy4COOfk0lV
U/DsxQzWcwKBgQCXUk1xG0XmWgey+7YpN0xoJvj3SViwWIr+48sHvTIpWdYDWeD7
Prw/HDJNW4/bQjdRwDBh8Xk7nHQwrwaxJjOnsD8XMsuKlTa5PX/hwxdsseB4kSf8
sUByNzFvMVuKxvMm7z8EUbQoiBE5GcsBhzLmn6q6oTX0Y3sf9tivsABjkQKBgBJY
vmz0mRJ4ED2H/5l0tCj97uPYHtcyr48eKhXEe4jT637A569qYYudLZju00nu62ZA
x/r6USYb33mHii8lZYfIOA/M0SchyThr10j51h1ozgpk+DoaNDaX5BZVPrIugrhb
UTetqck5xSlatJ5Fu5lcCb2jVTObudemB1RojV/xAoGAAQY1hsHN24PToeNQheUA
tjLcI0p6XodDruOBd35tMTmQVYVRd45WbpvCKqBhrbChOr0tEDxNscgGWt0uGKiX
c16DuFO11akP5i4ifDolgGCwX/Wf3UYOw71K3ZbShx4zENbHHNby0Vi6UOQtfUzG
8ZczrTaTlOqhHY547PKlYD8=
-----END PRIVATE KEY-----`

func TestNewTLSConfiguration(t *testing.T) {
	tc := NewTLSConfiguration()
	if tc == nil {
		t.Fatalf("NewTLSConfiguration returned nil, should return pointer")
	}
}

func TestLoadTLSConfigurationFromMap(t *testing.T) {
	input := map[string]interface{}{
		"trusted_ca_certs": []string{sampleCert},
		"client_certs": []map[string]string{
			{
				"certificate": sampleCert,
				"key":         sampleKey,
			}},
		"disable_cert_verification": true,
	}
	tc, err := LoadTLSConfigurationFromMap(&input)
	if err != nil {
		t.Fatalf("failed to load TLS configuration from input map: %s", err)
	}
	if len(tc.TrustedCACerts) != 1 {
		t.Errorf("wrong number of CA certs loaded: %d", len(tc.TrustedCACerts))
	}
	if len(tc.ClientCerts) != 1 {
		t.Errorf("wrong number of client certs loaded: %d", len(tc.ClientCerts))
	}
	if !tc.DisableCertVerification {
		t.Errorf("expected DisableCertVerification to be true after load")
	}
}
