/*
 * pld.v
 *
 * Copyright (c) 2017, Forest Crossman <cyrozap@gmail.com>
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

`include "macrocell.v"

`define BYTE2BIT(BYTE) (BYTE)*8
`define PLD_CONFIG_LEN `BYTE2BIT(8'h40)

module PLD #(
	parameter PLD = 1'b0
) (
	// Config
	input wire [`PLD_CONFIG_LEN-1:0] config,

	// Normal I/O
	input wire clk,
	input wire reset,

	input wire pld_en,
	input wire selin,

	input wire [11:0] in,

	output wire [3:0] out,
	output wire selout
);

wire [7:0] pt_int;
wire [3:0] ot_int;
wire [4:0] sel_int;
assign sel_int[0] = selin;
assign selout = sel_int[4];

`define IT_COMP 0
`define IT_TRUE 1
`define IT_CONFIG(IT, TC, PT) config[`BYTE2BIT(IT*4+PLD+TC*2)+PT]
`define IT_VAL(IT, PT) (in[IT] ? `IT_CONFIG(IT, `IT_TRUE, PT) : `IT_CONFIG(IT, `IT_COMP, PT))
genvar pt;
generate
for (pt = 0; pt < 8; pt = pt + 1) begin
	assign pt_int[pt] = `IT_VAL(0, pt) & `IT_VAL(1, pt) & `IT_VAL(2, pt) & `IT_VAL(3, pt) & `IT_VAL(4, pt) & `IT_VAL(5, pt) & `IT_VAL(6, pt) & `IT_VAL(7, pt) & `IT_VAL(8, pt) & `IT_VAL(9, pt) & `IT_VAL(10, pt) & `IT_VAL(11, pt);
end
endgenerate

`define OT_CONFIG(OT) config[`BYTE2BIT(8'h30+PLD+OT*2) +: 8]
genvar ot;
generate
for (ot = 0; ot < 4; ot = ot + 1) begin
	assign ot_int[ot] = |(pt_int & `OT_CONFIG(ot));
end
endgenerate

`define COENBIT(MC) config[`BYTE2BIT(8'h38+PLD)+MC*2]
`define CONSTBIT(MC) config[`BYTE2BIT(8'h38+PLD)+MC*2+1]
`define XORFBBITS(MC) config[`BYTE2BIT(8'h3A+PLD)+MC*2 +: 2]
`define SSELBIT(MC) config[`BYTE2BIT(8'h3C+PLD)+MC*2]
`define RSELBIT(MC) config[`BYTE2BIT(8'h3C+PLD)+MC*2+1]
`define BYPBIT(MC) config[`BYTE2BIT(8'h3E+PLD)+MC*2]
genvar mc;
generate
for (mc = 0; mc < 4; mc = mc + 1) begin : mc_gen
	MC MC (
		// Config
		.coen(`COENBIT(mc)),
		.const(`CONSTBIT(mc)),
		.xorfb(`XORFBBITS(mc)),
		.ssel(`SSELBIT(mc)),
		.rsel(`RSELBIT(mc)),
		.byp(`BYPBIT(mc)),

		// Normal I/O
		.clk(clk),
		.reset(reset),

		.in(ot_int[mc]),
		.selin(sel_int[mc]),
		.cpt0(pt_int[mc*2]),
		.cpt1(pt_int[mc*2+1]),
		.pld_en(pld_en),

		.out(out[mc]),
		.selout(sel_int[mc+1])
	);
end
endgenerate

endmodule
