// Copyright 2024 VNXME
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package l4rdp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"

	"github.com/96368a/caddy-l4/layer4"
)

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Unexpected error: %s\n", err)
	}
}

func Test_MatchRDP_ProcessTPKTHeader(t *testing.T) {
	p := [][]byte{
		packetValid1[0:4], packetValid2[0:4], packetValid3[0:4], packetValid4[0:4],
		packetValid5[0:4], packetValid6[0:4], packetValid7[0:4], packetValid8[0:4], packetValid9[0:4],
	}
	for _, b := range p {
		func() {
			s := &TPKTHeader{}
			errFrom := s.FromBytes(b)
			assertNoError(t, errFrom)
			sb, errTo := s.ToBytes()
			assertNoError(t, errTo)
			if !bytes.Equal(b, sb) {
				t.Fatalf("test %T bytes processing: resulting bytes [% x] don't match original bytes [% x]", *s, b, sb)
			}
		}()
	}
}

func Test_MatchRDP_ProcessX224Crq(t *testing.T) {
	p := [][]byte{
		packetValid1[4:11], packetValid2[4:11], packetValid3[4:11], packetValid4[4:11],
		packetValid5[4:11], packetValid6[4:11], packetValid7[4:11], packetValid8[4:11], packetValid9[4:11],
	}
	for _, b := range p {
		func() {
			s := &X224Crq{}
			errFrom := s.FromBytes(b)
			assertNoError(t, errFrom)
			sb, errTo := s.ToBytes()
			assertNoError(t, errTo)
			if !bytes.Equal(b, sb) {
				t.Fatalf("test %T bytes processing: resulting bytes [% x] don't match original bytes [% x]", *s, b, sb)
			}
		}()
	}
}

func Test_MatchRDP_ProcessRDPCookie(t *testing.T) {
	p := [][]byte{
		packetValid3[11:35], packetValid4[11:35], packetValid6[11:35],
	}
	for _, b := range p {
		func() {
			s := string(b)
			if s != RDPCookiePrefix+"a0123"+string(ASCIIByteCR)+string(ASCIIByteLF) {
				t.Fatalf("test RDPCookie bytes processing: resulting bytes [% x] don't match original bytes [% x]", b, []byte(s))
			}
		}()
	}
}

func Test_MatchRDP_ProcessRDPToken(t *testing.T) {
	p := [][]byte{
		packetValid5[11:56], packetValid7[11:56], packetValid8[11:56],
	}
	for _, b := range p {
		func() {
			s := &RDPToken{}
			errFrom := s.FromBytes(b)
			assertNoError(t, errFrom)
			sb, errTo := s.ToBytes()
			assertNoError(t, errTo)
			if !bytes.Equal(b, sb) {
				t.Fatalf("test %T bytes processing: resulting bytes [% x] don't match original bytes [% x]", *s, b, sb)
			}
		}()
	}
}

func Test_MatchRDP_ProcessRDPNegReq(t *testing.T) {
	p := [][]byte{
		packetValid1[11:19], packetValid2[11:19], packetValid3[35:43], packetValid4[35:43],
		packetValid5[56:64], packetValid8[56:64],
	}
	for _, b := range p {
		func() {
			s := &RDPNegReq{}
			errFrom := s.FromBytes(b)
			assertNoError(t, errFrom)
			sb, errTo := s.ToBytes()
			assertNoError(t, errTo)
			if !bytes.Equal(b, sb) {
				t.Fatalf("test %T bytes processing: resulting bytes [% x] don't match original bytes [% x]", *s, b, sb)
			}
		}()
	}
}

func Test_MatchRDP_ProcessRDPCorrInfo(t *testing.T) {
	p := [][]byte{
		packetValid2[19:55], packetValid3[43:79], packetValid8[64:100],
	}
	for _, b := range p {
		func() {
			s := &RDPCorrInfo{}
			errFrom := s.FromBytes(b)
			assertNoError(t, errFrom)
			sb, errTo := s.ToBytes()
			assertNoError(t, errTo)
			if !bytes.Equal(b, sb) {
				t.Fatalf("test %T bytes processing: resulting bytes [% x] don't match original bytes [% x]", *s, b, sb)
			}
		}()
	}
}

