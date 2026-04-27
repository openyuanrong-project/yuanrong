package tls

import (
	"crypto/x509"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

// TestLoadRootCAs is used to test the root certificate loading error.
func TestLoadRootCAs(t *testing.T) {
	convey.Convey("LoadRootCAs", t, func() {
		convey.Convey("error case 1", func() {
			caFiles := ""
			_, err := LoadRootCAs(caFiles)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("error case 2", func() {
			dir, err := ioutil.TempDir("", "*")
			require.NoError(t, err)
			cryptoFile, err := ioutil.TempFile(dir, "crypto")
			require.NoError(t, err)
			_, err = LoadRootCAs(cryptoFile.Name())
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("error case 3", func() {
			dir, err := ioutil.TempDir("", "*")
			require.NoError(t, err)
			cryptoFile, err := ioutil.TempFile(dir, "test")
			require.NoError(t, err)
			err = ioutil.WriteFile(filepath.Join(dir, "test"), []byte("a"), os.ModePerm)
			require.NoError(t, err)
			_, err = LoadRootCAs(cryptoFile.Name())
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestVerifyCert2 is used to test certificate modification errors.
func TestVerifyCert2(t *testing.T) {
	convey.Convey("VerifyCert", t, func() {
		convey.Convey("error case 1", func() {
			rawCerts := [][]byte{}
			verifiedChains := [][]*x509.Certificate{}
			err := VerifyCert(rawCerts, verifiedChains)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "cert number is 0")
		})

		convey.Convey("error case 2", func() {
			rawCerts := [][]byte{[]byte("test1"), []byte("test2")}
			verifiedChains := [][]*x509.Certificate{}
			err := VerifyCert(rawCerts, verifiedChains)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("success", func() {
			defer gomonkey.ApplyFunc(x509.ParseCertificate, func(der []byte) (*x509.Certificate, error) {
				return &x509.Certificate{}, nil
			}).Reset()
			defer gomonkey.ApplyMethod(reflect.TypeOf(&x509.Certificate{}), "Verify",
				func(_ *x509.Certificate, opts x509.VerifyOptions) (chains [][]*x509.Certificate, err error) {
					return nil, nil
				}).Reset()
			rawCerts := [][]byte{[]byte("test1"), []byte("test2")}
			verifiedChains := [][]*x509.Certificate{}
			err := VerifyCert(rawCerts, verifiedChains)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
