package schema

import (
	"strings"
	"testing"
)

func TestVarintEncodingZero(t *testing.T) {
	var num Varint = 0

	if text := num.Encode(); text != "0" {
		t.Fatalf("Expected %v but was %v", "0", text)
	}
}

func TestVarintEncodingLarge(t *testing.T) {
	var num Varint = 72374

	if text := num.Encode(); text != "k15m6" {
		t.Fatalf("Expected %v but was %v", "k15m6", text)
	}
}

func TestVarintEncodingNegative(t *testing.T) {
	var num Varint = -1

	if text := num.Encode(); text != "weyyyyyyyyyyyf" {
		t.Fatalf("Expected %v but was %v", "weyyyyyyyyyyyf", text)
	}
}

func TestVarintEncodingSingle(t *testing.T) {
	var num Varint = 0x10

	if text := num.Encode(); text != "g0" {
		t.Fatalf("Expected %v but was %v", "g0", text)
	}
}

func TestVarintDecodingZero(t *testing.T) {
	var num Varint = -1

	consumed, err := num.Decode("0")
	if err != nil {
		t.Fatal(err)
	}
	if num != 0 {
		t.Fatalf("Expected %v but was %v", 0, num)
	}
	if consumed != 1 {
		t.Fatal("not all bytes consumed")
	}
}

func TestVarintDecodingLarge(t *testing.T) {
	var num Varint

	consumed, err := num.Decode("k15m6")
	if err != nil {
		t.Fatal(err)
	}
	if num != 72374 {
		t.Fatalf("Expected %v but was %v", 72374, num)
	}
	if consumed != 5 {
		t.Fatal("not all bytes consumed")
	}
}

func TestVarintDecodingNegative(t *testing.T) {
	var num Varint

	consumed, err := num.Decode("weyyyyyyyyyyyf")
	if err != nil {
		t.Fatal(err)
	}
	if num != -1 {
		t.Fatalf("Expected %v but was %v", -1, num)
	}
	if consumed != 14 {
		t.Fatal("not all bytes consumed")
	}
}

func TestVarintDecodingSingle(t *testing.T) {
	var num Varint

	consumed, err := num.Decode("g0")
	if err != nil {
		t.Fatal(err)
	}
	if num != 16 {
		t.Fatalf("Expected %v but was %v", 16, num)
	}
	if consumed != 2 {
		t.Fatal("not all bytes consumed")
	}
}

func TestVarintDecodingMatchCase(t *testing.T) {
	var num Varint

	consumed, err := num.Decode("HZ")
	if err == nil {
		t.Fatal("expected an error")
	}
	if consumed != 0 {
		t.Fatalf("Expected no bytes to be consumed")
	}
	if num != 0 {
		t.Fatal("Should not have changed on bad input")
	}
}

func TestVarintDecodingFailOnInvalidInput(t *testing.T) {
	var num Varint

	consumed, err := num.Decode("gu")
	if err == nil {
		t.Fatal("expected an error")
	}
	if consumed != 0 {
		t.Fatalf("Expected no bytes to be consumed")
	}
	if num != 0 {
		t.Fatal("Should not have changed on bad input")
	}
}

func TestVarintDecodingFailOnEmpty(t *testing.T) {
	var num Varint = 3

	consumed, err := num.Decode("")
	if err == nil {
		t.Fatal("expected an error")
	}
	if consumed != 0 {
		t.Fatalf("Expected no bytes to be consumed")
	}
	if num != 3 {
		t.Fatal("Should not have changed on bad input")
	}
}

func TestVarintDecodingFailOnTooShort(t *testing.T) {
	var num Varint = 3

	consumed, err := num.Decode("h")
	if err == nil {
		t.Fatal("expected an error")
	}
	if consumed != 0 {
		t.Fatalf("Expected no bytes to be consumed")
	}
	if num != 3 {
		t.Fatal("Should not have changed on bad input")
	}
}

func TestVarintDecodingSucceedsOnExcess(t *testing.T) {
	var num Varint = 3

	consumed, err := num.Decode("00")
	if err != nil {
		t.Fatal(err)
	}
	if num != 0 {
		t.Fatalf("Expected %v but was %v", 0, num)
	}
	if consumed != 1 {
		t.Fatal("wrong number of bytes consumed")
	}
}

func TestDecodeFailsOnOverflow(t *testing.T) {
	v := new(Varint)
	_, err := v.Decode("y00b00es00000000")
	if err == nil {
		t.Fatal("expected overflow")
	}
}

func TestDecodeFailsOnOverflowUppercase(t *testing.T) {
	v := new(Varint)
	_, err := v.Decode("Y00b00es00000000")
	if err == nil {
		t.Fatal("expected overflow")
	}
}

func TestRoundTripLowers(t *testing.T) {
	cases := []string{
		"A",
	}
	for _, data := range cases {
		v := new(Varint)
		n, err := v.Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		out := v.Encode()
		if strings.Compare(out, strings.ToLower(data[:n])) != 0 {
			t.Log("mismatch! ", out, strings.ToLower(data[:n]))
			t.Fail()
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		num := Varint(i)
		num.Encode()
	}
}

func BenchmarkRoundTrip(b *testing.B) {
	for i := 0; i < b.N; i++ {
		num := Varint(i)
		text := num.Encode()
		num = 0
		if err := num.DecodeAll(text); err != nil || int(num) != i {
			b.Fatal("bad encode")
		}
	}
}