func Test_MatchRDP_Match(t *testing.T) {
	type test struct {
		matcher     *MatchRDP
		data        []byte
		shouldMatch bool
	}

	tests := []test{
		// without filters
		{matcher: &MatchRDP{}, data: packetTooShort, shouldMatch: false},
		{matcher: &MatchRDP{}, data: packetInvalid1, shouldMatch: false},
		{matcher: &MatchRDP{}, data: packetInvalid2, shouldMatch: false},
		{matcher: &MatchRDP{}, data: packetInvalid3, shouldMatch: false},
		{matcher: &MatchRDP{}, data: packetInvalid4, shouldMatch: false},
		{matcher: &MatchRDP{}, data: packetInvalid5, shouldMatch: false},
		{matcher: &MatchRDP{}, data: packetInvalid6, shouldMatch: false},
		{matcher: &MatchRDP{}, data: packetSemiValid1, shouldMatch: false},
		{matcher: &MatchRDP{}, data: packetSemiValid2, shouldMatch: false},
		{matcher: &MatchRDP{}, data: packetSemiValid3, shouldMatch: false},
		{matcher: &MatchRDP{}, data: packetSemiValid4, shouldMatch: false},
		{matcher: &MatchRDP{}, data: packetValid1, shouldMatch: true},
		{matcher: &MatchRDP{}, data: packetValid2, shouldMatch: true},
		{matcher: &MatchRDP{}, data: packetValid3, shouldMatch: true},
		{matcher: &MatchRDP{}, data: packetValid4, shouldMatch: true},
		{matcher: &MatchRDP{}, data: packetValid5, shouldMatch: true},
		{matcher: &MatchRDP{}, data: packetValid6, shouldMatch: true},
		{matcher: &MatchRDP{}, data: packetValid7, shouldMatch: true},
		{matcher: &MatchRDP{}, data: packetValid8, shouldMatch: true},
		{matcher: &MatchRDP{}, data: packetValid9, shouldMatch: true},
		{matcher: &MatchRDP{}, data: packetExtraByte, shouldMatch: false},
		// with filtered hash
		{matcher: &MatchRDP{CookieHash: ""}, data: packetValid3, shouldMatch: true},
		{matcher: &MatchRDP{CookieHash: "a0123"}, data: packetValid3, shouldMatch: true},
		{matcher: &MatchRDP{CookieHash: "admin"}, data: packetValid3, shouldMatch: false},
		{matcher: &MatchRDP{CookieHashRegexp: ""}, data: packetValid3, shouldMatch: true},
		{matcher: &MatchRDP{CookieHashRegexp: "^[a-z]\\d+$"}, data: packetValid3, shouldMatch: true},
		{matcher: &MatchRDP{CookieHashRegexp: "^[A-Z]\\d+$"}, data: packetValid3, shouldMatch: false},
		// with filtered port
		{matcher: &MatchRDP{CookiePorts: []uint16{}}, data: packetValid5, shouldMatch: true},
		{matcher: &MatchRDP{CookiePorts: []uint16{3389}}, data: packetValid5, shouldMatch: true},
		{matcher: &MatchRDP{CookiePorts: []uint16{5000}}, data: packetValid5, shouldMatch: false},
		// with filtered IP
		{matcher: &MatchRDP{CookieIPs: []string{}}, data: packetValid7, shouldMatch: true},
		{matcher: &MatchRDP{CookieIPs: []string{"127.0.0.1/8"}}, data: packetValid7, shouldMatch: true},
		{matcher: &MatchRDP{CookieIPs: []string{"192.168.0.1/16"}}, data: packetValid7, shouldMatch: false},
		// with filtered info
		{matcher: &MatchRDP{CustomInfo: ""}, data: packetValid9, shouldMatch: true},
		{matcher: &MatchRDP{CustomInfo: "anything could be here"}, data: packetValid9, shouldMatch: true},
		{matcher: &MatchRDP{CustomInfo: "arbitrary text"}, data: packetValid9, shouldMatch: false},
		{matcher: &MatchRDP{CustomInfoRegexp: ""}, data: packetValid9, shouldMatch: true},
		{matcher: &MatchRDP{CustomInfoRegexp: "^([A-Za-z0-9 ]+)$"}, data: packetValid9, shouldMatch: true},
		{matcher: &MatchRDP{CustomInfoRegexp: "^\\x00\\x01\\x02\\x03\\x04$"}, data: packetValid9, shouldMatch: false},
	}

	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.Background()})
	defer cancel()

	for i, tc := range tests {
		func() {
			err := tc.matcher.Provision(ctx)
			assertNoError(t, err)

			in, out := net.Pipe()
			defer func() {
				_, _ = io.Copy(io.Discard, out)
				_ = out.Close()
			}()

			cx := layer4.WrapConnection(out, []byte{}, zap.NewNop())
			go func() {
				_, err := in.Write(tc.data)
				assertNoError(t, err)
				_ = in.Close()
			}()

			matched, err := tc.matcher.Match(cx)
			assertNoError(t, err)

			if matched != tc.shouldMatch {
				if tc.shouldMatch {
					t.Fatalf("test %d: matcher did not match | %+v\n", i, tc.matcher)
				} else {
					t.Fatalf("test %d: matcher should not match | %+v\n", i, tc.matcher)
				}
			}
		}()
	}
}

