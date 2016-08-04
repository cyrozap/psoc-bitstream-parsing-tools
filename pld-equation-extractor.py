#!/usr/bin/env python3
#
# pld-equation-extractor
#
# Copyright (c) 2016, Forest Crossman <cyrozap@gmail.com>
#
# Permission to use, copy, modify, and/or distribute this software for any
# purpose with or without fee is hereby granted, provided that the above
# copyright notice and this permission notice appear in all copies.
#
# THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
# WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
# MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
# ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
# WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
# ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
# OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
#

import argparse

MC_CFG_TYPES = {
    0: "CEN_CONST",
    1: "XORFB",
    2: "SET_RESET",
    3: "BYPASS"
}

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("udb_file", type=str, help="The name of the UDB configuration memory binary you want to parse.")
    args = vars(parser.parse_args())

    udb_config = open(args["udb_file"], 'rb').read()
    for byte in range(0, 0x30):
        pld_number = byte & 1
        true_comp = (byte & 2) == 2
        it_number = byte >> 2
        if udb_config[byte] != 0:
            for bit in range(0, 8):
                pt_number = bit
                if udb_config[byte] & (1 << bit):
                    print("PLD{}, {}, IT{}, PT{}".format(pld_number, "True" if true_comp else "Complement", it_number, pt_number))

    for byte in range(0x30, 0x38):
        pld_number = byte & 1
        ot_number = (byte >> 1) & 3
        if udb_config[byte] != 0:
            for bit in range(0, 8):
                pt_number = bit
                if udb_config[byte] & (1 << bit):
                    print("PLD{}, OT{}, PT{}".format(pld_number, ot_number, pt_number))

    for byte in range(0x38, 0x40):
        pld_number = byte & 1
        mc_cfg_type = MC_CFG_TYPES[(byte >> 1) & 3]
        if udb_config[byte] != 0:
            print("PLD{}, MC CFG {}: {:#02x}".format(pld_number, mc_cfg_type, udb_config[byte]))
            for bit in range(0,8):
                mc_number = (bit >> 1) & 3
                if mc_cfg_type == "SET_RESET":
                    if udb_config[byte] & (1 << bit):
                        if not (bit & 1):
                            print("PLD{}, MC{} SET_SEL".format(pld_number, mc_number))
                        else:
                            print("PLD{}, MC{} RESET_SEL".format(pld_number, mc_number))
                if mc_cfg_type == "BYPASS":
                    if udb_config[byte] & (1 << bit):
                        print("PLD{}, MC{} BYPASS".format(pld_number, mc_number))
