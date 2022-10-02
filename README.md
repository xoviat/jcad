# JCAD - Generate JLCPCBA files directly from KiCad PCB files

This project aims to simplify and clarify the process of ordering SMT assembly boards from JLCPCB.
The problem is as such: given a KiCAD a .kicad_pcb file, what is the simpliest and fastest way
to order SMT assembled boards from JLCPCB? The current process is as follows:

1. Generate a BOM in EEschema
2. Generate a POS file in PCBnew
3. Generate Gerbers in PCBnew
4. Run [tooling](https://github.com/wokwi/kicad-jlcpcb-bom-plugin) to transform the BOM and pos file
5. Remove thorugh-hole components from the BOM
6. Upload the files to JLCPCB and order the SMT board
7. Order additional through-hole components from a distributor, and solder these components yourself

## How JCAD works

JCAD associates each combination of designator, footprint, and value on a KiCad PCB with a specific
JLCPCB part. The association is global, meaning that once set-up, fabrication files for JLCPCB can
be regenerated in seconds, significantly improving iteration speed. JCAD uses KiCad's Python API
to extract data from the KiCad files, then matches this data with its internal database to prepare
the JLCPCB files.

In the event that a component has a rotational mismatch or needs to be changed to a different
component, JCAD includes options to do just that. However, this workflow may be subject to change in
future updates.

Note: Component rotations should always be specified within KiCad footprints. Because component rotation
is set during DFM review, the JCAD rotation database may be removed in a future release.

# Terminology

*Component Database* - The database downloaded from jlcpcb.com/parts
*Association Database* - The internal database created to associate a KiCad part with a JLCPCB part.