// Packet examples
var packetTooShort = []byte{
	0x00, 0x00, 0x00, 0x00, // TPKTHeader
}
var packetInvalid1 = []byte{
	0x00, 0x00, 0x00, 0x00, // TPKTHeader
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
}
var packetInvalid2 = []byte{
	0x03, 0x00, 0x00, 0x0B, // TPKTHeader
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
}
var packetInvalid3 = []byte{
	0x03, 0x00, 0x00, 0x13, // TPKTHeader
	0x0E, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
}
var packetInvalid4 = []byte{
	0x03, 0x00, 0x00, 0x37, // TPKTHeader
	0x32, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x01, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
	0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (1/3)
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (2/3)
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (3/3)
}
var packetInvalid5 = []byte{
	0x03, 0x00, 0x00, 0x4F, // TPKTHeader
	0x4A, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x43, 0x6F, 0x6F, 0x6B, 0x69, 0x65, 0x3A, 0x20, 0x6D, 0x73, 0x74, 0x73, 0x68, 0x61, 0x73, 0x68, 0x3D, // RDPCookie (1/3)
	// RDPCookie (2/3)
	0x0D, 0x0A, // RDPCookie (3/3)
	0x01, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
	0x06, 0x00, 0x24, 0x00, // RDPCorrInfo (1/3)
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (2/3)
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (3/3)
}
var packetInvalid6 = []byte{
	0x03, 0x00, 0x00, 0x1E, // TPKTHeader
	0x19, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPToken (1/1)
	0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
}
var packetSemiValid1 = []byte{ // we can't be sure it's RDP
	0x03, 0x00, 0x00, 0x0B, // TPKTHeader
	0x06, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
}
var packetSemiValid2 = []byte{ // we assume cookie hash must have at least 1 symbol
	0x03, 0x00, 0x00, 0x4F, // TPKTHeader
	0x4A, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x43, 0x6F, 0x6F, 0x6B, 0x69, 0x65, 0x3A, 0x20, 0x6D, 0x73, 0x74, 0x73, 0x68, 0x61, 0x73, 0x68, 0x3D, // RDPCookie (1/3)
	// RDPCookie (2/3)
	0x0D, 0x0A, // RDPCookie (3/3)
	0x01, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
	0x06, 0x00, 0x24, 0x00, // RDPCorrInfo (1/3)
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (2/3)
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (3/3)
}
var packetSemiValid3 = []byte{ // we assume custom info must have at least 1 symbol
	0x03, 0x00, 0x00, 0x39, // TPKTHeader
	0x34, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	// RDPCustom (1/2)
	0x0D, 0x0A, // RDPCustom (2/2)
	0x01, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
	0x06, 0x00, 0x24, 0x00, // RDPCorrInfo (1/3)
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (2/3)
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (3/3)
}
var packetSemiValid4 = []byte{ // an empty RDPToken doesn't seem to be a valid part of RDP Connection Request
	0x03, 0x00, 0x00, 0x1E, // TPKTHeader
	0x19, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x03, 0x00, 0x00, 0x0B, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPToken (1/1)
	0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
}
var packetValid1 = []byte{
	0x03, 0x00, 0x00, 0x13, // TPKTHeader
	0x0E, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
}
var packetValid2 = []byte{
	0x03, 0x00, 0x00, 0x37, // TPKTHeader
	0x32, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x01, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
	0x06, 0x00, 0x24, 0x00, // RDPCorrInfo (1/3)
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (2/3)
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (3/3)
}
var packetValid3 = []byte{
	0x03, 0x00, 0x00, 0x4F, // TPKTHeader
	0x4A, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x43, 0x6F, 0x6F, 0x6B, 0x69, 0x65, 0x3A, 0x20, 0x6D, 0x73, 0x74, 0x73, 0x68, 0x61, 0x73, 0x68, 0x3D, // RDPCookie (1/3)
	0x61, 0x30, 0x31, 0x32, 0x33, // RDPCookie (2/3)
	0x0D, 0x0A, // RDPCookie (3/3)
	0x01, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
	0x06, 0x00, 0x24, 0x00, // RDPCorrInfo (1/3)
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (2/3)
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (3/3)
}
var packetValid4 = []byte{
	0x03, 0x00, 0x00, 0x2B, // TPKTHeader
	0x26, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x43, 0x6F, 0x6F, 0x6B, 0x69, 0x65, 0x3A, 0x20, 0x6D, 0x73, 0x74, 0x73, 0x68, 0x61, 0x73, 0x68, 0x3D, // RDPCookie (1/3)
	0x61, 0x30, 0x31, 0x32, 0x33, // RDPCookie (2/3)
	0x0D, 0x0A, // RDPCookie (3/3)
	0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
}
var packetValid5 = []byte{
	0x03, 0x00, 0x00, 0x40, // TPKTHeader
	0x3B, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x03, 0x00, 0x00, 0x2D, 0x28, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPToken (1/6)
	0x43, 0x6F, 0x6F, 0x6B, 0x69, 0x65, 0x3A, 0x20, 0x6D, 0x73, 0x74, 0x73, 0x3D, // RDPToken (2/6)
	0x31, 0x36, 0x37, 0x37, 0x37, 0x33, 0x34, 0x33, // RDPToken (3/6)
	0x2E, 0x31, 0x35, 0x36, 0x32, 0x39, // RDPToken (4/6)
	0x2E, 0x30, 0x30, 0x30, 0x30, // RDPToken (5/6)
	0x0D, 0x0A, // RDPToken (6/6)
	0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
}
var packetValid6 = []byte{
	0x03, 0x00, 0x00, 0x23, // TPKTHeader
	0x1E, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x43, 0x6F, 0x6F, 0x6B, 0x69, 0x65, 0x3A, 0x20, 0x6D, 0x73, 0x74, 0x73, 0x68, 0x61, 0x73, 0x68, 0x3D, // RDPCookie (1/3)
	0x61, 0x30, 0x31, 0x32, 0x33, // RDPCookie (2/3)
	0x0D, 0x0A, // RDPCookie (3/3)
}
var packetValid7 = []byte{
	0x03, 0x00, 0x00, 0x38, // TPKTHeader
	0x33, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x03, 0x00, 0x00, 0x2D, 0x28, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPToken (1/6)
	0x43, 0x6F, 0x6F, 0x6B, 0x69, 0x65, 0x3A, 0x20, 0x6D, 0x73, 0x74, 0x73, 0x3D, // RDPToken (2/6)
	0x31, 0x36, 0x37, 0x37, 0x37, 0x33, 0x34, 0x33, // RDPToken (3/6)
	0x2E, 0x31, 0x35, 0x36, 0x32, 0x39, // RDPToken (4/6)
	0x2E, 0x30, 0x30, 0x30, 0x30, // RDPToken (5/6)
	0x0D, 0x0A, // RDPToken (6/6)
}
var packetValid8 = []byte{
	0x03, 0x00, 0x00, 0x64, // TPKTHeader
	0x5F, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x03, 0x00, 0x00, 0x2D, 0x28, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPToken (1/6)
	0x43, 0x6F, 0x6F, 0x6B, 0x69, 0x65, 0x3A, 0x20, 0x6D, 0x73, 0x74, 0x73, 0x3D, // RDPToken (2/6)
	0x31, 0x36, 0x37, 0x37, 0x37, 0x33, 0x34, 0x33, // RDPToken (3/6)
	0x2E, 0x31, 0x35, 0x36, 0x32, 0x39, // RDPToken (4/6)
	0x2E, 0x30, 0x30, 0x30, 0x30, // RDPToken (5/6)
	0x0D, 0x0A, // RDPToken (6/6)
	0x01, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
	0x06, 0x00, 0x24, 0x00, // RDPCorrInfo (1/3)
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (2/3)
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (3/3)
}
var packetValid9 = []byte{
	0x03, 0x00, 0x00, 0x23, // TPKTHeader
	0x1E, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x61, 0x6E, 0x79, 0x74, 0x68, 0x69, 0x6E, 0x67, 0x20, 0x63, 0x6F, 0x75, 0x6C, 0x64, 0x20, // RDPCustom (1/3)
	0x62, 0x65, 0x20, 0x68, 0x65, 0x72, 0x65, // RDPCustom (2/3)
	0x0D, 0x0A, // RDPCustom (3/3)
}
var packetExtraByte = []byte{
	0x03, 0x00, 0x00, 0x64, // TPKTHeader
	0x5F, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // X224Crq
	0x03, 0x00, 0x00, 0x2D, 0x28, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPToken (1/6)
	0x43, 0x6F, 0x6F, 0x6B, 0x69, 0x65, 0x3A, 0x20, 0x6D, 0x73, 0x74, 0x73, 0x3D, // RDPToken (2/6)
	0x31, 0x36, 0x37, 0x37, 0x37, 0x33, 0x34, 0x33, // RDPToken (3/6)
	0x2E, 0x31, 0x35, 0x36, 0x32, 0x39, // RDPToken (4/6)
	0x2E, 0x30, 0x30, 0x30, 0x30, // RDPToken (5/6)
	0x0D, 0x0A, // RDPToken (6/6)
	0x01, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPNegReq
	0x06, 0x00, 0x24, 0x00, // RDPCorrInfo (1/3)
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (2/3)
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // RDPCorrInfo (3/3)
	0x00, // wrong byte
}
