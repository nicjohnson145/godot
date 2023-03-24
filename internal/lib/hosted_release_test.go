package lib

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestAutoDetectAsset(t *testing.T) {
	testData := []struct {
		name   string
		assets []string
		os     string
		arch   string
		out    string
	}{
		{
			name: "zoxide",
			assets: []string{
				"zoxide-0.9.0-aarch64-apple-darwin.tar.gz",
				"zoxide-0.9.0-aarch64-pc-windows-msvc.zip",
				"zoxide-0.9.0-aarch64-unknown-linux-musl.tar.gz",
				"zoxide-0.9.0-arm-unknown-linux-musleabihf.tar.gz",
				"zoxide-0.9.0-armv7-unknown-linux-musleabihf.tar.gz",
				"zoxide-0.9.0-x86_64-apple-darwin.tar.gz",
				"zoxide-0.9.0-x86_64-pc-windows-msvc.zip",
				"zoxide-0.9.0-x86_64-unknown-linux-musl.tar.gz",
				"zoxide_0.9.0_amd64.deb",
				"zoxide_0.9.0_arm64.deb",
			},
			os:   "linux",
			arch: "amd64",
			out:  "zoxide-0.9.0-x86_64-unknown-linux-musl.tar.gz",
		},
		{
			name: "tagbot",
			assets: []string{
				"tagbot_1.3.0_checksums.txt",
				"tagbot_darwin_amd64",
				"tagbot_linux_amd64",
				"tagbot_windows_amd64.exe",
			},
			os: "linux",
			arch: "amd64",
			out: "tagbot_linux_amd64",
		},
		{
			name: "protoc-gen-grpc-web",
			assets: []string{
				"protoc-gen-grpc-web-1.4.2-darwin-aarch64",
				"protoc-gen-grpc-web-1.4.2-darwin-aarch64.sha256",
				"protoc-gen-grpc-web-1.4.2-darwin-x86_64",
				"protoc-gen-grpc-web-1.4.2-darwin-x86_64.sha256",
				"protoc-gen-grpc-web-1.4.2-linux-aarch64",
				"protoc-gen-grpc-web-1.4.2-linux-aarch64.sha256",
				"protoc-gen-grpc-web-1.4.2-linux-x86_64",
				"protoc-gen-grpc-web-1.4.2-linux-x86_64.sha256",
				"protoc-gen-grpc-web-1.4.2-windows-aarch64.exe",
				"protoc-gen-grpc-web-1.4.2-windows-aarch64.exe.sha256",
				"protoc-gen-grpc-web-1.4.2-windows-x86_64.exe",
				"protoc-gen-grpc-web-1.4.2-windows-x86_64.exe.sha256",
			},
			os: "linux",
			arch: "amd64",
			out: "protoc-gen-grpc-web-1.4.2-linux-x86_64",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			assets := lo.Map(tc.assets, func(name string, _ int) asset {
				return asset{Name: name}
			})

			got, err := autoDetectAsset(LoggerWithLevel(zerolog.ErrorLevel), assets, tc.os, tc.arch)
			require.NoError(t, err)
			require.Equal(t, tc.out, got.Name)
		})
	}
}
