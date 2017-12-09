/*
 * macrocell.v
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

module MC (
	// Config
	input wire coen,
	input wire const,
	input wire [1:0] xorfb,
	input wire ssel,
	input wire rsel,
	input wire byp,

	// Normal I/O
	input wire clk,
	input wire reset,

	input wire in,
	input wire selin,
	input wire cpt0,
	input wire cpt1,
	input wire pld_en,

	output wire out,
	output wire selout
);

reg out_reg;
reg xorfb_mux;

wire out_int = in ^ xorfb_mux;
wire set_int = ssel ? reset : 1'b0;
wire res_int = rsel ? reset : 1'b0;

assign out = (byp & pld_en) ? out_int : out_reg;
assign selout = coen & (selin ? !cpt1 : cpt0);

// xorfb mux
always @* begin
	case (xorfb)
		2'h0 : xorfb_mux = const;
		2'h1 : xorfb_mux = !selin;
		2'h2 : xorfb_mux = out_reg;
		2'h3 : xorfb_mux = !out_reg;
	endcase
end

// DFF
always @(posedge clk, posedge set_int, posedge res_int) begin
	if (res_int)
		out_reg <= 1'b0;
	else if (set_int)
		out_reg <= 1'b1;
	else
		out_reg <= out_int;
end

endmodule
