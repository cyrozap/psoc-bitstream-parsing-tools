# psoc-bitstream-parsing-tools

## Usage

### Examples

Dump a UDB configuration to Verilog and use Yosys to simplify it.

```bash
./udb-config-parser/udb-config-parser ./examples/udb-and2.bin > udb-and2.v
yosys -p 'read_verilog udb-and2.v; synth; write_verilog udb-and2-simplified.v'
```
