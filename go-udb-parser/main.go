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
	"bytes"
	// "encoding/json"
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

func (u *UDB) LoadConfig(config []byte) error {
	if len(config) < 0x40 {
		return errors.New("config data too short")
	}

	for i := 0; i < 0x30; i++ {
		pld := i & 1
		it := i >> 2
		tc := (i & 2) != 0
		var pt uint
		for pt = 0; pt < NumProductTerms; pt++ {
			value := (config[i] & (1 << pt)) != 0
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
		var pt uint
		for pt = 0; pt < NumProductTerms; pt++ {
			value := (config[i] & (1 << pt)) != 0
			u.PLDs[pld].ProductTerms[pt].Enabled[ot] = value
		}
	}

	for i := 0x38; i < 0x40; i++ {
		pld := i & 1
		mct := (i >> 1) & 3
		var bit uint
		for bit = 0; bit < 8; bit++ {
			mc := (bit >> 1) & 3
			switch mct {
			case 0: // CEN_CONST
				cc := (config[i] & (1 << bit)) != 0
				if (bit & 1) == 0 {
					u.PLDs[pld].Macrocells[mc].COEN = cc
				} else {
					u.PLDs[pld].Macrocells[mc].CONST = cc
				}
			case 1: // XORFB
				if (bit & 1) == 0 {
					u.PLDs[pld].Macrocells[mc].XORFB = xorfb((config[i] >> bit) & 3)
				}
			case 2: // SET_RESET
				sr := (config[i] & (1 << bit)) != 0
				if (bit & 1) == 0 {
					u.PLDs[pld].Macrocells[mc].SSEL = sr
				} else {
					u.PLDs[pld].Macrocells[mc].RSEL = sr
				}
			case 3: // BYPASS
				if (bit & 1) == 0 {
					bypass := (config[i] & (1 << bit)) != 0
					u.PLDs[pld].Macrocells[mc].BYP = bypass
				}
			}
		}
	}
	return nil
}

func (u *UDB) GetVerilog() string {
	var buffer bytes.Buffer

	buffer.WriteString("module UDB();\n")
	buffer.WriteString("\n")

	buffer.WriteString("input wire clk;\n")
	buffer.WriteString("input wire reset;\n")
	buffer.WriteString("\n")

	buffer.WriteString(fmt.Sprintf("input wire pld_en[%d:0];\n", NumPLDs-1))
	buffer.WriteString("input wire selin;\n")
	buffer.WriteString("\n")

	buffer.WriteString(fmt.Sprintf("input wire it[%d:0];\n", NumInputTerms-1))
	buffer.WriteString("\n")

	buffer.WriteString(fmt.Sprintf("output wire out[%d:0];\n", NumMacrocells*NumPLDs-1))
	buffer.WriteString("output wire selout;\n")
	buffer.WriteString("\n")

	for pld := 0; pld < NumPLDs; pld++ {
		p := u.PLDs[pld]
		for pt := 0; pt < NumProductTerms; pt++ {
			buffer.WriteString(fmt.Sprintf("wire pld%d_pt%d = ", pld, pt))
			for it := 0; it < NumInputTerms; it++ {
				if (!p.InputTerms[it].True[pt]) && (!p.InputTerms[it].Complement[pt]) ||
					(p.InputTerms[it].True[pt]) && (p.InputTerms[it].Complement[pt]) {
					buffer.WriteString("1'b1")
				} else if (p.InputTerms[it].True[pt]) && (!p.InputTerms[it].Complement[pt]) {
					buffer.WriteString(fmt.Sprintf("it[%d]", it))
				} else if (!p.InputTerms[it].True[pt]) && (p.InputTerms[it].Complement[pt]) {
					buffer.WriteString(fmt.Sprintf("!it[%d]", it))
				}
				if (it + 1) < NumInputTerms {
					buffer.WriteString(" & ")
				}
			}
			buffer.WriteString(";\n")
		}

		for ot := 0; ot < NumOrTerms; ot++ {
			buffer.WriteString(fmt.Sprintf("wire pld%d_or%d = ", pld, ot))
			for pt := 0; pt < NumProductTerms; pt++ {
				if p.ProductTerms[pt].Enabled[ot] {
					buffer.WriteString(fmt.Sprintf("pld%d_pt%d", pld, pt))
				} else {
					buffer.WriteString("1'b0")
				}
				if (pt + 1) < NumProductTerms {
					buffer.WriteString(" | ")
				}
			}
			buffer.WriteString(";\n")
		}

		for mc := 0; mc < NumMacrocells; mc++ {
			for cpt := 0; cpt < 2; cpt++ {
				buffer.WriteString(fmt.Sprintf("wire pld%d_mc%d_cpt%d = pld%d_pt%d;\n", pld, mc, cpt, pld, mc*2+cpt))
			}
		}
		buffer.WriteString("\n")
	}

	buffer.WriteString("wire pld0_mc0_selin = selin;\n")
	for pld := 0; pld < NumPLDs; pld++ {
		for mc := 0; mc < NumMacrocells; mc++ {
			if u.PLDs[pld].Macrocells[mc].COEN {
				buffer.WriteString(fmt.Sprintf("wire pld%d_mc%d_selout = (pld%d_mc%d_cpt0 & !pld%d_mc%d_selin) | (!pld%d_mc%d_cpt1 & pld%d_mc%d_selin);\n", pld, mc, pld, mc, pld, mc, pld, mc, pld, mc))
			} else {
				buffer.WriteString(fmt.Sprintf("wire pld%d_mc%d_selout = 1'b0;\n", pld, mc))
			}

			if mc < NumMacrocells-1 {
				buffer.WriteString(fmt.Sprintf("wire pld%d_mc%d_selin = pld%d_mc%d_selout;\n", pld, mc+1, pld, mc))
			} else if pld < NumPLDs-1 {
				buffer.WriteString(fmt.Sprintf("wire pld%d_mc0_selin = pld%d_mc%d_selout;\n", pld+1, pld, mc))
			}
		}
	}
	buffer.WriteString(fmt.Sprintf("assign selout = pld%d_mc%d_selout;\n", NumPLDs-1, NumMacrocells-1))
	buffer.WriteString("\n")

	for pld := 0; pld < NumPLDs; pld++ {
		for mc := 0; mc < NumMacrocells; mc++ {
			buffer.WriteString(fmt.Sprintf("reg out%d_reg;\n", pld*4+mc))
			buffer.WriteString(fmt.Sprintf("wire out%d_int = pld%d_or%d ^ ", pld*4+mc, pld, mc))
			switch u.PLDs[pld].Macrocells[mc].XORFB {
			case dff:
				buffer.WriteString("1'b0")
			case carry:
				buffer.WriteString(fmt.Sprintf("pld%d_mc%d_selin", pld, mc))
			case tffHigh:
				buffer.WriteString(fmt.Sprintf("out%d_reg", pld*4+mc))
			case tffLow:
				buffer.WriteString(fmt.Sprintf("!out%d_reg", pld*4+mc))
			}
			buffer.WriteString(";\n")
		}
	}
	buffer.WriteString("\n")

	for pld := 0; pld < NumPLDs; pld++ {
		for mc := 0; mc < NumMacrocells; mc++ {
			if u.PLDs[pld].Macrocells[mc].BYP {
				buffer.WriteString(fmt.Sprintf("assign out[%d] = (out%d_reg & !pld_en[%d]) | (out%d_int & pld_en[%d]);\n", pld*4+mc, pld*4+mc, pld, pld*4+mc, pld))
			} else {
				buffer.WriteString(fmt.Sprintf("assign out[%d] = out%d_reg;\n", pld*4+mc, pld*4+mc))
			}
		}
	}

	buffer.WriteString("\n")
	for pld := 0; pld < NumPLDs; pld++ {
		for mc := 0; mc < NumMacrocells; mc++ {
			cell := u.PLDs[pld].Macrocells[mc]
			buffer.WriteString("always @(posedge clk")
			if cell.SSEL || cell.RSEL {
				buffer.WriteString(" or posedge reset")
			}
			buffer.WriteString(") begin\n")

			if cell.SSEL || cell.RSEL {
				buffer.WriteString("if reset begin\n")
			}
			if cell.SSEL {
				buffer.WriteString(fmt.Sprintf("out%d_reg <= 1'b1;\n", pld*4+mc))
			}
			if cell.RSEL {
				buffer.WriteString(fmt.Sprintf("out%d_reg <= 1'b0;\n", pld*4+mc))
			}
			if cell.SSEL || cell.RSEL {
				buffer.WriteString("end else begin\n")
			}
			buffer.WriteString(fmt.Sprintf("out%d_reg <= out%d_int;\n", pld*4+mc, pld*4+mc))
			if cell.SSEL || cell.RSEL {
				buffer.WriteString("end\n")
			}

			buffer.WriteString("end\n")
		}
	}
	buffer.WriteString("\n")

	buffer.WriteString("endmodule\n")

	return buffer.String()
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
	err = u.LoadConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// dump, err := json.MarshalIndent(u, "", "  ")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%s\n", dump)

	verilog := u.GetVerilog()
	fmt.Printf("%s", verilog)
}
