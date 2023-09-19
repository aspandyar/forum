package main

import "crypto/tls"

func ConfigureCipherSuites(config *tls.Config, cipherSuites []uint16) {
	config.CipherSuites = cipherSuites
}

func SetMinTLSVersion(config *tls.Config, version uint16) {
	config.MinVersion = version
}

func SetMaxTLSVersion(config *tls.Config, version uint16) {
	config.MaxVersion = version
}
