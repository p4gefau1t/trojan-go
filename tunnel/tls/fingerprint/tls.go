package fingerprint

import (
	"crypto/tls"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	. "github.com/refraction-networking/utls"
)

func GetClientHelloSpec(name string, websocket bool) (*ClientHelloSpec, error) {
	// TODO fix websocket
	var spec *ClientHelloSpec
	switch name {
	case "chrome":
		spec = &ClientHelloSpec{
			CipherSuites: []uint16{
				GREASE_PLACEHOLDER,
				TLS_AES_128_GCM_SHA256,
				TLS_AES_256_GCM_SHA384,
				TLS_CHACHA20_POLY1305_SHA256,
				TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				TLS_RSA_WITH_AES_128_GCM_SHA256,
				TLS_RSA_WITH_AES_256_GCM_SHA384,
				TLS_RSA_WITH_AES_128_CBC_SHA,
				TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			CompressionMethods: []byte{
				0x00, // compressionNone
			},
			Extensions: []TLSExtension{
				&UtlsGREASEExtension{},
				&SNIExtension{},
				&UtlsExtendedMasterSecretExtension{},
				&RenegotiationInfoExtension{Renegotiation: RenegotiateOnceAsClient},
				&SupportedCurvesExtension{[]CurveID{
					CurveID(GREASE_PLACEHOLDER),
					X25519,
					CurveP256,
					CurveP384,
				}},
				&SupportedPointsExtension{SupportedPoints: []byte{
					0x00, // pointFormatUncompressed
				}},
				&SessionTicketExtension{},
				&ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
				&StatusRequestExtension{},
				&SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []SignatureScheme{
					ECDSAWithP256AndSHA256,
					PSSWithSHA256,
					PKCS1WithSHA256,
					ECDSAWithP384AndSHA384,
					PSSWithSHA384,
					PKCS1WithSHA384,
					PSSWithSHA512,
					PKCS1WithSHA512,
				}},
				&SCTExtension{},
				&KeyShareExtension{[]KeyShare{
					{Group: CurveID(GREASE_PLACEHOLDER), Data: []byte{0}},
					{Group: X25519},
				}},
				&PSKKeyExchangeModesExtension{[]uint8{
					PskModeDHE,
				}},
				&SupportedVersionsExtension{[]uint16{
					GREASE_PLACEHOLDER,
					VersionTLS13,
					VersionTLS12,
					VersionTLS11,
					VersionTLS10,
				}},
				&FakeCertCompressionAlgsExtension{[]CertCompressionAlgo{
					CertCompressionBrotli,
				}},
				&UtlsGREASEExtension{},
				&UtlsPaddingExtension{GetPaddingLen: BoringPaddingStyle},
			},
		}
	case "firefox":
		spec = &ClientHelloSpec{
			TLSVersMin: VersionTLS10,
			TLSVersMax: VersionTLS13,
			CipherSuites: []uint16{
				TLS_AES_128_GCM_SHA256,
				TLS_CHACHA20_POLY1305_SHA256,
				TLS_AES_256_GCM_SHA384,
				TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				FAKE_TLS_DHE_RSA_WITH_AES_128_CBC_SHA,
				FAKE_TLS_DHE_RSA_WITH_AES_256_CBC_SHA,
				TLS_RSA_WITH_AES_128_CBC_SHA,
				TLS_RSA_WITH_AES_256_CBC_SHA,
				TLS_RSA_WITH_3DES_EDE_CBC_SHA,
			},
			CompressionMethods: []byte{
				0, //compressionNone,
			},
			Extensions: []TLSExtension{
				&SNIExtension{},
				&UtlsExtendedMasterSecretExtension{},
				&RenegotiationInfoExtension{Renegotiation: RenegotiateOnceAsClient},
				&SupportedCurvesExtension{[]CurveID{
					X25519,
					CurveP256,
					CurveP384,
					CurveP521,
					CurveID(FakeFFDHE2048),
					CurveID(FakeFFDHE3072),
				}},
				&SupportedPointsExtension{SupportedPoints: []byte{
					0, //pointFormatUncompressed,
				}},
				&SessionTicketExtension{},
				&ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
				&StatusRequestExtension{},
				&KeyShareExtension{[]KeyShare{
					{Group: X25519},
					{Group: CurveP256},
				}},
				&SupportedVersionsExtension{[]uint16{
					VersionTLS13,
					VersionTLS12,
					VersionTLS11,
					VersionTLS10}},
				&SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []SignatureScheme{
					ECDSAWithP256AndSHA256,
					ECDSAWithP384AndSHA384,
					ECDSAWithP521AndSHA512,
					PSSWithSHA256,
					PSSWithSHA384,
					PSSWithSHA512,
					PKCS1WithSHA256,
					PKCS1WithSHA384,
					PKCS1WithSHA512,
					ECDSAWithSHA1,
					PKCS1WithSHA1,
				}},
				&PSKKeyExchangeModesExtension{[]uint8{1 /*pskModeDHE*/}},
				&FakeRecordSizeLimitExtension{0x4001},
				&UtlsPaddingExtension{GetPaddingLen: BoringPaddingStyle},
			}}
	case "ios":
		spec = &ClientHelloSpec{
			CipherSuites: []uint16{
				TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				DISABLED_TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384,
				TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
				TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				DISABLED_TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384,
				TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
				TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				TLS_RSA_WITH_AES_256_GCM_SHA384,
				TLS_RSA_WITH_AES_128_GCM_SHA256,
				DISABLED_TLS_RSA_WITH_AES_256_CBC_SHA256,
				TLS_RSA_WITH_AES_128_CBC_SHA256,
				TLS_RSA_WITH_AES_256_CBC_SHA,
				TLS_RSA_WITH_AES_128_CBC_SHA,
				0xc008,
				TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
				TLS_RSA_WITH_3DES_EDE_CBC_SHA,
			},
			CompressionMethods: []byte{
				0, //compressionNone,
			},
			Extensions: []TLSExtension{
				&RenegotiationInfoExtension{Renegotiation: RenegotiateOnceAsClient},
				&SNIExtension{},
				&UtlsExtendedMasterSecretExtension{},
				&SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []SignatureScheme{
					ECDSAWithP256AndSHA256,
					PSSWithSHA256,
					PKCS1WithSHA256,
					ECDSAWithP384AndSHA384,
					ECDSAWithSHA1,
					PSSWithSHA384,
					PSSWithSHA384,
					PKCS1WithSHA384,
					PSSWithSHA512,
					PKCS1WithSHA512,
					PKCS1WithSHA1,
				}},
				&StatusRequestExtension{},
				&NPNExtension{},
				&SCTExtension{},
				&ALPNExtension{AlpnProtocols: []string{"h2", "h2-16", "h2-15", "h2-14", "spdy/3.1", "spdy/3", "http/1.1"}},
				&SupportedPointsExtension{SupportedPoints: []byte{
					0, //pointFormatUncompressed,
				}},
				&SupportedCurvesExtension{[]CurveID{
					X25519,
					CurveP256,
					CurveP384,
					CurveP521,
				}},
			},
		}
	}
	if spec == nil {
		return nil, common.NewError("Invalid fingerprint:" + name)
	}
	if websocket {
		for i := range spec.Extensions {
			if alpn, ok := spec.Extensions[i].(*ALPNExtension); ok {
				alpn.AlpnProtocols = []string{"http/1.1"}
				spec.Extensions[i] = alpn
				log.Debug("websocket http/1.1")
			}
		}
	}
	return spec, nil
}

func ParseCipher(s []string) []uint16 {
	all := tls.CipherSuites()
	var result []uint16
	for _, p := range s {
		found := true
		for _, q := range all {
			if q.Name == p {
				result = append(result, q.ID)
				break
			}
			if !found {
				log.Warn("invalid cipher suite", p, "skipped")
			}
		}
	}
	return result
}
