/*
 * go-udb-parser
 *
 * Copyright (c) 2016, Forest Crossman <cyrozap@gmail.com>
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	NumPLDs         = 2
	NumOrTerms      = 4
	NumMacrocells   = 4
	NumProductTerms = 8
	NumInputTerms   = 12
)

type xorfb int

const (
	dff xorfb = iota
	carry
	tffHigh
	tffLow
)

type InputTerm struct {
	True       [NumProductTerms]bool
	Complement [NumProductTerms]bool
}

type ProductTerm struct {
	Enabled [NumOrTerms]bool
}

type Macrocell struct {
	COEN  bool
	CONST bool
	XORFB xorfb
	RSEL  bool
	SSEL  bool
	BYP   bool
}

type PLD struct {
	InputTerms   [NumInputTerms]InputTerm
	ProductTerms [NumProductTerms]ProductTerm
	Macrocells   [NumMacrocells]Macrocell
}

type UDB struct {
	PLDs [NumPLDs]PLD
}

func ParseConfig(u *UDB, config []byte) error {
	if len(config) < 0x40 {
		return errors.New("config data too short")
	}

	for i := 0; i < 0x30; i++ {
		pld := i & 1
		it := i >> 2
		tc := (i & 2) != 0
		for pt := 0; pt < NumProductTerms; pt++ {
			value := (config[i] & (1 << uint(pt))) != 0
			if tc {
				u.PLDs[pld].InputTerms[it].True[pt] = value
			} else {
				u.PLDs[pld].InputTerms[it].Complement[pt] = value
			}
		}
	}

	for i := 0x30; i < 0x38; i++ {
		pld := i & 1
		ot := (i >> 1) & 3
		for pt := 0; pt < NumProductTerms; pt++ {
			value := (config[i] & (1 << uint(pt))) != 0
			u.PLDs[pld].ProductTerms[pt].Enabled[ot] = value
		}
	}

	for i := 0x38; i < 0x40; i++ {
		pld := i & 1
		mct := (i >> 1) & 3
		for bit := 0; bit < 8; bit++ {
			mc := (bit >> 1) & 3
			switch mct {
			case 0: // CEN_CONST
				cc := (config[i] & (1 << uint(bit))) != 0
				if (bit & 1) == 0 {
					u.PLDs[pld].Macrocells[mc].COEN = cc
				} else {
					u.PLDs[pld].Macrocells[mc].CONST = cc
				}
			case 1: // XORFB
				if (bit & 1) == 0 {
					u.PLDs[pld].Macrocells[mc].XORFB = xorfb((config[i] >> uint(bit)) & 3)
				}
			case 2: // SET_RESET
				sr := (config[i] & (1 << uint(bit))) != 0
				if (bit & 1) == 0 {
					u.PLDs[pld].Macrocells[mc].SSEL = sr
				} else {
					u.PLDs[pld].Macrocells[mc].RSEL = sr
				}
			case 3: // BYPASS
				if (bit & 1) == 0 {
					bypass := (config[i] & (1 << uint(bit))) != 0
					u.PLDs[pld].Macrocells[mc].BYP = bypass
				}
			}
		}
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s udb_config.bin\n", os.Args[0])
		os.Exit(1)
	}
	f := os.Args[1]
	config, err := ioutil.ReadFile(f)
	if err != nil {
		log.Fatal(err)
	}

	u := new(UDB)
	err = ParseConfig(u, config)
	if err != nil {
		log.Fatal(err)
	}
	dump, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", dump)
}
