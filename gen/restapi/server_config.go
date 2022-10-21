// Code generated by go-swagger; DO NOT EDIT.

package restapi

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	FlagScheme           = "scheme"
	FlagCleanupTimeout   = "cleanup-timeout"
	FlagGracefulTimeout  = "graceful-timeout"
	FlagMaxHeaderSize    = "max-header-size"
	FlagListenAddress    = "listen-address"
	FlagListenLimit      = "listen-limit"
	FlagKeepAlive        = "keep-alive"
	FlagReadTimeout      = "read-timeout"
	FlagWriteTimeout     = "write-timeout"
	FlagTLSListenAddress = "tls-listen-address"
	FlagTLSCertificate   = "tls-certificate"
	FlagTLSKey           = "tls-key"
	FlagTLSCa            = "tls-ca"
	FlagTLSListenLimit   = "tls-listen-limit"
	FlagTLSKeepAlive     = "tls-keep-alive"
	FlagTLSReadTimeout   = "tls-read-timeout"
	FlagTLSWriteTimeout  = "tls-write-timeout"
)

// BindDefaultFlag init default flag.
func BindDefaultFlags(flagSet *pflag.FlagSet) {
	flagSet.StringSlice(FlagScheme, defaultSchemes, "the listeners to enable, this can be repeated and defaults to the schemes in the swagger spec")

	flagSet.Duration(FlagCleanupTimeout, 10*time.Second, "grace period for which to wait before killing idle connections")
	flagSet.Duration(FlagGracefulTimeout, 15*time.Second, "grace period for which to wait before shutting down the server")
	flagSet.Int(FlagMaxHeaderSize, 1000000, "controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body")

	flagSet.String(FlagListenAddress, "localhost:8080", "the IP and port to listen on")
	flagSet.Int(FlagListenLimit, 0, "limit the number of outstanding requests")
	flagSet.Duration(FlagKeepAlive, 3*time.Minute, "sets the TCP keep-alive timeouts on accepted connections. It prunes dead TCP connections ( e.g. closing laptop mid-download)")
	flagSet.Duration(FlagReadTimeout, 30*time.Second, "maximum duration before timing out read of the request")
	flagSet.Duration(FlagWriteTimeout, 30*time.Second, "maximum duration before timing out write of the response")

	flagSet.String(FlagTLSListenAddress, "localhost:8081", "the IP and port to listen on")
	flagSet.String(FlagTLSCertificate, "", "the certificate file to use for secure connections")
	flagSet.String(FlagTLSKey, "", "the private key file to use for secure connections (without passphrase)")
	flagSet.String(FlagTLSCa, "", "the certificate authority certificate file to be used with mutual tls auth")
	flagSet.Int(FlagTLSListenLimit, 0, "limit the number of outstanding requests")
	flagSet.Duration(FlagTLSKeepAlive, 3*time.Minute, "sets the TCP keep-alive timeouts on accepted connections. It prunes dead TCP connections ( e.g. closing laptop mid-download)")
	flagSet.Duration(FlagTLSReadTimeout, 30*time.Second, "maximum duration before timing out read of the request")
	flagSet.Duration(FlagTLSWriteTimeout, 30*time.Second, "maximum duration before timing out write of the response")
}

// BindFlagsToConfig maps flags to viper config in specific section.
func BindFlagsToConfig(v *viper.Viper, flagSet *pflag.FlagSet, section string) error {
	if err := v.BindPFlag(section+FlagScheme, flagSet.Lookup(FlagScheme)); err != nil {
		return err
	}

	if err := v.BindPFlag(section+FlagCleanupTimeout, flagSet.Lookup(FlagCleanupTimeout)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagGracefulTimeout, flagSet.Lookup(FlagGracefulTimeout)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagMaxHeaderSize, flagSet.Lookup(FlagMaxHeaderSize)); err != nil {
		return err
	}

	if err := v.BindPFlag(section+FlagListenAddress, flagSet.Lookup(FlagListenAddress)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagListenLimit, flagSet.Lookup(FlagListenLimit)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagKeepAlive, flagSet.Lookup(FlagKeepAlive)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagReadTimeout, flagSet.Lookup(FlagReadTimeout)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagWriteTimeout, flagSet.Lookup(FlagWriteTimeout)); err != nil {
		return err
	}

	if err := v.BindPFlag(section+FlagTLSListenAddress, flagSet.Lookup(FlagTLSListenAddress)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagTLSCertificate, flagSet.Lookup(FlagTLSCertificate)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagTLSKey, flagSet.Lookup(FlagTLSKey)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagTLSCa, flagSet.Lookup(FlagTLSCa)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagTLSListenLimit, flagSet.Lookup(FlagTLSListenLimit)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagTLSKeepAlive, flagSet.Lookup(FlagTLSKeepAlive)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagTLSReadTimeout, flagSet.Lookup(FlagTLSReadTimeout)); err != nil {
		return err
	}
	if err := v.BindPFlag(section+FlagTLSWriteTimeout, flagSet.Lookup(FlagTLSWriteTimeout)); err != nil {
		return err
	}

	return nil
}
